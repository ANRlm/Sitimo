package service

import (
	"context"
	"strings"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/search"
	"mathlib/server/internal/store"
)

func (s *Service) ListProblems(ctx context.Context, params ProblemListParams) (domain.Paginated[domain.Problem], error) {
	return s.repo.ListProblems(ctx, store.ProblemListOptions{
		Keyword:        params.Keyword,
		Subject:        params.Subject,
		Grade:          params.Grade,
		Difficulty:     params.Difficulty,
		Type:           params.Type,
		TagIDs:         params.TagIDs,
		HasImage:       params.HasImage,
		ScoreMin:       params.ScoreMin,
		ScoreMax:       params.ScoreMax,
		SortBy:         params.SortBy,
		SortOrder:      params.SortOrder,
		Page:           params.Page,
		PageSize:       params.PageSize,
		IncludeDeleted: params.Deleted,
	})
}

func (s *Service) CreateProblem(ctx context.Context, input domain.ProblemWriteInput) (*domain.ProblemDetail, []string, error) {
	if err := s.validate.Struct(input); err != nil {
		return nil, nil, err
	}
	created, err := s.repo.CreateProblem(ctx, input)
	return created, search.LatexWarnings(input.Latex), err
}

func (s *Service) UpdateProblem(ctx context.Context, id string, input domain.ProblemWriteInput) (*domain.ProblemDetail, []string, error) {
	if err := s.validate.Struct(input); err != nil {
		return nil, nil, err
	}
	updated, err := s.repo.UpdateProblem(ctx, id, input)
	return updated, search.LatexWarnings(input.Latex), err
}

func (s *Service) PreviewBatchImport(input domain.ImportPreviewRequest) domain.ImportPreviewResponse {
	startToken := strings.TrimSpace(input.SeparatorStart)
	endToken := strings.TrimSpace(input.SeparatorEnd)
	if startToken == "" {
		startToken = `\begin{problem}`
	}
	if endToken == "" {
		endToken = `\end{problem}`
	}

	chunks := strings.Split(input.Latex, startToken)
	result := domain.ImportPreviewResponse{
		Parsed:   make([]domain.ImportPreviewDraft, 0),
		Errors:   make([]map[string]any, 0),
		Warnings: make([]string, 0),
	}

	index := 0
	for _, chunk := range chunks {
		chunk = strings.TrimSpace(chunk)
		if chunk == "" {
			continue
		}
		index++

		position := strings.Index(chunk, endToken)
		if position < 0 {
			message := "缺少结束分隔符，已终止该题解析。"
			result.Parsed = append(result.Parsed, domain.ImportPreviewDraft{
				ID:         newImportID(index),
				Title:      "题目 #" + itoa(index),
				Latex:      chunk,
				Difficulty: domain.DifficultyMedium,
				Status:     "error",
				Error:      &message,
				Warnings:   search.LatexWarnings(chunk),
				TagNames:   []string{},
			})
			result.Errors = append(result.Errors, map[string]any{"index": index, "message": message})
			continue
		}

		body := strings.TrimSpace(chunk[:position])
		result.Parsed = append(result.Parsed, domain.ImportPreviewDraft{
			ID:         newImportID(index),
			Title:      "题目 #" + itoa(index),
			Latex:      body,
			Difficulty: parseDefaultDifficulty(input.Defaults),
			Status:     "success",
			Warnings:   search.LatexWarnings(body),
			Subject:    mapStringPtr(input.Defaults, "subject"),
			Grade:      mapStringPtr(input.Defaults, "grade"),
			Source:     mapStringPtr(input.Defaults, "source"),
			TagNames:   parseTagNames(input.Defaults["tagNames"]),
		})
	}

	if len(result.Parsed) == 0 {
		result.Warnings = append(result.Warnings, "未解析到任何题目，请检查分隔符。")
	}
	return result
}

func (s *Service) CommitBatchImport(ctx context.Context, drafts []domain.ImportPreviewDraft) ([]domain.ProblemDetail, error) {
	results := make([]domain.ProblemDetail, 0, len(drafts))
	tags, err := s.repo.ListTags(ctx)
	if err != nil {
		return nil, err
	}
	tagIDByName := map[string]string{}
	for _, tag := range tags {
		tagIDByName[tag.Name] = tag.ID
	}

	for _, draft := range drafts {
		if draft.Status == "error" {
			continue
		}
		tagIDs := make([]string, 0, len(draft.TagNames))
		for _, tagName := range draft.TagNames {
			if id, ok := tagIDByName[tagName]; ok {
				tagIDs = append(tagIDs, id)
			}
		}
		created, _, err := s.CreateProblem(ctx, domain.ProblemWriteInput{
			Latex:      draft.Latex,
			Type:       domain.ProblemTypeSolve,
			Difficulty: draft.Difficulty,
			Subject:    draft.Subject,
			Grade:      draft.Grade,
			Source:     draft.Source,
			TagIDs:     tagIDs,
		})
		if err != nil {
			return nil, err
		}
		results = append(results, *created)
	}
	return results, nil
}
