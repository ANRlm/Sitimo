package service

import (
	"context"
	"fmt"
	"strings"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/parser"
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
	hasFiles := len(input.Files) > 0
	hasLatex := input.Latex != ""

	// Structural parser mode: use the new file-based parser.
	if hasFiles {
		result := parser.BuildImportPreview(input.Files, input.Defaults)
		// Ensure backward-compatible fields for downstream consumers.
		for i := range result.Parsed {
			if result.Parsed[i].Difficulty == "" {
				result.Parsed[i].Difficulty = domain.DifficultyMedium
			}
		}
		if hasLatex {
			result.Warnings = append(result.Warnings, "同时提供了 LaTeX 源码和文件，仅使用文件进行结构解析，忽略了 LaTeX 源码。")
		}
		if len(result.Parsed) == 0 && len(result.Errors) == 0 {
			result.Warnings = append(result.Warnings, "未解析到任何题目，请检查文件内容。")
		}
		return result
	}

	// Neither files nor LaTeX content provided.
	if !hasLatex {
		return domain.ImportPreviewResponse{
			Parsed:   make([]domain.ImportPreviewDraft, 0),
			Errors:   []map[string]any{{"message": "请提供 LaTeX 源码或上传 .tex 文件"}},
			Warnings: make([]string, 0),
		}
	}

	// Legacy delimiter-based mode.
	startToken := strings.TrimSpace(input.SeparatorStart)
	endToken := strings.TrimSpace(input.SeparatorEnd)
	if startToken == "" {
		startToken = `\begin{problem}`
	}
	// Single-delimiter mode when endToken is empty: each chunk between
	// startTokens is treated as one problem (useful for \item-based .tex files).
	singleDelimiter := endToken == ""

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
		// Skip LaTeX document preamble chunks containing \documentclass.
		if strings.Contains(chunk, `\documentclass`) {
			continue
		}

		index++

		var body string
		if singleDelimiter {
			body = chunk
		} else {
			position := strings.Index(chunk, endToken)
			if position < 0 {
				message := "缺少结束分隔符，已终止该题解析。"
				result.Parsed = append(result.Parsed, domain.ImportPreviewDraft{
					ID:            newImportID(index),
					Title:         "题目 #" + itoa(index),
					Latex:         chunk,
					Difficulty:    domain.DifficultyMedium,
					Status:        "error",
					Error:         &message,
					Warnings:      search.LatexWarnings(chunk),
					TagNames:      []string{},
					InferredType:  domain.ProblemTypeSolve,
					NeedsReview:   false,
				})
				result.Errors = append(result.Errors, map[string]any{"index": index, "message": message})
				continue
			}
			body = strings.TrimSpace(chunk[:position])
		}

		result.Parsed = append(result.Parsed, domain.ImportPreviewDraft{
			ID:            newImportID(index),
			Title:         "题目 #" + itoa(index),
			Latex:         body,
			Difficulty:    parseDefaultDifficulty(input.Defaults),
			Status:        "success",
			Warnings:      search.LatexWarnings(body),
			Subject:       mapStringPtr(input.Defaults, "subject"),
			Grade:         mapStringPtr(input.Defaults, "grade"),
			Source:        mapStringPtr(input.Defaults, "source"),
			TagNames:      parseTagNames(input.Defaults["tagNames"]),
			InferredType:  domain.ProblemTypeSolve,
			NeedsReview:   false,
		})
	}

	if len(result.Parsed) == 0 {
		result.Warnings = append(result.Warnings, "未解析到任何题目，请检查分隔符。")
	}
	return result
}

func (s *Service) CommitBatchImport(ctx context.Context, drafts []domain.ImportPreviewDraft) ([]domain.ProblemDetail, error) {
	tags, err := s.repo.ListTags(ctx)
	if err != nil {
		return nil, err
	}
	tagIDByName := map[string]string{}
	for _, tag := range tags {
		tagIDByName[tag.Name] = tag.ID
	}

	validDrafts := make([]domain.ProblemWriteInput, 0, len(drafts))
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
		problemType := draft.InferredType
		if problemType == "" {
			problemType = domain.ProblemTypeSolve
		}

		validDrafts = append(validDrafts, domain.ProblemWriteInput{
			Latex:         draft.Latex,
			AnswerLatex:   draft.AnswerLatex,
			SolutionLatex: draft.SolutionLatex,
			Type:          problemType,
			Difficulty:    draft.Difficulty,
			Subject:       draft.Subject,
			Grade:         draft.Grade,
			Source:        draft.Source,
			TagIDs:        tagIDs,
		})

		if draft.NeedsReview {
			s.logger.Warn().
				Str("component", "import").
				Str("draft_id", draft.ID).
				Str("title", draft.Title).
				Msg(fmt.Sprintf("题目 %s 类型为混合型或未识别，建议人工审核", draft.Title))
		}
	}

	if len(validDrafts) == 0 {
		return nil, nil
	}

	tx, err := s.repo.DB().Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	ids, err := s.repo.CreateProblemsTx(ctx, tx, validDrafts)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	results := make([]domain.ProblemDetail, 0, len(ids))
	for _, id := range ids {
		detail, err := s.repo.GetProblemDetail(ctx, id, true)
		if err != nil {
			return nil, err
		}
		results = append(results, *detail)
	}
	return results, nil
}
