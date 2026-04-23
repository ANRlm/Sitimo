package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mathlib/server/internal/config"
	"mathlib/server/internal/domain"
	"mathlib/server/internal/store/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db            *pgxpool.Pool
	publicBaseURL string
	storageRoot   string
	queries       *sqlc.Queries
}

func NewRepository(ctx context.Context, cfg config.Config) (*Repository, error) {
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	repo := &Repository{
		db:            pool,
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
		storageRoot:   cfg.StorageRoot,
		queries:       sqlc.New(pool),
	}

	if err := repo.ensureStorageLayout(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *Repository) Close() {
	if r.db != nil {
		r.db.Close()
	}
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

func (r *Repository) DB() *pgxpool.Pool {
	return r.db
}

func (r *Repository) Queries() *sqlc.Queries {
	return r.queries
}

func (r *Repository) StorageRoot() string {
	return r.storageRoot
}

func (r *Repository) ensureStorageLayout() error {
	paths := []string{
		filepath.Join(r.storageRoot, "original"),
		filepath.Join(r.storageRoot, "thumbnails"),
		filepath.Join(r.storageRoot, "derived"),
		filepath.Join(r.storageRoot, "derived", "exports"),
	}

	for _, path := range paths {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func marshalJSON(value any) ([]byte, error) {
	if value == nil {
		return []byte(`{}`), nil
	}
	return json.Marshal(value)
}

func unmarshalJSON[T any](raw []byte, target *T) error {
	if len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("unmarshal json: %w", err)
	}
	return nil
}

func copySettings(input domain.SettingsPayload) domain.SettingsPayload {
	output := make(domain.SettingsPayload, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}
