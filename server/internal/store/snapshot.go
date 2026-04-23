package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"mathlib/server/internal/domain"
)

type Snapshot struct {
	Tags     []domain.Tag
	Problems []domain.ProblemDetail
	Images   []domain.ImageAsset
	Papers   []domain.PaperDetail
	Exports  []domain.ExportJob
	Settings domain.SettingsPayload
}

type problemRow struct {
	ID            string
	Code          string
	Latex         string
	AnswerLatex   sql.NullString
	SolutionLatex sql.NullString
	ProblemType   string
	Difficulty    string
	Subjective    sql.NullFloat64
	Subject       sql.NullString
	Grade         sql.NullString
	Source        sql.NullString
	Notes         sql.NullString
	FormulaTokens sql.NullString
	Version       int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     sql.NullTime
}

type imageRow struct {
	ID            string
	Filename      string
	MIME          string
	Size          int64
	Width         int
	Height        int
	StoragePath   string
	ThumbnailPath string
	Description   sql.NullString
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     sql.NullTime
}

type paperRow struct {
	ID           string
	Title        string
	Subtitle     sql.NullString
	SchoolName   sql.NullString
	ExamName     sql.NullString
	Subject      sql.NullString
	DurationMin  sql.NullInt64
	TotalScore   sql.NullFloat64
	Description  sql.NullString
	Status       string
	Instructions sql.NullString
	FooterText   sql.NullString
	HeaderJSON   []byte
	LayoutJSON   []byte
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    sql.NullTime
}

type exportRow struct {
	ID           string
	PaperID      string
	PaperTitle   string
	Format       string
	Variant      string
	Status       string
	Progress     int
	DownloadPath sql.NullString
	ErrorMessage sql.NullString
	CreatedAt    time.Time
	StartedAt    sql.NullTime
	CompletedAt  sql.NullTime
}

type paperItemRow struct {
	ID            string
	PaperID       string
	ProblemID     string
	OrderIndex    int
	Score         float64
	ImagePosition sql.NullString
	BlankLines    int
}

func (r *Repository) LoadSnapshot(ctx context.Context, includeDeleted bool) (Snapshot, error) {
	tags, err := r.loadTags(ctx)
	if err != nil {
		return Snapshot{}, err
	}

	problemRows, err := r.loadProblemRows(ctx, includeDeleted)
	if err != nil {
		return Snapshot{}, err
	}

	imageRows, err := r.loadImageRows(ctx, includeDeleted)
	if err != nil {
		return Snapshot{}, err
	}

	paperRows, err := r.loadPaperRows(ctx, includeDeleted)
	if err != nil {
		return Snapshot{}, err
	}

	exportRows, err := r.loadExportRows(ctx)
	if err != nil {
		return Snapshot{}, err
	}

	settings, err := r.loadSettings(ctx)
	if err != nil {
		return Snapshot{}, err
	}

	problemTagIDs, err := r.loadStringRelations(ctx, "SELECT problem_id, tag_id FROM problem_tags")
	if err != nil {
		return Snapshot{}, err
	}

	imageTagIDs, err := r.loadStringRelations(ctx, "SELECT image_id, tag_id FROM image_tags")
	if err != nil {
		return Snapshot{}, err
	}

	problemImageIDs, imageProblemIDs, err := r.loadProblemImageRelations(ctx)
	if err != nil {
		return Snapshot{}, err
	}

	paperItemRows, err := r.loadPaperItemRows(ctx)
	if err != nil {
		return Snapshot{}, err
	}

	tagByID := make(map[string]domain.Tag, len(tags))
	for _, tag := range tags {
		tagByID[tag.ID] = tag
	}

	images := make([]domain.ImageAsset, 0, len(imageRows))
	imageByID := make(map[string]domain.ImageAsset, len(imageRows))
	for _, row := range imageRows {
		asset := domain.ImageAsset{
			ID:               row.ID,
			Filename:         row.Filename,
			MIME:             row.MIME,
			Size:             row.Size,
			Width:            row.Width,
			Height:           row.Height,
			URL:              fmt.Sprintf("%s/api/v1/images/%s/file", r.publicBaseURL, row.ID),
			ThumbnailURL:     fmt.Sprintf("%s/api/v1/images/%s/thumbnail", r.publicBaseURL, row.ID),
			TagIDs:           copyStrings(imageTagIDs[row.ID]),
			LinkedProblemIDs: copyStrings(imageProblemIDs[row.ID]),
			Description:      nullStringPtr(row.Description),
			CreatedAt:        row.CreatedAt,
			UpdatedAt:        row.UpdatedAt,
			IsDeleted:        row.DeletedAt.Valid,
		}
		imageByID[asset.ID] = asset
		images = append(images, asset)
	}

	problems := make([]domain.ProblemDetail, 0, len(problemRows))
	problemByID := make(map[string]domain.ProblemDetail, len(problemRows))
	for _, row := range problemRows {
		tagIDs := copyStrings(problemTagIDs[row.ID])
		imageIDs := copyStrings(problemImageIDs[row.ID])

		detail := domain.ProblemDetail{
			Problem: domain.Problem{
				ID:              row.ID,
				Code:            row.Code,
				Latex:           row.Latex,
				AnswerLatex:     nullStringPtr(row.AnswerLatex),
				SolutionLatex:   nullStringPtr(row.SolutionLatex),
				Type:            domain.ProblemType(row.ProblemType),
				Difficulty:      domain.Difficulty(row.Difficulty),
				SubjectiveScore: nullFloatPtr(row.Subjective),
				Subject:         nullStringPtr(row.Subject),
				Grade:           nullStringPtr(row.Grade),
				Source:          nullStringPtr(row.Source),
				TagIDs:          tagIDs,
				ImageIDs:        imageIDs,
				Notes:           nullStringPtr(row.Notes),
				CreatedAt:       row.CreatedAt,
				UpdatedAt:       row.UpdatedAt,
				Version:         row.Version,
				IsDeleted:       row.DeletedAt.Valid,
			},
			Tags:   make([]domain.Tag, 0, len(tagIDs)),
			Images: make([]domain.ImageAsset, 0, len(imageIDs)),
		}

		for _, tagID := range tagIDs {
			if tag, ok := tagByID[tagID]; ok {
				detail.Tags = append(detail.Tags, tag)
			}
		}

		for _, imageID := range imageIDs {
			if imageAsset, ok := imageByID[imageID]; ok {
				detail.Images = append(detail.Images, imageAsset)
			}
		}

		problemByID[detail.ID] = detail
		problems = append(problems, detail)
	}

	papers := make([]domain.PaperDetail, 0, len(paperRows))
	for _, row := range paperRows {
		var header map[string]any
		if len(row.HeaderJSON) > 0 {
			if err := json.Unmarshal(row.HeaderJSON, &header); err != nil {
				return Snapshot{}, fmt.Errorf("decode paper header: %w", err)
			}
		}
		if header == nil {
			header = map[string]any{}
		}

		layout := domain.PaperLayout{
			Columns:           1,
			FontSize:          12,
			LineHeight:        1.3,
			PaperSize:         "A4",
			ShowAnswerVersion: true,
		}
		if len(row.LayoutJSON) > 0 {
			if err := json.Unmarshal(row.LayoutJSON, &layout); err != nil {
				return Snapshot{}, fmt.Errorf("decode paper layout: %w", err)
			}
		}

		detail := domain.PaperDetail{
			Paper: domain.Paper{
				ID:         row.ID,
				Title:      row.Title,
				Subtitle:   nullStringPtr(row.Subtitle),
				SchoolName: nullStringPtr(row.SchoolName),
				ExamName:   nullStringPtr(row.ExamName),
				Subject:    nullStringPtr(row.Subject),
				Duration:   nullIntPtr(row.DurationMin),
				TotalScore: nullFloatPtr(row.TotalScore),
				Items:      make([]domain.PaperItem, 0),
				Layout:     layout,
				CreatedAt:  row.CreatedAt,
				UpdatedAt:  row.UpdatedAt,
			},
			Description:  nullStringPtr(row.Description),
			Status:       domain.PaperStatus(row.Status),
			Instructions: nullStringPtr(row.Instructions),
			FooterText:   nullStringPtr(row.FooterText),
			Header:       header,
			ItemDetails:  make([]domain.PaperItemDetail, 0),
		}

		for _, itemRow := range paperItemRows {
			if itemRow.PaperID != row.ID {
				continue
			}
			item := domain.PaperItem{
				ID:            itemRow.ID,
				ProblemID:     itemRow.ProblemID,
				Score:         itemRow.Score,
				OrderIndex:    itemRow.OrderIndex,
				ImagePosition: nullStringValue(itemRow.ImagePosition, "below"),
				BlankLines:    itemRow.BlankLines,
			}
			detail.Items = append(detail.Items, item)

			var problem *domain.ProblemDetail
			if loaded, ok := problemByID[item.ProblemID]; ok {
				copyProblem := loaded
				problem = &copyProblem
			}

			detail.ItemDetails = append(detail.ItemDetails, domain.PaperItemDetail{
				PaperItem: item,
				Problem:   problem,
			})
		}

		slices.SortFunc(detail.Items, func(a, b domain.PaperItem) int { return a.OrderIndex - b.OrderIndex })
		slices.SortFunc(detail.ItemDetails, func(a, b domain.PaperItemDetail) int { return a.OrderIndex - b.OrderIndex })
		papers = append(papers, detail)
	}

	exports := make([]domain.ExportJob, 0, len(exportRows))
	for _, row := range exportRows {
		job := domain.ExportJob{
			ID:           row.ID,
			PaperID:      row.PaperID,
			PaperTitle:   row.PaperTitle,
			Format:       domain.ExportFormat(row.Format),
			Variant:      domain.ExportVariant(row.Variant),
			Status:       domain.ExportStatus(row.Status),
			Progress:     row.Progress,
			DownloadURL:  nil,
			ErrorMessage: nullStringPtr(row.ErrorMessage),
			CreatedAt:    row.CreatedAt,
			StartedAt:    nullTimePtr(row.StartedAt),
			CompletedAt:  nullTimePtr(row.CompletedAt),
		}
		if row.DownloadPath.Valid {
			url := fmt.Sprintf("%s/api/v1/exports/%s/download", r.publicBaseURL, row.ID)
			job.DownloadURL = &url
		}
		exports = append(exports, job)
	}

	// Fill dynamic problemCount after relationships are known.
	tagProblemCounts := map[string]int{}
	for _, problem := range problems {
		if problem.IsDeleted {
			continue
		}
		for _, tagID := range problem.TagIDs {
			tagProblemCounts[tagID]++
		}
	}

	for idx := range tags {
		tags[idx].ProblemCount = tagProblemCounts[tags[idx].ID]
	}
	for idx := range problems {
		problems[idx].Tags = remapProblemTags(problems[idx].TagIDs, tags)
	}

	return Snapshot{
		Tags:     tags,
		Problems: problems,
		Images:   images,
		Papers:   papers,
		Exports:  exports,
		Settings: settings,
	}, nil
}

func (r *Repository) loadTags(ctx context.Context) ([]domain.Tag, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, category, color, description FROM tags ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]domain.Tag, 0)
	for rows.Next() {
		var (
			tag         domain.Tag
			description sql.NullString
		)
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Category, &tag.Color, &description); err != nil {
			return nil, err
		}
		tag.Description = nullStringPtr(description)
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (r *Repository) loadProblemRows(ctx context.Context, includeDeleted bool) ([]problemRow, error) {
	query := `SELECT id, code, latex, answer_latex, solution_latex, problem_type, difficulty,
		subjective_score::float8, subject, grade, source, notes, formula_tokens, version,
		created_at, updated_at, deleted_at
		FROM problems`
	if !includeDeleted {
		query += ` WHERE deleted_at IS NULL`
	}
	query += ` ORDER BY updated_at DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]problemRow, 0)
	for rows.Next() {
		var item problemRow
		if err := rows.Scan(
			&item.ID,
			&item.Code,
			&item.Latex,
			&item.AnswerLatex,
			&item.SolutionLatex,
			&item.ProblemType,
			&item.Difficulty,
			&item.Subjective,
			&item.Subject,
			&item.Grade,
			&item.Source,
			&item.Notes,
			&item.FormulaTokens,
			&item.Version,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) loadImageRows(ctx context.Context, includeDeleted bool) ([]imageRow, error) {
	query := `SELECT id, filename, mime, size_bytes, width, height, storage_path, thumbnail_path, description, created_at, updated_at, deleted_at FROM images`
	if !includeDeleted {
		query += ` WHERE deleted_at IS NULL`
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]imageRow, 0)
	for rows.Next() {
		var item imageRow
		if err := rows.Scan(
			&item.ID,
			&item.Filename,
			&item.MIME,
			&item.Size,
			&item.Width,
			&item.Height,
			&item.StoragePath,
			&item.ThumbnailPath,
			&item.Description,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) loadPaperRows(ctx context.Context, includeDeleted bool) ([]paperRow, error) {
	query := `SELECT id, title, subtitle, school_name, exam_name, subject, duration_min,
		total_score::float8, description, status, instructions, footer_text, header_json, layout_json,
		created_at, updated_at, deleted_at
		FROM papers`
	if !includeDeleted {
		query += ` WHERE deleted_at IS NULL`
	}
	query += ` ORDER BY updated_at DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]paperRow, 0)
	for rows.Next() {
		var item paperRow
		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Subtitle,
			&item.SchoolName,
			&item.ExamName,
			&item.Subject,
			&item.DurationMin,
			&item.TotalScore,
			&item.Description,
			&item.Status,
			&item.Instructions,
			&item.FooterText,
			&item.HeaderJSON,
			&item.LayoutJSON,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) loadPaperItemRows(ctx context.Context) ([]paperItemRow, error) {
	rows, err := r.db.Query(ctx, `SELECT id, paper_id, problem_id, order_index, score::float8, image_position, blank_lines FROM paper_items ORDER BY paper_id, order_index ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]paperItemRow, 0)
	for rows.Next() {
		var item paperItemRow
		if err := rows.Scan(&item.ID, &item.PaperID, &item.ProblemID, &item.OrderIndex, &item.Score, &item.ImagePosition, &item.BlankLines); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) loadExportRows(ctx context.Context) ([]exportRow, error) {
	rows, err := r.db.Query(ctx, `SELECT id, paper_id, paper_title, format, variant, status, progress, download_path, error_message, created_at, started_at, completed_at FROM export_jobs ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]exportRow, 0)
	for rows.Next() {
		var item exportRow
		if err := rows.Scan(
			&item.ID,
			&item.PaperID,
			&item.PaperTitle,
			&item.Format,
			&item.Variant,
			&item.Status,
			&item.Progress,
			&item.DownloadPath,
			&item.ErrorMessage,
			&item.CreatedAt,
			&item.StartedAt,
			&item.CompletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) loadSettings(ctx context.Context) (domain.SettingsPayload, error) {
	rows, err := r.db.Query(ctx, `SELECT key, value FROM app_settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(domain.SettingsPayload)
	for rows.Next() {
		var (
			key   string
			value []byte
		)
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}

		var decoded any
		if len(value) > 0 {
			if err := json.Unmarshal(value, &decoded); err != nil {
				return nil, err
			}
		}
		settings[key] = decoded
	}
	return settings, rows.Err()
}

func (r *Repository) loadStringRelations(ctx context.Context, query string) (map[string][]string, error) {
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string][]string{}
	for rows.Next() {
		var owner, target string
		if err := rows.Scan(&owner, &target); err != nil {
			return nil, err
		}
		result[owner] = append(result[owner], target)
	}
	return result, rows.Err()
}

func (r *Repository) loadProblemImageRelations(ctx context.Context) (map[string][]string, map[string][]string, error) {
	rows, err := r.db.Query(ctx, `SELECT problem_id, image_id FROM problem_images ORDER BY problem_id, order_index ASC`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	problemImageIDs := map[string][]string{}
	imageProblemIDs := map[string][]string{}
	for rows.Next() {
		var problemID, imageID string
		if err := rows.Scan(&problemID, &imageID); err != nil {
			return nil, nil, err
		}
		problemImageIDs[problemID] = append(problemImageIDs[problemID], imageID)
		imageProblemIDs[imageID] = append(imageProblemIDs[imageID], problemID)
	}
	return problemImageIDs, imageProblemIDs, rows.Err()
}

func remapProblemTags(tagIDs []string, tags []domain.Tag) []domain.Tag {
	tagByID := make(map[string]domain.Tag, len(tags))
	for _, tag := range tags {
		tagByID[tag.ID] = tag
	}

	result := make([]domain.Tag, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		if tag, ok := tagByID[tagID]; ok {
			result = append(result, tag)
		}
	}
	return result
}

func nullStringPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	value := v.String
	return &value
}

func nullFloatPtr(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	value := v.Float64
	return &value
}

func nullIntPtr(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	value := int(v.Int64)
	return &value
}

func nullTimePtr(v sql.NullTime) *time.Time {
	if !v.Valid {
		return nil
	}
	value := v.Time
	return &value
}

func nullStringValue(v sql.NullString, fallback string) string {
	if !v.Valid || v.String == "" {
		return fallback
	}
	return v.String
}

func copyStrings(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	return append([]string(nil), items...)
}
