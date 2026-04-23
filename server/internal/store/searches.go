package store

import (
	"context"
	"encoding/json"

	"mathlib/server/internal/domain"
)

func (r *Repository) CreateSearchHistory(ctx context.Context, query string, filters map[string]any, resultCount int) error {
	raw, err := json.Marshal(filters)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `INSERT INTO search_history (id, query, filters, result_count, created_at) VALUES ($1, $2, $3, $4, now())`,
		newID(), query, raw, resultCount,
	)
	return err
}

func (r *Repository) ListSearchHistory(ctx context.Context, limit int) ([]domain.SearchHistoryEntry, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.Query(ctx, `SELECT id, query, filters, result_count, created_at FROM search_history ORDER BY created_at DESC LIMIT $1`, limit)
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

func (r *Repository) DeleteSearchHistory(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM search_history WHERE id = $1`, id)
	return err
}

func (r *Repository) CreateSavedSearch(ctx context.Context, name, query string, filters map[string]any) (*domain.SavedSearchEntry, error) {
	id := newID()
	raw, err := json.Marshal(filters)
	if err != nil {
		return nil, err
	}
	_, err = r.db.Exec(ctx, `INSERT INTO saved_searches (id, name, query, filters, created_at) VALUES ($1, $2, $3, $4, now())`,
		id, name, query, raw,
	)
	if err != nil {
		return nil, err
	}
	items, err := r.ListSavedSearches(ctx)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.ID == id {
			copyItem := item
			return &copyItem, nil
		}
	}
	return nil, nil
}

func (r *Repository) ListSavedSearches(ctx context.Context) ([]domain.SavedSearchEntry, error) {
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

func (r *Repository) DeleteSavedSearch(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM saved_searches WHERE id = $1`, id)
	return err
}
