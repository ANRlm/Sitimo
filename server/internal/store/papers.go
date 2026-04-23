package store

import (
	"context"
	"fmt"
	"time"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/store/sqlc"
)

func (r *Repository) CreatePaper(ctx context.Context, input domain.PaperWriteInput) (*domain.PaperDetail, error) {
	id := newID()
	return r.upsertPaper(ctx, id, input, true)
}

func (r *Repository) UpdatePaper(ctx context.Context, id string, input domain.PaperWriteInput) (*domain.PaperDetail, error) {
	return r.upsertPaper(ctx, id, input, false)
}

func (r *Repository) upsertPaper(ctx context.Context, id string, input domain.PaperWriteInput, isCreate bool) (*domain.PaperDetail, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	queries := r.queries.WithTx(tx)

	now := time.Now().UTC()
	totalScore := input.TotalScore
	if totalScore == nil {
		calculated := 0.0
		for _, item := range input.Items {
			calculated += item.Score
		}
		totalScore = &calculated
	}

	layoutBytes, err := marshalJSON(input.Layout)
	if err != nil {
		return nil, err
	}
	headerBytes, err := marshalJSON(map[string]any{})
	if err != nil {
		return nil, err
	}
	totalScoreValue, err := pgNumericFromPtr(totalScore)
	if err != nil {
		return nil, err
	}

	if isCreate {
		_, err = queries.InsertPaper(ctx, sqlc.InsertPaperParams{
			ID:           id,
			Title:        input.Title,
			Subtitle:     pgTextFromPtr(input.Subtitle),
			SchoolName:   pgTextFromPtr(input.SchoolName),
			ExamName:     pgTextFromPtr(input.ExamName),
			Subject:      pgTextFromPtr(input.Subject),
			DurationMin:  pgInt4FromPtr(input.Duration),
			TotalScore:   totalScoreValue,
			Description:  pgTextFromPtr(input.Description),
			Status:       string(input.Status),
			Instructions: pgTextFromPtr(input.Instructions),
			FooterText:   pgTextFromPtr(input.FooterText),
			HeaderJson:   headerBytes,
			LayoutJson:   layoutBytes,
			CreatedAt:    pgTimestamptzFromTime(now),
			UpdatedAt:    pgTimestamptzFromTime(now),
		})
	} else {
		_, err = queries.UpdatePaper(ctx, sqlc.UpdatePaperParams{
			ID:           id,
			Title:        input.Title,
			Subtitle:     pgTextFromPtr(input.Subtitle),
			SchoolName:   pgTextFromPtr(input.SchoolName),
			ExamName:     pgTextFromPtr(input.ExamName),
			Subject:      pgTextFromPtr(input.Subject),
			DurationMin:  pgInt4FromPtr(input.Duration),
			TotalScore:   totalScoreValue,
			Description:  pgTextFromPtr(input.Description),
			Status:       string(input.Status),
			Instructions: pgTextFromPtr(input.Instructions),
			FooterText:   pgTextFromPtr(input.FooterText),
			HeaderJson:   headerBytes,
			LayoutJson:   layoutBytes,
			UpdatedAt:    pgTimestamptzFromTime(now),
		})
	}
	if err != nil {
		return nil, err
	}

	if err := queries.DeletePaperItems(ctx, id); err != nil {
		return nil, err
	}
	for index, item := range input.Items {
		itemID := item.ID
		if itemID == "" {
			itemID = fmt.Sprintf("%s-item-%d", id, index+1)
		}
		scoreValue, err := pgNumericFromFloat(item.Score)
		if err != nil {
			return nil, err
		}
		if err := queries.InsertPaperItem(ctx, sqlc.InsertPaperItemParams{
			ID:            itemID,
			PaperID:       id,
			ProblemID:     item.ProblemID,
			OrderIndex:    int32(index),
			Score:         scoreValue,
			ImagePosition: pgTextFromString(item.ImagePosition),
			BlankLines:    int32(item.BlankLines),
		}); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetPaperDetail(ctx, id, true)
}

func (r *Repository) DeletePaper(ctx context.Context, id string) error {
	return r.queries.SoftDeletePaper(ctx, id)
}

func (r *Repository) DuplicatePaper(ctx context.Context, id string) (*domain.PaperDetail, error) {
	loaded, err := r.GetPaperDetail(ctx, id, true)
	if err != nil {
		return nil, err
	}
	title := loaded.Title + "（副本）"
	input := domain.PaperWriteInput{
		Title:        title,
		Subtitle:     loaded.Subtitle,
		SchoolName:   loaded.SchoolName,
		ExamName:     loaded.ExamName,
		Subject:      loaded.Subject,
		Duration:     loaded.Duration,
		TotalScore:   loaded.TotalScore,
		Description:  loaded.Description,
		Status:       loaded.Status,
		Instructions: loaded.Instructions,
		FooterText:   loaded.FooterText,
		Layout:       loaded.Layout,
		Items:        loaded.Items,
	}
	return r.CreatePaper(ctx, input)
}
