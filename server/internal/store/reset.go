package store

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	imgutil "mathlib/server/internal/image"
	"mathlib/server/internal/search"

	"mathlib/server/internal/domain"

	"github.com/jackc/pgx/v5"
)

func (r *Repository) ResetDemoData(ctx context.Context, seed domain.DemoSeed) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `TRUNCATE TABLE
		image_tags,
		problem_tags,
		problem_images,
		problem_versions,
		paper_items,
		export_jobs,
		search_history,
		saved_searches,
		app_settings,
		papers,
		problems,
		images,
		tags,
		problem_code_counters
		RESTART IDENTITY CASCADE`)
	if err != nil {
		return err
	}

	for _, tag := range seed.Tags {
		if _, err := tx.Exec(ctx,
			`INSERT INTO tags (id, name, category, color, description, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, now(), now())`,
			tag.ID, tag.Name, tag.Category, tag.Color, tag.Description,
		); err != nil {
			return err
		}
	}

	for _, imageAsset := range seed.Images {
		storagePath := seedOriginalPath(imageAsset.ID)
		thumbnailPath := seedThumbnailPath(imageAsset.ID)
		var deletedAt any
		if imageAsset.IsDeleted {
			deletedAt = imageAsset.UpdatedAt
		}

		if _, err := tx.Exec(ctx, `INSERT INTO images
			(id, filename, mime, size_bytes, width, height, storage_path, thumbnail_path, description, created_at, updated_at, deleted_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
			imageAsset.ID,
			imageAsset.Filename,
			imageAsset.MIME,
			imageAsset.Size,
			imageAsset.Width,
			imageAsset.Height,
			storagePath,
			thumbnailPath,
			imageAsset.Description,
			imageAsset.CreatedAt,
			imageAsset.UpdatedAt,
			deletedAt,
		); err != nil {
			return err
		}
		for _, tagID := range imageAsset.TagIDs {
			if _, err := tx.Exec(ctx, `INSERT INTO image_tags (image_id, tag_id) VALUES ($1, $2)`, imageAsset.ID, tagID); err != nil {
				return err
			}
		}
	}

	for _, problem := range seed.Problems {
		var deletedAt any
		if problem.IsDeleted {
			deletedAt = problem.UpdatedAt
		}

		if _, err := tx.Exec(ctx, `INSERT INTO problems
			(id, code, latex, answer_latex, solution_latex, problem_type, difficulty, subjective_score, subject, grade, source, notes, formula_tokens, version, created_at, updated_at, deleted_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
			problem.ID,
			problem.Code,
			problem.Latex,
			problem.AnswerLatex,
			problem.SolutionLatex,
			problem.Type,
			problem.Difficulty,
			problem.SubjectiveScore,
			problem.Subject,
			problem.Grade,
			problem.Source,
			problem.Notes,
			search.TokenizeLatex(problem.Latex),
			problem.Version,
			problem.CreatedAt,
			problem.UpdatedAt,
			deletedAt,
		); err != nil {
			return err
		}

		for _, tagID := range problem.TagIDs {
			if _, err := tx.Exec(ctx, `INSERT INTO problem_tags (problem_id, tag_id) VALUES ($1, $2)`, problem.ID, tagID); err != nil {
				return err
			}
		}
		for orderIndex, imageID := range problem.ImageIDs {
			if _, err := tx.Exec(ctx, `INSERT INTO problem_images (problem_id, image_id, order_index) VALUES ($1, $2, $3)`, problem.ID, imageID, orderIndex); err != nil {
				return err
			}
		}

		if err := insertSeedProblemVersions(ctx, tx, problem, false); err != nil {
			return err
		}
	}

	for _, paper := range seed.Papers {
		headerBytes, err := marshalJSON(map[string]any{})
		if err != nil {
			return err
		}
		layoutBytes, err := marshalJSON(paper.Layout)
		if err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `INSERT INTO papers
			(id, title, subtitle, school_name, exam_name, subject, duration_min, total_score, description, status, instructions, footer_text, header_json, layout_json, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
			paper.ID,
			paper.Title,
			paper.Subtitle,
			paper.SchoolName,
			paper.ExamName,
			paper.Subject,
			paper.Duration,
			paper.TotalScore,
			paper.Description,
			paper.Status,
			paper.Instructions,
			paper.FooterText,
			headerBytes,
			layoutBytes,
			paper.CreatedAt,
			paper.UpdatedAt,
		); err != nil {
			return err
		}

		for _, item := range paper.Items {
			if _, err := tx.Exec(ctx, `INSERT INTO paper_items (id, paper_id, problem_id, order_index, score, image_position, blank_lines) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				item.ID,
				paper.ID,
				item.ProblemID,
				item.OrderIndex,
				item.Score,
				item.ImagePosition,
				item.BlankLines,
			); err != nil {
				return err
			}
		}
	}

	for _, job := range seed.ExportJobs {
		var downloadPath any
		if job.Status == domain.ExportStatusDone {
			downloadPath = seedExportPath(job.ID, job.Format)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO export_jobs
			(id, paper_id, paper_title, format, variant, status, progress, download_path, error_message, created_at, completed_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
			job.ID,
			job.PaperID,
			job.PaperTitle,
			job.Format,
			job.Variant,
			job.Status,
			job.Progress,
			downloadPath,
			job.ErrorMessage,
			job.CreatedAt,
			job.CompletedAt,
		); err != nil {
			return err
		}
	}

	for _, entry := range seed.SearchHistory {
		raw, err := marshalJSON(entry.Filters)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO search_history (id, query, filters, result_count, created_at) VALUES ($1, $2, $3, $4, $5)`,
			entry.ID,
			entry.Query,
			raw,
			entry.ResultCount,
			entry.CreatedAt,
		); err != nil {
			return err
		}
	}

	for _, entry := range seed.SavedSearches {
		raw, err := marshalJSON(entry.Filters)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO saved_searches (id, name, query, filters, created_at) VALUES ($1, $2, $3, $4, $5)`,
			entry.ID,
			entry.Name,
			entry.Query,
			raw,
			entry.CreatedAt,
		); err != nil {
			return err
		}
	}

	for key, value := range seed.Settings {
		raw, err := marshalJSON(value)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO app_settings (key, value, updated_at) VALUES ($1, $2, now())`, key, raw); err != nil {
			return err
		}
	}

	counterMap := map[int]int{}
	for _, problem := range seed.Problems {
		year, serial := parseProblemCode(problem.Code)
		if year == 0 {
			continue
		}
		if serial > counterMap[year] {
			counterMap[year] = serial
		}
	}
	for year, serial := range counterMap {
		if _, err := tx.Exec(ctx, `INSERT INTO problem_code_counters (year, current_serial, updated_at) VALUES ($1, $2, now())`, year, serial); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	for _, imageAsset := range seed.Images {
		originalAbs := filepath.Join(r.storageRoot, seedOriginalPath(imageAsset.ID))
		thumbnailAbs := filepath.Join(r.storageRoot, seedThumbnailPath(imageAsset.ID))
		if err := imgutil.WritePlaceholderPNG(originalAbs, imageAsset.Width, imageAsset.Height, seedColor(imageAsset.ID)); err != nil {
			return err
		}
		if err := imgutil.WritePlaceholderPNG(thumbnailAbs, min(imageAsset.Width, 400), min(imageAsset.Height, 300), seedColor("thumb-"+imageAsset.ID)); err != nil {
			return err
		}
	}

	for _, job := range seed.ExportJobs {
		if job.Status != domain.ExportStatusDone {
			continue
		}
		artifactPath := filepath.Join(r.storageRoot, seedExportPath(job.ID, job.Format))
		if err := writeSeedExportArtifact(artifactPath, job); err != nil {
			return err
		}
	}

	return nil
}

func insertSeedProblemVersions(ctx context.Context, tx pgx.Tx, problem domain.SeedProblem, ignoreConflicts bool) error {
	entries := seedProblemVersionEntries(problem)
	conflictClause := ""
	if ignoreConflicts {
		conflictClause = " ON CONFLICT DO NOTHING"
	}

	for _, entry := range entries {
		snapshot := entry.Snapshot
		if len(snapshot) == 0 {
			snapshot = seedProblemSnapshot(problem)
		}
		raw, err := marshalJSON(snapshot)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO problem_versions (id, problem_id, version, snapshot, created_at) VALUES ($1, $2, $3, $4, $5)`+conflictClause,
			fmt.Sprintf("%s-v%d", problem.ID, entry.Version),
			problem.ID,
			entry.Version,
			raw,
			entry.CreatedAt,
		); err != nil {
			return err
		}
	}

	return nil
}

func seedProblemVersionEntries(problem domain.SeedProblem) []domain.SeedProblemVersion {
	entries := make([]domain.SeedProblemVersion, 0, len(problem.Versions)+1)
	if len(problem.Versions) > 0 {
		entries = append(entries, problem.Versions...)
	}

	hasCurrent := false
	for _, entry := range entries {
		if entry.Version == problem.Version {
			hasCurrent = true
			break
		}
	}
	if !hasCurrent {
		entries = append(entries, domain.SeedProblemVersion{
			Version:   problem.Version,
			CreatedAt: problem.UpdatedAt,
			Snapshot:  seedProblemSnapshot(problem),
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Version == entries[j].Version {
			return entries[i].CreatedAt.Before(entries[j].CreatedAt)
		}
		return entries[i].Version < entries[j].Version
	})
	return entries
}

func seedProblemSnapshot(problem domain.SeedProblem) map[string]any {
	return map[string]any{
		"id":              problem.ID,
		"code":            problem.Code,
		"latex":           problem.Latex,
		"answerLatex":     problem.AnswerLatex,
		"solutionLatex":   problem.SolutionLatex,
		"type":            problem.Type,
		"difficulty":      problem.Difficulty,
		"subjectiveScore": problem.SubjectiveScore,
		"subject":         problem.Subject,
		"grade":           problem.Grade,
		"source":          problem.Source,
		"tagIds":          problem.TagIDs,
		"imageIds":        problem.ImageIDs,
		"notes":           problem.Notes,
		"createdAt":       problem.CreatedAt,
		"updatedAt":       problem.UpdatedAt,
		"version":         problem.Version,
		"isDeleted":       problem.IsDeleted,
	}
}

func parseProblemCode(code string) (int, int) {
	parts := strings.Split(code, "-")
	if len(parts) != 3 {
		return 0, 0
	}
	year, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0
	}
	serial, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0
	}
	return year, serial
}

func seedOriginalPath(id string) string {
	return filepath.Join("original", id[:2], id+".png")
}

func seedThumbnailPath(id string) string {
	return filepath.Join("thumbnails", id[:2], id+".png")
}

func seedExportPath(id string, format domain.ExportFormat) string {
	ext := ".pdf"
	if format == domain.ExportFormatLatex {
		ext = ".zip"
	}
	return filepath.Join("derived", "exports", id+ext)
}

func seedColor(seed string) colorRGBA {
	var total int
	for _, ch := range seed {
		total += int(ch)
	}
	return colorRGBA{
		R: uint8(90 + total%100),
		G: uint8(120 + (total/3)%90),
		B: uint8(140 + (total/5)%70),
		A: 255,
	}
}

type colorRGBA struct {
	R, G, B, A uint8
}

func (c colorRGBA) RGBA() (r uint32, g uint32, b uint32, a uint32) {
	r = uint32(c.R)
	r |= r << 8
	g = uint32(c.G)
	g |= g << 8
	b = uint32(c.B)
	b |= b << 8
	a = uint32(c.A)
	a |= a << 8
	return
}

func writeSeedExportArtifact(path string, job domain.SeedExportJob) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if job.Format == domain.ExportFormatLatex {
		return writeSeedLatexBundle(path, job)
	}
	return os.WriteFile(path, minimalPDF(job.PaperTitle), 0o644)
}

func writeSeedLatexBundle(path string, job domain.SeedExportJob) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	mainTex, err := zipWriter.Create("main.tex")
	if err != nil {
		return err
	}
	content := fmt.Sprintf("%% !TeX program = xelatex\n%% Auto generated seed export\n%% %s\n\\RequirePackage{iftex}\n\\documentclass{article}\n\\ifPDFTeX\n\\usepackage[utf8]{inputenc}\n\\usepackage[T1]{fontenc}\n\\usepackage{CJKutf8}\n\\newcommand{\\MathLibBeginDocument}{\\begin{CJK*}{UTF8}{gbsn}}\n\\newcommand{\\MathLibEndDocument}{\\end{CJK*}}\n\\else\n\\usepackage[UTF8]{ctex}\n\\newcommand{\\MathLibBeginDocument}{}\n\\newcommand{\\MathLibEndDocument}{}\n\\fi\n\\begin{document}\n\\MathLibBeginDocument\n%s\n\\MathLibEndDocument\n\\end{document}\n", job.PaperTitle, job.PaperTitle)
	if _, err := mainTex.Write([]byte(content)); err != nil {
		return err
	}

	readme, err := zipWriter.Create("README.txt")
	if err != nil {
		return err
	}
	readmeContent := "MathLib demo LaTeX bundle\nThe project prefers XeLaTeX but also includes a pdfLaTeX fallback for Overleaf.\nAdd any required figures under the images/ folder.\n"
	if _, err := readme.Write([]byte(readmeContent)); err != nil {
		return err
	}

	latexmkrc, err := zipWriter.Create("latexmkrc")
	if err != nil {
		return err
	}
	if _, err := latexmkrc.Write([]byte("$pdf_mode = 5;\n$xelatex = 'xelatex -interaction=nonstopmode -file-line-error -synctex=1 %O %S';\n")); err != nil {
		return err
	}

	return nil
}

func minimalPDF(title string) []byte {
	stream := fmt.Sprintf("BT /F1 18 Tf 72 720 Td (%s) Tj ET", strings.NewReplacer("(", "\\(", ")", "\\)").Replace(title))
	body := fmt.Sprintf("1 0 obj<< /Type /Catalog /Pages 2 0 R >>endobj\n2 0 obj<< /Type /Pages /Kids [3 0 R] /Count 1 >>endobj\n3 0 obj<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>endobj\n4 0 obj<< /Length %d >>stream\n%s\nendstream\nendobj\n5 0 obj<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>endobj\n", len(stream), stream)
	prefix := "%PDF-1.4\n"
	xrefStart := len(prefix) + len(body)
	xref := "xref\n0 6\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \n0000000241 00000 n \n0000000321 00000 n \ntrailer<< /Root 1 0 R /Size 6 >>\nstartxref\n" + strconv.Itoa(xrefStart) + "\n%%EOF"
	return []byte(prefix + body + xref)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (r *Repository) NextProblemCode(ctx context.Context, tx pgx.Tx) (string, error) {
	year := time.Now().Year()
	var current int
	err := tx.QueryRow(ctx, `INSERT INTO problem_code_counters (year, current_serial, updated_at)
		VALUES ($1, 1, now())
		ON CONFLICT (year)
		DO UPDATE SET current_serial = problem_code_counters.current_serial + 1, updated_at = now()
		RETURNING current_serial`,
		year,
	).Scan(&current)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("P-%d-%04d", year, current), nil
}

func (r *Repository) LoadDemoData(ctx context.Context, seed domain.DemoSeed) (map[string]int, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	stats := map[string]int{
		"tags":            0,
		"problems":        0,
		"images":          0,
		"papers":          0,
		"problemVersions": 0,
		"exportJobs":      0,
		"searchHistory":   0,
		"savedSearches":   0,
	}

	for _, tag := range seed.Tags {
		if !strings.HasPrefix(tag.ID, "demo-") {
			continue
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO tags (id, name, category, color, description, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, now(), now())
			 ON CONFLICT (id) DO NOTHING`,
			tag.ID, tag.Name, tag.Category, tag.Color, tag.Description,
		); err != nil {
			return nil, err
		}
		stats["tags"]++
	}

	for _, imageAsset := range seed.Images {
		if !strings.HasPrefix(imageAsset.ID, "demo-") {
			continue
		}
		storagePath := seedOriginalPath(imageAsset.ID)
		thumbnailPath := seedThumbnailPath(imageAsset.ID)
		var deletedAt any
		if imageAsset.IsDeleted {
			deletedAt = imageAsset.UpdatedAt
		}

		if _, err := tx.Exec(ctx, `INSERT INTO images
			(id, filename, mime, size_bytes, width, height, storage_path, thumbnail_path, description, created_at, updated_at, deleted_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
			ON CONFLICT (id) DO NOTHING`,
			imageAsset.ID,
			imageAsset.Filename,
			imageAsset.MIME,
			imageAsset.Size,
			imageAsset.Width,
			imageAsset.Height,
			storagePath,
			thumbnailPath,
			imageAsset.Description,
			imageAsset.CreatedAt,
			imageAsset.UpdatedAt,
			deletedAt,
		); err != nil {
			return nil, err
		}
		for _, tagID := range imageAsset.TagIDs {
			if _, err := tx.Exec(ctx, `INSERT INTO image_tags (image_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, imageAsset.ID, tagID); err != nil {
				return nil, err
			}
		}
		stats["images"]++
	}

	for _, problem := range seed.Problems {
		if !strings.HasPrefix(problem.ID, "demo-") {
			continue
		}
		var deletedAt any
		if problem.IsDeleted {
			deletedAt = problem.UpdatedAt
		}

		if _, err := tx.Exec(ctx, `INSERT INTO problems
			(id, code, latex, answer_latex, solution_latex, problem_type, difficulty, subjective_score, subject, grade, source, notes, formula_tokens, version, created_at, updated_at, deleted_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
			ON CONFLICT (id) DO NOTHING`,
			problem.ID,
			problem.Code,
			problem.Latex,
			problem.AnswerLatex,
			problem.SolutionLatex,
			problem.Type,
			problem.Difficulty,
			problem.SubjectiveScore,
			problem.Subject,
			problem.Grade,
			problem.Source,
			problem.Notes,
			search.TokenizeLatex(problem.Latex),
			problem.Version,
			problem.CreatedAt,
			problem.UpdatedAt,
			deletedAt,
		); err != nil {
			return nil, err
		}

		for _, tagID := range problem.TagIDs {
			if _, err := tx.Exec(ctx, `INSERT INTO problem_tags (problem_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, problem.ID, tagID); err != nil {
				return nil, err
			}
		}
		for orderIndex, imageID := range problem.ImageIDs {
			if _, err := tx.Exec(ctx, `INSERT INTO problem_images (problem_id, image_id, order_index) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, problem.ID, imageID, orderIndex); err != nil {
				return nil, err
			}
		}

		if err := insertSeedProblemVersions(ctx, tx, problem, true); err != nil {
			return nil, err
		}
		stats["problemVersions"] += len(seedProblemVersionEntries(problem))
		stats["problems"]++
	}

	for _, paper := range seed.Papers {
		if !strings.HasPrefix(paper.ID, "demo-") {
			continue
		}
		headerBytes, err := marshalJSON(map[string]any{})
		if err != nil {
			return nil, err
		}
		layoutBytes, err := marshalJSON(paper.Layout)
		if err != nil {
			return nil, err
		}

		if _, err := tx.Exec(ctx, `INSERT INTO papers
			(id, title, subtitle, school_name, exam_name, subject, duration_min, total_score, description, status, instructions, footer_text, header_json, layout_json, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
			ON CONFLICT (id) DO NOTHING`,
			paper.ID,
			paper.Title,
			paper.Subtitle,
			paper.SchoolName,
			paper.ExamName,
			paper.Subject,
			paper.Duration,
			paper.TotalScore,
			paper.Description,
			paper.Status,
			paper.Instructions,
			paper.FooterText,
			headerBytes,
			layoutBytes,
			paper.CreatedAt,
			paper.UpdatedAt,
		); err != nil {
			return nil, err
		}

		for _, item := range paper.Items {
			if _, err := tx.Exec(ctx, `INSERT INTO paper_items (id, paper_id, problem_id, order_index, score, image_position, blank_lines) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING`,
				item.ID,
				paper.ID,
				item.ProblemID,
				item.OrderIndex,
				item.Score,
				item.ImagePosition,
				item.BlankLines,
			); err != nil {
				return nil, err
			}
		}
		stats["papers"]++
	}

	for _, job := range seed.ExportJobs {
		if !strings.HasPrefix(job.ID, "demo-") && !strings.HasPrefix(job.PaperID, "demo-") {
			continue
		}
		var downloadPath any
		if job.Status == domain.ExportStatusDone {
			downloadPath = seedExportPath(job.ID, job.Format)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO export_jobs
			(id, paper_id, paper_title, format, variant, status, progress, download_path, error_message, created_at, completed_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
			ON CONFLICT (id) DO NOTHING`,
			job.ID,
			job.PaperID,
			job.PaperTitle,
			job.Format,
			job.Variant,
			job.Status,
			job.Progress,
			downloadPath,
			job.ErrorMessage,
			job.CreatedAt,
			job.CompletedAt,
		); err != nil {
			return nil, err
		}
		stats["exportJobs"]++
	}

	for _, entry := range seed.SearchHistory {
		if !strings.HasPrefix(entry.ID, "demo-") {
			continue
		}
		raw, err := marshalJSON(entry.Filters)
		if err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO search_history (id, query, filters, result_count, created_at) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id) DO NOTHING`,
			entry.ID,
			entry.Query,
			raw,
			entry.ResultCount,
			entry.CreatedAt,
		); err != nil {
			return nil, err
		}
		stats["searchHistory"]++
	}

	for _, entry := range seed.SavedSearches {
		if !strings.HasPrefix(entry.ID, "demo-") {
			continue
		}
		raw, err := marshalJSON(entry.Filters)
		if err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO saved_searches (id, name, query, filters, created_at) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id) DO NOTHING`,
			entry.ID,
			entry.Name,
			entry.Query,
			raw,
			entry.CreatedAt,
		); err != nil {
			return nil, err
		}
		stats["savedSearches"]++
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	for _, imageAsset := range seed.Images {
		if !strings.HasPrefix(imageAsset.ID, "demo-") {
			continue
		}
		originalAbs := filepath.Join(r.storageRoot, seedOriginalPath(imageAsset.ID))
		thumbnailAbs := filepath.Join(r.storageRoot, seedThumbnailPath(imageAsset.ID))
		if err := imgutil.WritePlaceholderPNG(originalAbs, imageAsset.Width, imageAsset.Height, seedColor(imageAsset.ID)); err != nil {
			return nil, err
		}
		if err := imgutil.WritePlaceholderPNG(thumbnailAbs, min(imageAsset.Width, 400), min(imageAsset.Height, 300), seedColor("thumb-"+imageAsset.ID)); err != nil {
			return nil, err
		}
	}

	for _, job := range seed.ExportJobs {
		if job.Status != domain.ExportStatusDone {
			continue
		}
		if !strings.HasPrefix(job.ID, "demo-") && !strings.HasPrefix(job.PaperID, "demo-") {
			continue
		}
		artifactPath := filepath.Join(r.storageRoot, seedExportPath(job.ID, job.Format))
		if err := writeSeedExportArtifact(artifactPath, job); err != nil {
			return nil, err
		}
	}

	return stats, nil
}

func (r *Repository) ClearDemoData(ctx context.Context) (map[string]int, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	stats := map[string]int{
		"tags":            0,
		"problems":        0,
		"images":          0,
		"papers":          0,
		"problemVersions": 0,
		"exportJobs":      0,
		"searchHistory":   0,
		"savedSearches":   0,
	}

	var exportJobCount int
	if err := tx.QueryRow(ctx, `WITH deleted AS (DELETE FROM export_jobs WHERE id LIKE 'demo-%' OR paper_id LIKE 'demo-%' RETURNING 1) SELECT COUNT(*) FROM deleted`).Scan(&exportJobCount); err != nil {
		return nil, err
	}
	stats["exportJobs"] = exportJobCount

	var searchHistoryCount int
	if err := tx.QueryRow(ctx, `WITH deleted AS (DELETE FROM search_history WHERE id LIKE 'demo-%' RETURNING 1) SELECT COUNT(*) FROM deleted`).Scan(&searchHistoryCount); err != nil {
		return nil, err
	}
	stats["searchHistory"] = searchHistoryCount

	var savedSearchCount int
	if err := tx.QueryRow(ctx, `WITH deleted AS (DELETE FROM saved_searches WHERE id LIKE 'demo-%' RETURNING 1) SELECT COUNT(*) FROM deleted`).Scan(&savedSearchCount); err != nil {
		return nil, err
	}
	stats["savedSearches"] = savedSearchCount

	var problemVersionCount int
	if err := tx.QueryRow(ctx, `WITH deleted AS (DELETE FROM problem_versions WHERE problem_id LIKE 'demo-%' RETURNING 1) SELECT COUNT(*) FROM deleted`).Scan(&problemVersionCount); err != nil {
		return nil, err
	}
	stats["problemVersions"] = problemVersionCount

	var paperCount int
	if err := tx.QueryRow(ctx, `WITH deleted AS (DELETE FROM papers WHERE id LIKE 'demo-%' RETURNING 1) SELECT COUNT(*) FROM deleted`).Scan(&paperCount); err != nil {
		return nil, err
	}
	stats["papers"] = paperCount

	var problemCount int
	if err := tx.QueryRow(ctx, `WITH deleted AS (DELETE FROM problems WHERE id LIKE 'demo-%' RETURNING 1) SELECT COUNT(*) FROM deleted`).Scan(&problemCount); err != nil {
		return nil, err
	}
	stats["problems"] = problemCount

	var imageCount int
	if err := tx.QueryRow(ctx, `WITH deleted AS (DELETE FROM images WHERE id LIKE 'demo-%' RETURNING 1) SELECT COUNT(*) FROM deleted`).Scan(&imageCount); err != nil {
		return nil, err
	}
	stats["images"] = imageCount

	var tagCount int
	if err := tx.QueryRow(ctx, `WITH deleted AS (DELETE FROM tags WHERE id LIKE 'demo-%' RETURNING 1) SELECT COUNT(*) FROM deleted`).Scan(&tagCount); err != nil {
		return nil, err
	}
	stats["tags"] = tagCount

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *Repository) GetDemoDataStatus(ctx context.Context) (bool, map[string]int, error) {
	stats := map[string]int{
		"tags":            0,
		"problems":        0,
		"images":          0,
		"papers":          0,
		"problemVersions": 0,
		"exportJobs":      0,
		"searchHistory":   0,
		"savedSearches":   0,
	}

	var tagCount, problemCount, imageCount, paperCount, problemVersionCount, exportJobCount, searchHistoryCount, savedSearchCount int

	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM tags WHERE id LIKE 'demo-%'`).Scan(&tagCount); err != nil {
		return false, nil, err
	}
	stats["tags"] = tagCount

	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM problems WHERE id LIKE 'demo-%'`).Scan(&problemCount); err != nil {
		return false, nil, err
	}
	stats["problems"] = problemCount

	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM images WHERE id LIKE 'demo-%'`).Scan(&imageCount); err != nil {
		return false, nil, err
	}
	stats["images"] = imageCount

	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM papers WHERE id LIKE 'demo-%'`).Scan(&paperCount); err != nil {
		return false, nil, err
	}
	stats["papers"] = paperCount

	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM problem_versions WHERE problem_id LIKE 'demo-%'`).Scan(&problemVersionCount); err != nil {
		return false, nil, err
	}
	stats["problemVersions"] = problemVersionCount

	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM export_jobs WHERE id LIKE 'demo-%' OR paper_id LIKE 'demo-%'`).Scan(&exportJobCount); err != nil {
		return false, nil, err
	}
	stats["exportJobs"] = exportJobCount

	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM search_history WHERE id LIKE 'demo-%'`).Scan(&searchHistoryCount); err != nil {
		return false, nil, err
	}
	stats["searchHistory"] = searchHistoryCount

	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM saved_searches WHERE id LIKE 'demo-%'`).Scan(&savedSearchCount); err != nil {
		return false, nil, err
	}
	stats["savedSearches"] = savedSearchCount

	loaded := false
	for _, count := range stats {
		if count > 0 {
			loaded = true
			break
		}
	}

	return loaded, stats, nil
}
