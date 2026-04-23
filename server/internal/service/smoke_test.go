package service_test

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"mathlib/server/internal/config"
	"mathlib/server/internal/domain"
	exportsvc "mathlib/server/internal/export"
	"mathlib/server/internal/service"
	"mathlib/server/internal/store"
	"mathlib/server/internal/worker"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

func TestSmokeFlows(t *testing.T) {
	ctx := context.Background()

	cfg, cleanup := loadIsolatedTestConfig(t)
	cfg.PublicBaseURL = "http://localhost:8080"
	cfg.StorageRoot = t.TempDir()
	t.Cleanup(cleanup)

	repo, err := store.NewRepository(ctx, cfg)
	if err != nil {
		t.Skipf("skip smoke test, database unavailable: %v", err)
	}
	t.Cleanup(repo.Close)
	if err := repo.Ping(ctx); err != nil {
		t.Skipf("skip smoke test, database unavailable: %v", err)
	}

	applyMigrations(t, repo.DB())
	resetDatabase(t, repo.DB())

	broadcaster := worker.NewBroadcaster()
	runCtx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	exporter := exportsvc.NewManager(repo, broadcaster, zerolog.Nop())
	exporter.Start(runCtx)

	svc := service.New(cfg, repo, zerolog.Nop(), broadcaster, exporter)

	t.Run("problem_crud_and_versions", func(t *testing.T) {
		resetDatabase(t, repo.DB())
		tag := mustCreateTag(t, repo, "函数")

		created, _, err := svc.CreateProblem(ctx, domain.ProblemWriteInput{
			Latex:      "Evaluate integral and function behavior.",
			Type:       domain.ProblemTypeSolve,
			Difficulty: domain.DifficultyMedium,
			Subject:    stringPtr("数学"),
			Grade:      stringPtr("高三"),
			TagIDs:     []string{tag.ID},
		})
		if err != nil {
			t.Fatalf("create problem: %v", err)
		}
		if created.Version != 1 {
			t.Fatalf("expected version 1, got %d", created.Version)
		}
		if created.Code == "" {
			t.Fatal("expected generated problem code")
		}

		updated, _, err := svc.UpdateProblem(ctx, created.ID, domain.ProblemWriteInput{
			Latex:      "Updated integral prompt with function details.",
			Type:       domain.ProblemTypeSolve,
			Difficulty: domain.DifficultyHard,
			Subject:    stringPtr("数学"),
			Grade:      stringPtr("高三"),
			TagIDs:     []string{tag.ID},
		})
		if err != nil {
			t.Fatalf("update problem: %v", err)
		}
		if updated.Version != 2 {
			t.Fatalf("expected version 2, got %d", updated.Version)
		}

		versions, err := repo.ListProblemVersions(ctx, created.ID)
		if err != nil {
			t.Fatalf("list versions: %v", err)
		}
		if len(versions) != 2 {
			t.Fatalf("expected 2 versions, got %d", len(versions))
		}
	})

	t.Run("image_replacement", func(t *testing.T) {
		resetDatabase(t, repo.DB())

		oldImage := mustCreateImage(t, repo, "old.png", "original/old.png", "thumbnails/old.png")
		newImage := mustCreateImage(t, repo, "new.png", "original/new.png", "thumbnails/new.png")

		created, _, err := svc.CreateProblem(ctx, domain.ProblemWriteInput{
			Latex:      "Problem with one image.",
			Type:       domain.ProblemTypeSolve,
			Difficulty: domain.DifficultyEasy,
			ImageIDs:   []string{oldImage.ID},
		})
		if err != nil {
			t.Fatalf("create problem with image: %v", err)
		}

		if err := repo.ReplaceProblemImage(ctx, created.ID, oldImage.ID, newImage.ID); err != nil {
			t.Fatalf("replace problem image: %v", err)
		}

		loaded, err := repo.GetProblemDetail(ctx, created.ID, true)
		if err != nil {
			t.Fatalf("load updated problem: %v", err)
		}
		if len(loaded.ImageIDs) != 1 || loaded.ImageIDs[0] != newImage.ID {
			t.Fatalf("expected new image to replace old one, got %v", loaded.ImageIDs)
		}
	})

	t.Run("search_hit_with_snippet", func(t *testing.T) {
		resetDatabase(t, repo.DB())

		created, _, err := svc.CreateProblem(ctx, domain.ProblemWriteInput{
			Latex:      "Evaluate integral of x squared on [0,1].",
			Type:       domain.ProblemTypeSolve,
			Difficulty: domain.DifficultyMedium,
		})
		if err != nil {
			t.Fatalf("create searchable problem: %v", err)
		}
		_, _, err = svc.CreateProblem(ctx, domain.ProblemWriteInput{
			Latex:      "Probability counting exercise.",
			Type:       domain.ProblemTypeSolve,
			Difficulty: domain.DifficultyEasy,
		})
		if err != nil {
			t.Fatalf("create unrelated problem: %v", err)
		}

		results, err := svc.Search(ctx, "integral", "", nil)
		if err != nil {
			t.Fatalf("search problems: %v", err)
		}

		index := slices.IndexFunc(results, func(item domain.SearchResult) bool {
			return item.ID == created.ID
		})
		if index < 0 {
			t.Fatalf("expected created problem in search results, got %d results", len(results))
		}
		if !strings.Contains(results[index].Snippet, "<mark>") {
			t.Fatalf("expected highlighted snippet, got %q", results[index].Snippet)
		}
	})

	t.Run("empty_search_not_recorded", func(t *testing.T) {
		resetDatabase(t, repo.DB())

		_, _, err := svc.CreateProblem(ctx, domain.ProblemWriteInput{
			Latex:      "Default browse problem.",
			Type:       domain.ProblemTypeSolve,
			Difficulty: domain.DifficultyEasy,
		})
		if err != nil {
			t.Fatalf("create browse problem: %v", err)
		}

		results, err := svc.Search(ctx, "   ", "", nil)
		if err != nil {
			t.Fatalf("default browse search: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("expected 1 browse result, got %d", len(results))
		}

		history, err := repo.ListSearchHistory(ctx, 10)
		if err != nil {
			t.Fatalf("list search history: %v", err)
		}
		if len(history) != 0 {
			t.Fatalf("expected empty search not to create history, got %d entries", len(history))
		}
	})

	t.Run("paper_blank_lines_round_trip", func(t *testing.T) {
		resetDatabase(t, repo.DB())

		problem, _, err := svc.CreateProblem(ctx, domain.ProblemWriteInput{
			Latex:      "证明数列满足递推关系后的结论。",
			Type:       domain.ProblemTypeProof,
			Difficulty: domain.DifficultyMedium,
		})
		if err != nil {
			t.Fatalf("create problem for paper blank lines: %v", err)
		}

		paper, err := svc.CreatePaper(ctx, domain.PaperWriteInput{
			Title:  "Blank Line Smoke Paper",
			Status: domain.PaperStatusDraft,
			Layout: domain.PaperLayout{
				Columns:           1,
				FontSize:          12,
				LineHeight:        1.4,
				PaperSize:         "A4",
				ShowAnswerVersion: false,
			},
			Items: []domain.PaperItem{
				{
					ProblemID:     problem.ID,
					Score:         12,
					OrderIndex:    0,
					ImagePosition: "below",
					BlankLines:    4,
				},
			},
		})
		if err != nil {
			t.Fatalf("create paper with blank lines: %v", err)
		}

		loaded, err := repo.GetPaperDetail(ctx, paper.ID, false)
		if err != nil {
			t.Fatalf("get paper detail: %v", err)
		}
		if len(loaded.Items) != 1 || loaded.Items[0].BlankLines != 4 {
			t.Fatalf("expected paper detail to preserve blank lines, got %+v", loaded.Items)
		}
		if len(loaded.ItemDetails) != 1 || loaded.ItemDetails[0].BlankLines != 4 {
			t.Fatalf("expected paper item details to preserve blank lines, got %+v", loaded.ItemDetails)
		}

		snapshot, err := repo.ExportAll(ctx)
		if err != nil {
			t.Fatalf("export all with paper blank lines: %v", err)
		}
		index := slices.IndexFunc(snapshot.Papers, func(item domain.SeedPaper) bool {
			return item.ID == paper.ID
		})
		if index < 0 {
			t.Fatalf("expected exported snapshot to include paper %s", paper.ID)
		}
		if len(snapshot.Papers[index].Items) != 1 || snapshot.Papers[index].Items[0].BlankLines != 4 {
			t.Fatalf("expected exported snapshot to preserve blank lines, got %+v", snapshot.Papers[index].Items)
		}
	})

	t.Run("latex_export_completed", func(t *testing.T) {
		resetDatabase(t, repo.DB())

		imagePath := filepath.Join(repo.StorageRoot(), "original", "export-problem.png")
		if err := os.MkdirAll(filepath.Dir(imagePath), 0o755); err != nil {
			t.Fatalf("create export image dir: %v", err)
		}
		if err := os.WriteFile(imagePath, []byte("demo image content"), 0o644); err != nil {
			t.Fatalf("write export image file: %v", err)
		}

		image := mustCreateImage(t, repo, "export-problem.png", "original/export-problem.png", "thumbnails/export-problem.png")
		problem, _, err := svc.CreateProblem(ctx, domain.ProblemWriteInput{
			Latex:      "Export smoke test problem.",
			Type:       domain.ProblemTypeSolve,
			Difficulty: domain.DifficultyMedium,
			ImageIDs:   []string{image.ID},
		})
		if err != nil {
			t.Fatalf("create problem for export: %v", err)
		}

		paper, err := svc.CreatePaper(ctx, domain.PaperWriteInput{
			Title:  "Smoke Export Paper",
			Status: domain.PaperStatusDraft,
			Layout: domain.PaperLayout{
				Columns:           1,
				FontSize:          12,
				LineHeight:        1.4,
				PaperSize:         "A4",
				ShowAnswerVersion: false,
			},
			Items: []domain.PaperItem{
				{
					ProblemID:     problem.ID,
					Score:         10,
					OrderIndex:    0,
					ImagePosition: "below",
				},
			},
		})
		if err != nil {
			t.Fatalf("create paper: %v", err)
		}

		job, err := svc.CreateExport(ctx, domain.ExportCreateInput{
			PaperID: paper.ID,
			Format:  domain.ExportFormatLatex,
			Variant: domain.ExportVariantStudent,
		})
		if err != nil {
			t.Fatalf("create export: %v", err)
		}

		completed := waitForExport(t, repo, job.ID)
		if completed.Status != domain.ExportStatusDone {
			t.Fatalf("expected export done, got %s", completed.Status)
		}

		downloadPath, err := repo.GetExportDownloadPath(ctx, job.ID)
		if err != nil {
			t.Fatalf("get export download path: %v", err)
		}

		archiveBytes, err := os.ReadFile(filepath.Join(repo.StorageRoot(), downloadPath))
		if err != nil {
			t.Fatalf("read export artifact: %v", err)
		}
		archiveReader, err := zip.NewReader(bytes.NewReader(archiveBytes), int64(len(archiveBytes)))
		if err != nil {
			t.Fatalf("open export zip: %v", err)
		}

		var (
			mainTex      string
			hasImage     bool
			hasLatexmkrc bool
		)
		for _, file := range archiveReader.File {
			if file.Name == "main.tex" {
				reader, err := file.Open()
				if err != nil {
					t.Fatalf("open main.tex: %v", err)
				}
				body, err := io.ReadAll(reader)
				reader.Close()
				if err != nil {
					t.Fatalf("read main.tex: %v", err)
				}
				mainTex = string(body)
			}
			if strings.HasPrefix(file.Name, "images/") {
				hasImage = true
			}
			if file.Name == "latexmkrc" {
				hasLatexmkrc = true
			}
		}
		if !strings.Contains(mainTex, "Smoke Export Paper") {
			t.Fatalf("expected export artifact to contain paper title, got %q", mainTex)
		}
			if strings.Contains(mainTex, "/app/storage/") {
				t.Fatalf("expected latex export to avoid absolute storage paths, got %q", mainTex)
			}
			if !strings.Contains(mainTex, `\IfFileExists{images/`) {
				t.Fatalf("expected portable latex image guard, got %q", mainTex)
			}
			if !strings.Contains(mainTex, `\usepackage{CJKutf8}`) {
				t.Fatalf("expected latex bundle to include pdfLaTeX Chinese fallback, got %q", mainTex)
			}
			if strings.Contains(mainTex, `MathLib exported bundle requires XeLaTeX`) {
				t.Fatalf("expected latex bundle to avoid hard XeLaTeX requirement, got %q", mainTex)
			}
			if !hasImage {
				t.Fatal("expected latex bundle to contain staged image assets")
			}
			if !hasLatexmkrc {
				t.Fatal("expected latex bundle to contain latexmkrc for Overleaf")
			}
	})

	t.Run("pdf_export_tolerates_literal_newline_escapes", func(t *testing.T) {
		if _, err := exec.LookPath("xelatex"); err != nil {
			t.Skipf("skip pdf export smoke test, xelatex unavailable: %v", err)
		}

		resetDatabase(t, repo.DB())

		answerLatex := "答：先代入条件。\\n再整理结果。"
		solutionLatex := "解析：使用数学归纳法证明：\\n1. 当 \\(n=1\\) 时成立。\\n2. 假设 \\(n=k\\) 时成立，则 \\(n=k+1\\) 时..."
		problem, _, err := svc.CreateProblem(ctx, domain.ProblemWriteInput{
			Latex:         "证明数列结论成立。",
			AnswerLatex:   &answerLatex,
			SolutionLatex: &solutionLatex,
			Type:          domain.ProblemTypeProof,
			Difficulty:    domain.DifficultyMedium,
		})
		if err != nil {
			t.Fatalf("create problem for pdf export: %v", err)
		}

		paper, err := svc.CreatePaper(ctx, domain.PaperWriteInput{
			Title:  "PDF Literal Newline Smoke Paper",
			Status: domain.PaperStatusDraft,
			Layout: domain.PaperLayout{
				Columns:           1,
				FontSize:          12,
				LineHeight:        1.4,
				PaperSize:         "A4",
				ShowAnswerVersion: true,
			},
			Items: []domain.PaperItem{
				{
					ProblemID:     problem.ID,
					Score:         12,
					OrderIndex:    0,
					ImagePosition: "below",
				},
			},
		})
		if err != nil {
			t.Fatalf("create paper for pdf export: %v", err)
		}

		job, err := svc.CreateExport(ctx, domain.ExportCreateInput{
			PaperID: paper.ID,
			Format:  domain.ExportFormatPDF,
			Variant: domain.ExportVariantAnswer,
		})
		if err != nil {
			t.Fatalf("create pdf export: %v", err)
		}

		completed := waitForExport(t, repo, job.ID)
		if completed.Status != domain.ExportStatusDone {
			t.Fatalf("expected pdf export done, got %s (%v)", completed.Status, completed.ErrorMessage)
		}

		downloadPath, err := repo.GetExportDownloadPath(ctx, job.ID)
		if err != nil {
			t.Fatalf("get pdf export download path: %v", err)
		}
		if _, err := os.Stat(filepath.Join(repo.StorageRoot(), downloadPath)); err != nil {
			t.Fatalf("expected generated pdf artifact to exist: %v", err)
		}
	})

	t.Run("tag_merge", func(t *testing.T) {
		resetDatabase(t, repo.DB())

		sourceTag := mustCreateTag(t, repo, "待合并")
		targetTag := mustCreateTag(t, repo, "已保留")
		created, _, err := svc.CreateProblem(ctx, domain.ProblemWriteInput{
			Latex:      "Tag merge smoke test.",
			Type:       domain.ProblemTypeSolve,
			Difficulty: domain.DifficultyEasy,
			TagIDs:     []string{sourceTag.ID},
		})
		if err != nil {
			t.Fatalf("create problem for tag merge: %v", err)
		}

		if err := repo.MergeTag(ctx, sourceTag.ID, targetTag.ID); err != nil {
			t.Fatalf("merge tag: %v", err)
		}

		loaded, err := repo.GetProblemDetail(ctx, created.ID, true)
		if err != nil {
			t.Fatalf("load merged problem: %v", err)
		}
		if !slices.Contains(loaded.TagIDs, targetTag.ID) {
			t.Fatalf("expected merged problem to reference target tag, got %v", loaded.TagIDs)
		}
		if slices.Contains(loaded.TagIDs, sourceTag.ID) {
			t.Fatalf("expected source tag to be removed, got %v", loaded.TagIDs)
		}
	})
}

func mustCreateTag(t *testing.T, repo *store.Repository, name string) *domain.Tag {
	t.Helper()
	tag, err := repo.CreateTag(context.Background(), domain.Tag{
		Name:     name,
		Category: domain.TagCategoryTopic,
		Color:    "#0F766E",
	})
	if err != nil {
		t.Fatalf("create tag %q: %v", name, err)
	}
	return tag
}

func mustCreateImage(t *testing.T, repo *store.Repository, filename, storagePath, thumbnailPath string) *domain.ImageAsset {
	t.Helper()
	image, err := repo.CreateImageRecord(
		context.Background(),
		filename,
		"image/png",
		128,
		32,
		32,
		storagePath,
		thumbnailPath,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("create image %q: %v", filename, err)
	}
	return image
}

func waitForExport(t *testing.T, repo *store.Repository, jobID string) domain.ExportJob {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		job, err := repo.GetExportJob(context.Background(), jobID)
		if err != nil {
			t.Fatalf("load export job: %v", err)
		}
		if job.Status == domain.ExportStatusDone || job.Status == domain.ExportStatusFailed {
			return *job
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("export job %s did not finish before timeout", jobID)
	return domain.ExportJob{}
}

func applyMigrations(t *testing.T, db *pgxpool.Pool) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(migrationsDir(t), "*.sql"))
	if err != nil {
		t.Fatalf("glob migrations: %v", err)
	}
	slices.Sort(files)

	for _, path := range files {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read migration %s: %v", path, err)
		}
		for _, stmt := range splitSQLStatements(extractUpSQL(string(raw))) {
			if _, err := db.Exec(context.Background(), stmt); err != nil {
				t.Fatalf("apply migration %s: %v\nstatement:\n%s", path, err, stmt)
			}
		}
	}
}

func resetDatabase(t *testing.T, db *pgxpool.Pool) {
	t.Helper()
	statement := `TRUNCATE TABLE
  export_jobs,
  paper_items,
  papers,
  problem_versions,
  problem_images,
  problem_tags,
  problems,
  image_tags,
  images,
  tags,
  search_history,
  saved_searches,
  app_settings,
  problem_code_counters
RESTART IDENTITY CASCADE`
	if _, err := db.Exec(context.Background(), statement); err != nil {
		t.Fatalf("reset database: %v", err)
	}
}

func migrationsDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
}

func extractUpSQL(raw string) string {
	start := strings.Index(raw, "-- +goose Up")
	if start < 0 {
		return raw
	}
	raw = raw[start+len("-- +goose Up"):]
	if end := strings.Index(raw, "-- +goose Down"); end >= 0 {
		raw = raw[:end]
	}

	lines := strings.Split(raw, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "-- +goose") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
}

func splitSQLStatements(raw string) []string {
	statements := make([]string, 0)
	var builder strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	dollarTag := ""

	for i := 0; i < len(raw); i++ {
		if dollarTag != "" && strings.HasPrefix(raw[i:], dollarTag) {
			builder.WriteString(dollarTag)
			i += len(dollarTag) - 1
			dollarTag = ""
			continue
		}

		ch := raw[i]
		if !inSingleQuote && !inDoubleQuote && dollarTag == "" && ch == '$' {
			if tag, ok := scanDollarTag(raw[i:]); ok {
				dollarTag = tag
				builder.WriteString(tag)
				i += len(tag) - 1
				continue
			}
		}

		if ch == '\'' && !inDoubleQuote && dollarTag == "" {
			if i == 0 || raw[i-1] != '\\' {
				inSingleQuote = !inSingleQuote
			}
			builder.WriteByte(ch)
			continue
		}
		if ch == '"' && !inSingleQuote && dollarTag == "" {
			if i == 0 || raw[i-1] != '\\' {
				inDoubleQuote = !inDoubleQuote
			}
			builder.WriteByte(ch)
			continue
		}
		if ch == ';' && !inSingleQuote && !inDoubleQuote && dollarTag == "" {
			stmt := strings.TrimSpace(builder.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			builder.Reset()
			continue
		}
		builder.WriteByte(ch)
	}

	stmt := strings.TrimSpace(builder.String())
	if stmt != "" {
		statements = append(statements, stmt)
	}
	return statements
}

func scanDollarTag(raw string) (string, bool) {
	if len(raw) == 0 || raw[0] != '$' {
		return "", false
	}
	for i := 1; i < len(raw); i++ {
		ch := raw[i]
		if ch == '$' {
			return raw[:i+1], true
		}
		if ch != '_' && (ch < 'a' || ch > 'z') && (ch < 'A' || ch > 'Z') && (ch < '0' || ch > '9') {
			return "", false
		}
	}
	return "", false
}

func loadIsolatedTestConfig(t *testing.T) (config.Config, func()) {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	parsed, err := url.Parse(cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("parse database url: %v", err)
	}
	baseName := strings.TrimPrefix(parsed.Path, "/")
	if baseName == "" {
		baseName = "mathlib"
	}
	testDatabaseName := sanitizeDatabaseName(fmt.Sprintf("%s_smoke_%d", baseName, time.Now().UnixNano()))

	adminURL := *parsed
	adminURL.Path = "/postgres"

	adminPool, err := pgxpool.New(context.Background(), adminURL.String())
	if err != nil {
		t.Skipf("skip smoke test, admin database unavailable: %v", err)
	}
	defer adminPool.Close()

	if err := adminPool.Ping(context.Background()); err != nil {
		t.Skipf("skip smoke test, admin database unavailable: %v", err)
	}
	if _, err := adminPool.Exec(context.Background(), "CREATE DATABASE "+quoteIdentifier(testDatabaseName)); err != nil {
		t.Skipf("skip smoke test, cannot create isolated database: %v", err)
	}

	testURL := *parsed
	testURL.Path = "/" + testDatabaseName
	cfg.DatabaseURL = testURL.String()

	cleanup := func() {
		dropPool, err := pgxpool.New(context.Background(), adminURL.String())
		if err != nil {
			return
		}
		defer dropPool.Close()

		_, _ = dropPool.Exec(context.Background(), `SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = $1 AND pid <> pg_backend_pid()`, testDatabaseName)
		_, _ = dropPool.Exec(context.Background(), "DROP DATABASE IF EXISTS "+quoteIdentifier(testDatabaseName))
	}

	return cfg, cleanup
}

func sanitizeDatabaseName(value string) string {
	var builder strings.Builder
	for _, ch := range value {
		switch {
		case ch >= 'a' && ch <= 'z':
			builder.WriteRune(ch)
		case ch >= 'A' && ch <= 'Z':
			builder.WriteRune(ch + ('a' - 'A'))
		case ch >= '0' && ch <= '9':
			builder.WriteRune(ch)
		default:
			builder.WriteByte('_')
		}
	}
	if builder.Len() == 0 {
		return "mathlib_smoke"
	}
	return builder.String()
}

func quoteIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func stringPtr(value string) *string {
	return &value
}
