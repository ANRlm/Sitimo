package export

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/store"
	"mathlib/server/internal/worker"

	"github.com/rs/zerolog"
)

var errExportCancelled = errors.New("export cancelled")

var bareTextCommandPattern = regexp.MustCompile(`\\text\{([^{}]+)\}`)

type latexImageAsset struct {
	SourcePath string
	TexPath    string
}

type Manager struct {
	repo        *store.Repository
	broadcaster *worker.Broadcaster
	logger      zerolog.Logger
	queue       chan string
	once        sync.Once
}

func NewManager(repo *store.Repository, broadcaster *worker.Broadcaster, logger zerolog.Logger) *Manager {
	return &Manager{
		repo:        repo,
		broadcaster: broadcaster,
		logger:      logger,
		queue:       make(chan string, 64),
	}
}

func (m *Manager) Start(ctx context.Context) {
	m.once.Do(func() {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case jobID := <-m.queue:
					m.process(ctx, jobID)
				}
			}
		}()
	})
}

func (m *Manager) Enqueue(jobID string) {
	select {
	case m.queue <- jobID:
	default:
		go func() { m.queue <- jobID }()
	}
}

func (m *Manager) process(ctx context.Context, jobID string) {
	if err := m.repo.UpdateExportJobState(ctx, jobID, domain.ExportStatusProcessing, 10, nil, nil); err != nil {
		m.logger.Error().Err(err).Str("job_id", jobID).Msg("failed to mark export processing")
		return
	}
	m.publishJob(ctx, jobID)
	if m.handleCancellation(ctx, jobID) {
		return
	}

	job, err := m.repo.GetExportJob(ctx, jobID)
	if err != nil {
		m.fail(ctx, jobID, fmt.Sprintf("导出任务不存在：%v", err))
		return
	}

	paper, err := m.repo.GetPaperDetail(ctx, job.PaperID, true)
	if err != nil {
		m.fail(ctx, jobID, fmt.Sprintf("试卷不存在：%v", err))
		return
	}

	rendered, latexAssets, err := m.renderLatex(*paper, job.Variant)
	if err != nil {
		m.fail(ctx, jobID, fmt.Sprintf("生成 LaTeX 失败：%v", err))
		return
	}

	if err := m.repo.UpdateExportJobState(ctx, jobID, domain.ExportStatusProcessing, 40, nil, nil); err != nil {
		m.logger.Error().Err(err).Str("job_id", jobID).Msg("failed to update export progress")
	}
	m.publishJob(ctx, jobID)
	if m.handleCancellation(ctx, jobID) {
		return
	}

	artifactRel := filepath.Join("derived", "exports", m.outputFilename(job.ID, job.Format))
	artifactAbs := filepath.Join(m.repo.StorageRoot(), artifactRel)
	if err := os.MkdirAll(filepath.Dir(artifactAbs), 0o755); err != nil {
		m.fail(ctx, jobID, fmt.Sprintf("创建导出目录失败：%v", err))
		return
	}

	switch job.Format {
	case domain.ExportFormatLatex:
		if err := m.writeLatexBundle(artifactAbs, rendered, latexAssets); err != nil {
			m.fail(ctx, jobID, fmt.Sprintf("写入 LaTeX 压缩包失败：%v", err))
			return
		}
	case domain.ExportFormatPDF:
		if err := m.writePDF(ctx, jobID, artifactAbs, rendered, latexAssets); err != nil {
			if errors.Is(err, errExportCancelled) {
				m.fail(ctx, jobID, "用户取消")
				return
			}
			m.fail(ctx, jobID, err.Error())
			return
		}
	default:
		m.fail(ctx, jobID, "不支持的导出格式")
		return
	}
	if m.handleCancellation(ctx, jobID) {
		return
	}

	if err := m.repo.UpdateExportJobState(ctx, jobID, domain.ExportStatusDone, 100, &artifactRel, nil); err != nil {
		m.logger.Error().Err(err).Str("job_id", jobID).Msg("failed to mark export done")
	}
	m.publishJob(ctx, jobID)
}

func (m *Manager) fail(ctx context.Context, jobID string, message string) {
	_ = m.repo.UpdateExportJobState(ctx, jobID, domain.ExportStatusFailed, 100, nil, &message)
	m.publishJob(ctx, jobID)
}

func (m *Manager) publishJob(ctx context.Context, jobID string) {
	if err := m.repo.NotifyExportJobChanged(ctx, jobID); err != nil {
		m.logger.Warn().Err(err).Str("job_id", jobID).Msg("failed to notify export update")
	}
}

func (m *Manager) outputFilename(jobID string, format domain.ExportFormat) string {
	if format == domain.ExportFormatLatex {
		return jobID + ".zip"
	}
	return jobID + ".pdf"
}

func (m *Manager) writePDF(ctx context.Context, jobID, target string, latexSource string, latexAssets []latexImageAsset) error {
	tempDir, err := os.MkdirTemp("", "mathlib-export-*")
	if err != nil {
		return fmt.Errorf("创建临时目录失败：%w", err)
	}
	defer os.RemoveAll(tempDir)

	if err := m.stageLatexWorkspace(tempDir, latexSource, latexAssets, false); err != nil {
		return err
	}

	cmdCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go m.watchCancellation(cmdCtx, jobID, cancel)

	cmd := exec.CommandContext(cmdCtx, "xelatex", "-interaction=nonstopmode", "-output-directory", tempDir, filepath.Join(tempDir, "main.tex"))
	var stderr bytes.Buffer
	cmd.Stdout = &stderr
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if cmdCtx.Err() != nil {
			return errExportCancelled
		}
		return fmt.Errorf("XeLaTeX 编译失败：%s", trimError(stderr.String()))
	}
	pdfBytes, err := os.ReadFile(filepath.Join(tempDir, "main.pdf"))
	if err != nil {
		return fmt.Errorf("读取生成 PDF 失败：%w", err)
	}
	if err := os.WriteFile(target, pdfBytes, 0o644); err != nil {
		return fmt.Errorf("写入 PDF 失败：%w", err)
	}
	return nil
}

func (m *Manager) writeLatexBundle(target string, latexSource string, latexAssets []latexImageAsset) error {
	tempDir, err := os.MkdirTemp("", "mathlib-latex-bundle-*")
	if err != nil {
		return fmt.Errorf("创建临时 LaTeX 目录失败：%w", err)
	}
	defer os.RemoveAll(tempDir)

	if err := m.stageLatexWorkspace(tempDir, latexSource, latexAssets, true); err != nil {
		return err
	}
	if err := zipDirectory(tempDir, target); err != nil {
		return fmt.Errorf("打包 LaTeX 压缩包失败：%w", err)
	}
	return nil
}

func (m *Manager) stageLatexWorkspace(root string, latexSource string, latexAssets []latexImageAsset, includeReadme bool) error {
	mainTex := filepath.Join(root, "main.tex")
	if err := os.WriteFile(mainTex, []byte(latexSource), 0o644); err != nil {
		return fmt.Errorf("写入临时 LaTeX 文件失败：%w", err)
	}
	if includeReadme {
		if err := os.WriteFile(filepath.Join(root, "README.txt"), []byte(latexBundleReadme()), 0o644); err != nil {
			return fmt.Errorf("写入 LaTeX 说明文件失败：%w", err)
		}
		if err := os.WriteFile(filepath.Join(root, "latexmkrc"), []byte(latexmkrcContent()), 0o644); err != nil {
			return fmt.Errorf("写入 latexmkrc 失败：%w", err)
		}
	}

	for _, asset := range latexAssets {
		if err := copyLatexAsset(root, asset); err != nil {
			m.logger.Warn().Err(err).Str("source", asset.SourcePath).Str("target", asset.TexPath).Msg("failed to stage latex image asset")
		}
	}

	return nil
}

func (m *Manager) watchCancellation(ctx context.Context, jobID string, cancel context.CancelFunc) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cancelled, err := m.repo.IsExportCancellationRequested(ctx, jobID)
			if err != nil {
				continue
			}
			if cancelled {
				cancel()
				return
			}
		}
	}
}

func (m *Manager) handleCancellation(ctx context.Context, jobID string) bool {
	cancelled, err := m.repo.IsExportCancellationRequested(ctx, jobID)
	if err != nil || !cancelled {
		return false
	}
	m.fail(ctx, jobID, "用户取消")
	return true
}

func (m *Manager) renderLatex(paper domain.PaperDetail, variant domain.ExportVariant) (string, []latexImageAsset, error) {
	type renderImage struct {
		TexPath string
	}

	type renderItem struct {
		Index         int
		Score         float64
		ProblemLatex  string
		AnswerLatex   string
		SolutionLatex string
		Images        []renderImage
		BlankLines    int
	}

	items := make([]renderItem, 0, len(paper.ItemDetails))
	latexAssets := make([]latexImageAsset, 0)
	seenAssets := make(map[string]struct{})
	for index, item := range paper.ItemDetails {
		if item.Problem == nil {
			continue
		}
		rendered := renderItem{
			Index:         index + 1,
			Score:         item.Score,
			ProblemLatex:  normalizeLatexForExport(item.Problem.Latex),
			AnswerLatex:   normalizeLatexForExport(deref(item.Problem.AnswerLatex)),
			SolutionLatex: normalizeLatexForExport(deref(item.Problem.SolutionLatex)),
			Images:        make([]renderImage, 0, len(item.Problem.Images)),
			BlankLines:    item.BlankLines,
		}
		for _, imageAsset := range item.Problem.Images {
			info, err := m.repo.GetImageStorageInfo(context.Background(), imageAsset.ID)
			if err == nil {
				asset := latexImageReference(m.repo.StorageRoot(), info)
				rendered.Images = append(rendered.Images, renderImage{TexPath: asset.TexPath})
				if _, exists := seenAssets[asset.TexPath]; !exists {
					latexAssets = append(latexAssets, asset)
					seenAssets[asset.TexPath] = struct{}{}
				}
			}
		}
		items = append(items, rendered)
	}

	data := map[string]any{
		"Title":             paper.Title,
		"SchoolName":        deref(paper.SchoolName),
		"ExamName":          deref(paper.ExamName),
		"Columns":           paper.Layout.Columns,
		"FontSize":          paper.Layout.FontSize,
		"PaperSize":         strings.ToLower(paper.Layout.PaperSize),
		"LineHeight":        paper.Layout.LineHeight,
		"Variant":           variant,
		"Items":             items,
		"GeneratedAt":       time.Now().Format("2006-01-02 15:04"),
		"ShowAnswerVersion": variant == domain.ExportVariantAnswer || variant == domain.ExportVariantBoth,
	}

	const tpl = `% !TeX program = xelatex
\RequirePackage{iftex}
\documentclass[{{.FontSize}}pt{{if eq .Columns 2}},twocolumn{{end}}]{article}
\ifPDFTeX
\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}
\usepackage{CJKutf8}
\newcommand{\MathLibBeginDocument}{\begin{CJK*}{UTF8}{gbsn}}
\newcommand{\MathLibEndDocument}{\end{CJK*}}
\else
\usepackage[UTF8]{ctex}
\newcommand{\MathLibBeginDocument}{}
\newcommand{\MathLibEndDocument}{}
\fi
\usepackage{amsmath,amssymb,graphicx,geometry,enumitem}
\geometry{ {{.PaperSize}}paper, margin=2cm }
\linespread{ {{.LineHeight}} }
\begin{document}
\MathLibBeginDocument
\begin{center}
\Large\textbf{ {{.Title}} }\par
\vspace{0.5em}
\textbf{ {{.SchoolName}} {{if .ExamName}}· {{.ExamName}}{{end}} }\\
生成时间：{{.GeneratedAt}}
\end{center}

% Upload matching image files into ./images when compiling in Overleaf.

{{range .Items}}
\par\noindent\textbf{ {{.Index}}. }（{{printf "%.0f" .Score}}分）{{.ProblemLatex}}
{{range .Images}}
\begin{center}
\IfFileExists{images/{{.TexPath}}}{\includegraphics[width=0.65\textwidth]{images/{{.TexPath}}}}{\fbox{\parbox{0.7\textwidth}{\centering 缺少图像文件\\\texttt{\detokenize{images/{{.TexPath}}}}}}}
\end{center}
{{end}}
{{if gt .BlankLines 0}}
\par\vspace*{ {{.BlankLines}}\baselineskip }
{{end}}
\vspace{1.5em}
{{end}}
\MathLibEndDocument
\end{document}
`

	parsed, err := template.New("paper").Parse(tpl)
	if err != nil {
		return "", nil, err
	}
	var buffer bytes.Buffer
	if err := parsed.Execute(&buffer, data); err != nil {
		return "", nil, err
	}
	return buffer.String(), latexAssets, nil
}

func trimError(raw string) string {
	raw = strings.TrimSpace(raw)
	if len(raw) > 1200 {
		return raw[len(raw)-1200:]
	}
	return raw
}

func deref(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func normalizeLatexForExport(value string) string {
	if value == "" {
		return value
	}

	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	value = strings.ReplaceAll(value, `\r\n`, "\n\n")
	value = rewriteLatexSegments(value, normalizeLatexTextSegment, normalizeLatexMathSegment)

	var builder strings.Builder
	builder.Grow(len(value))

	for i := 0; i < len(value); i++ {
		if value[i] == '\\' && i+1 < len(value) && value[i+1] == 'n' {
			if shouldConvertLiteralEscapedNewline(value, i+2) {
				builder.WriteString("\n\n")
				i++
				continue
			}
		}
		builder.WriteByte(value[i])
	}

	return collapseBlankLines(builder.String())
}

func rewriteLatexSegments(value string, textTransform func(string) string, mathTransform func(string) string) string {
	var builder strings.Builder
	cursor := 0

	for cursor < len(value) {
		startIndex, open, close := findNextMathStart(value, cursor)
		if startIndex < 0 {
			builder.WriteString(textTransform(value[cursor:]))
			break
		}

		builder.WriteString(textTransform(value[cursor:startIndex]))
		builder.WriteString(open)

		mathStart := startIndex + len(open)
		mathLength := strings.Index(value[mathStart:], close)
		if mathLength < 0 {
			builder.WriteString(mathTransform(value[mathStart:]))
			break
		}

		mathEnd := mathStart + mathLength
		builder.WriteString(mathTransform(value[mathStart:mathEnd]))
		builder.WriteString(close)
		cursor = mathEnd + len(close)
	}

	return builder.String()
}

func findNextMathStart(value string, from int) (index int, open string, close string) {
	inlineIndex := strings.Index(value[from:], `\(`)
	displayIndex := strings.Index(value[from:], `\[`)

	switch {
	case inlineIndex < 0 && displayIndex < 0:
		return -1, "", ""
	case inlineIndex >= 0 && (displayIndex < 0 || inlineIndex <= displayIndex):
		return from + inlineIndex, `\(`, `\)`
	default:
		return from + displayIndex, `\[`, `\]`
	}
}

func normalizeLatexTextSegment(segment string) string {
	return bareTextCommandPattern.ReplaceAllString(segment, `$1`)
}

func normalizeLatexMathSegment(segment string) string {
	return strings.ReplaceAll(segment, "°", `^\circ`)
}

func shouldConvertLiteralEscapedNewline(value string, next int) bool {
	if next >= len(value) {
		return true
	}

	char := value[next]
	return !(char >= 'A' && char <= 'Z' || char >= 'a' && char <= 'z' || char == '@')
}

func collapseBlankLines(value string) string {
	var builder strings.Builder
	builder.Grow(len(value))

	newlineCount := 0
	for i := 0; i < len(value); i++ {
		if value[i] == '\n' {
			newlineCount++
			if newlineCount <= 2 {
				builder.WriteByte('\n')
			}
			continue
		}

		newlineCount = 0
		builder.WriteByte(value[i])
	}

	return builder.String()
}

func latexImageReference(storageRoot string, info *store.StoredImageInfo) latexImageAsset {
	return latexImageAsset{
		SourcePath: filepath.Join(storageRoot, info.StoragePath),
		TexPath:    latexImageFilename(info),
	}
}

func latexImageFilename(info *store.StoredImageInfo) string {
	ext := filepath.Ext(info.StoragePath)
	if ext == "" {
		ext = filepath.Ext(info.Filename)
	}
	if ext == "" {
		ext = ".png"
	}
	return info.ID + ext
}

func copyLatexAsset(tempDir string, asset latexImageAsset) error {
	targetPath := filepath.Join(tempDir, "images", filepath.FromSlash(asset.TexPath))
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}

	sourceFile, err := os.Open(asset.SourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	targetFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	if _, err := io.Copy(targetFile, sourceFile); err != nil {
		return err
	}
	return nil
}

func zipDirectory(sourceDir string, targetPath string) error {
	targetFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	zipWriter := zip.NewWriter(targetFile)
	defer zipWriter.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		sourceFile, err := os.Open(path)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, sourceFile)
		closeErr := sourceFile.Close()
		if err != nil {
			return err
		}
		return closeErr
	})
}

func latexBundleReadme() string {
	return strings.TrimSpace(`MathLib LaTeX bundle for Overleaf

1. Upload the whole ZIP to Overleaf and let it extract the project files.
2. Keep main.tex at the project root.
3. The bundle includes a latexmkrc file that prefers XeLaTeX automatically in Overleaf.
4. main.tex also has a pdfLaTeX-compatible fallback, so Overleaf defaults should still compile.
5. If you replace or add figures manually, put them under the images/ folder.
`)
}

func latexmkrcContent() string {
	return strings.TrimSpace(`
$pdf_mode = 5;
$xelatex = 'xelatex -interaction=nonstopmode -file-line-error -synctex=1 %O %S';
`) + "\n"
}
