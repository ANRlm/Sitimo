package store

import (
	"context"
	"errors"
	"time"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/store/sqlc"
)

func (r *Repository) CreateTag(ctx context.Context, input domain.Tag) (*domain.Tag, error) {
	id := input.ID
	if id == "" {
		id = newID()
	}
	now := time.Now().UTC()
	record, err := r.queries.InsertTag(ctx, sqlc.InsertTagParams{
		ID:          id,
		Name:        input.Name,
		Category:    string(input.Category),
		Color:       input.Color,
		Description: pgTextFromPtr(input.Description),
		CreatedAt:   pgTimestamptzFromTime(now),
		UpdatedAt:   pgTimestamptzFromTime(now),
	})
	if err != nil {
		return nil, err
	}

	count, err := r.queries.CountTagProblems(ctx, id)
	if err != nil {
		return nil, err
	}
	item := r.tagFromRecord(record, int(count))
	return &item, nil
}

func (r *Repository) UpdateTag(ctx context.Context, id string, input domain.Tag) (*domain.Tag, error) {
	record, err := r.queries.UpdateTag(ctx, sqlc.UpdateTagParams{
		ID:          id,
		Name:        input.Name,
		Category:    string(input.Category),
		Color:       input.Color,
		Description: pgTextFromPtr(input.Description),
		UpdatedAt:   pgTimestamptzFromTime(time.Now().UTC()),
	})
	if err != nil {
		return nil, err
	}
	count, err := r.queries.CountTagProblems(ctx, id)
	if err != nil {
		return nil, err
	}
	item := r.tagFromRecord(record, int(count))
	return &item, nil
}

func (r *Repository) DeleteTag(ctx context.Context, id string) error {
	return r.queries.DeleteTag(ctx, id)
}

func (r *Repository) MergeTag(ctx context.Context, sourceID, targetID string) error {
	if sourceID == targetID {
		return errors.New("cannot merge tag with itself")
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	queries := r.queries.WithTx(tx)

	if err := queries.UpdateProblemTagsForMerge(ctx, sqlc.UpdateProblemTagsForMergeParams{
		TagID:   targetID,
		TagID_2: sourceID,
	}); err != nil {
		return err
	}
	if err := queries.UpdateImageTagsForMerge(ctx, sqlc.UpdateImageTagsForMergeParams{
		TagID:   targetID,
		TagID_2: sourceID,
	}); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM problem_tags WHERE tag_id = $1`, sourceID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM image_tags WHERE tag_id = $1`, sourceID); err != nil {
		return err
	}
	if err := queries.DeleteTag(ctx, sourceID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
