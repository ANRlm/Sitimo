package store

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/store/sqlc"
)

func (r *Repository) UpsertSettings(ctx context.Context, payload domain.SettingsPayload) error {
	for key, value := range payload {
		raw, err := json.Marshal(value)
		if err != nil {
			return err
		}
		if err := r.queries.UpsertSetting(ctx, sqlc.UpsertSettingParams{
			Key:   key,
			Value: raw,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ExportAll(ctx context.Context) (domain.DemoSeed, error) {
	snapshot, err := r.LoadSnapshot(ctx, true)
	if err != nil {
		return domain.DemoSeed{}, err
	}
	problemVersions, err := r.loadProblemVersionsForExport(ctx)
	if err != nil {
		return domain.DemoSeed{}, err
	}
	searchHistory, err := r.listAllSearchHistory(ctx)
	if err != nil {
		return domain.DemoSeed{}, err
	}
	savedSearches, err := r.listAllSavedSearches(ctx)
	if err != nil {
		return domain.DemoSeed{}, err
	}

	seed := domain.DemoSeed{
		Tags:          make([]domain.SeedTag, 0, len(snapshot.Tags)),
		Problems:      make([]domain.SeedProblem, 0, len(snapshot.Problems)),
		Images:        make([]domain.SeedImage, 0, len(snapshot.Images)),
		Papers:        make([]domain.SeedPaper, 0, len(snapshot.Papers)),
		ExportJobs:    make([]domain.SeedExportJob, 0, len(snapshot.Exports)),
		SearchHistory: make([]domain.SeedSearchHistory, 0, len(searchHistory)),
		SavedSearches: make([]domain.SeedSavedSearchEntry, 0, len(savedSearches)),
		Settings:      snapshot.Settings,
	}

	for _, tag := range snapshot.Tags {
		seed.Tags = append(seed.Tags, domain.SeedTag{
			ID:          tag.ID,
			Name:        tag.Name,
			Category:    tag.Category,
			Color:       tag.Color,
			Description: tag.Description,
		})
	}
	for _, problem := range snapshot.Problems {
		seed.Problems = append(seed.Problems, domain.SeedProblem{
			ID:              problem.ID,
			Code:            problem.Code,
			Latex:           problem.Latex,
			AnswerLatex:     problem.AnswerLatex,
			SolutionLatex:   problem.SolutionLatex,
			Type:            problem.Type,
			Difficulty:      problem.Difficulty,
			SubjectiveScore: problem.SubjectiveScore,
			Subject:         problem.Subject,
			Grade:           problem.Grade,
			Source:          problem.Source,
			TagIDs:          problem.TagIDs,
			ImageIDs:        problem.ImageIDs,
			Notes:           problem.Notes,
			CreatedAt:       problem.CreatedAt,
			UpdatedAt:       problem.UpdatedAt,
			Version:         problem.Version,
			IsDeleted:       problem.IsDeleted,
			Versions:        slices.Clone(problemVersions[problem.ID]),
		})
	}
	for _, imageAsset := range snapshot.Images {
		seed.Images = append(seed.Images, domain.SeedImage{
			ID:               imageAsset.ID,
			Filename:         imageAsset.Filename,
			MIME:             imageAsset.MIME,
			Size:             imageAsset.Size,
			Width:            imageAsset.Width,
			Height:           imageAsset.Height,
			TagIDs:           imageAsset.TagIDs,
			LinkedProblemIDs: imageAsset.LinkedProblemIDs,
			Description:      imageAsset.Description,
			CreatedAt:        imageAsset.CreatedAt,
			UpdatedAt:        imageAsset.UpdatedAt,
			IsDeleted:        imageAsset.IsDeleted,
		})
	}
	for _, paper := range snapshot.Papers {
		seed.Papers = append(seed.Papers, domain.SeedPaper{
			ID:           paper.ID,
			Title:        paper.Title,
			Subtitle:     paper.Subtitle,
			SchoolName:   paper.SchoolName,
			ExamName:     paper.ExamName,
			Subject:      paper.Subject,
			Duration:     paper.Duration,
			TotalScore:   paper.TotalScore,
			Description:  paper.Description,
			Status:       paper.Status,
			Instructions: paper.Instructions,
			FooterText:   paper.FooterText,
			CreatedAt:    paper.CreatedAt,
			UpdatedAt:    paper.UpdatedAt,
			Layout:       paper.Layout,
			Items:        paper.Items,
		})
	}
	for _, job := range snapshot.Exports {
		seed.ExportJobs = append(seed.ExportJobs, domain.SeedExportJob{
			ID:           job.ID,
			PaperID:      job.PaperID,
			PaperTitle:   job.PaperTitle,
			Format:       job.Format,
			Variant:      job.Variant,
			Status:       job.Status,
			Progress:     job.Progress,
			ErrorMessage: job.ErrorMessage,
			CreatedAt:    job.CreatedAt,
			CompletedAt:  job.CompletedAt,
		})
	}
	for _, entry := range searchHistory {
		seed.SearchHistory = append(seed.SearchHistory, domain.SeedSearchHistory{
			ID:          entry.ID,
			Query:       entry.Query,
			Filters:     entry.Filters,
			ResultCount: entry.ResultCount,
			CreatedAt:   entry.CreatedAt,
		})
	}
	for _, entry := range savedSearches {
		seed.SavedSearches = append(seed.SavedSearches, domain.SeedSavedSearchEntry{
			ID:        entry.ID,
			Name:      entry.Name,
			Query:     entry.Query,
			Filters:   entry.Filters,
			CreatedAt: entry.CreatedAt,
		})
	}
	return seed, nil
}

func (r *Repository) ImportAll(ctx context.Context, seed domain.DemoSeed) error {
	return r.ResetDemoData(ctx, seed)
}

func (r *Repository) loadProblemVersionsForExport(ctx context.Context) (map[string][]domain.SeedProblemVersion, error) {
	rows, err := r.db.Query(ctx, `SELECT id, problem_id, version, snapshot, created_at FROM problem_versions ORDER BY problem_id ASC, version ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make(map[string][]domain.SeedProblemVersion)
	for rows.Next() {
		var (
			id        string
			problemID string
			version   int
			raw       []byte
			createdAt time.Time
		)
		if err := rows.Scan(&id, &problemID, &version, &raw, &createdAt); err != nil {
			return nil, err
		}

		snapshot, err := decodeVersionSnapshot(raw)
		if err != nil {
			return nil, fmt.Errorf("decode problem version %s: %w", id, err)
		}
		items[problemID] = append(items[problemID], domain.SeedProblemVersion{
			Version:   version,
			CreatedAt: createdAt,
			Snapshot:  snapshot,
		})
	}
	return items, rows.Err()
}

func (r *Repository) listAllSearchHistory(ctx context.Context) ([]domain.SearchHistoryEntry, error) {
	rows, err := r.db.Query(ctx, `SELECT id, query, filters, result_count, created_at FROM search_history ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.SearchHistoryEntry, 0)
	for rows.Next() {
		var (
			item domain.SearchHistoryEntry
			raw  []byte
		)
		if err := rows.Scan(&item.ID, &item.Query, &raw, &item.ResultCount, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.Filters = map[string]any{}
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &item.Filters); err != nil {
				return nil, err
			}
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) listAllSavedSearches(ctx context.Context) ([]domain.SavedSearchEntry, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, query, filters, created_at FROM saved_searches ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.SavedSearchEntry, 0)
	for rows.Next() {
		var (
			item domain.SavedSearchEntry
			raw  []byte
		)
		if err := rows.Scan(&item.ID, &item.Name, &item.Query, &raw, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.Filters = map[string]any{}
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &item.Filters); err != nil {
				return nil, err
			}
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func decodeVersionSnapshot(raw []byte) (map[string]any, error) {
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var snapshot map[string]any
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}
