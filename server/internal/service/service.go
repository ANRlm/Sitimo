package service

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"mathlib/server/internal/config"
	exportsvc "mathlib/server/internal/export"
	"mathlib/server/internal/store"
	"mathlib/server/internal/worker"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
)

type Service struct {
	repo        *store.Repository
	cfg         config.Config
	logger      zerolog.Logger
	validate    *validator.Validate
	broadcaster *worker.Broadcaster
	exporter    *exportsvc.Manager
}

func New(cfg config.Config, repo *store.Repository, logger zerolog.Logger, broadcaster *worker.Broadcaster, exporter *exportsvc.Manager) *Service {
	return &Service{
		repo:        repo,
		cfg:         cfg,
		logger:      logger,
		validate:    validator.New(validator.WithRequiredStructEnabled()),
		broadcaster: broadcaster,
		exporter:    exporter,
	}
}

func (s *Service) Repository() *store.Repository {
	return s.repo
}

func (s *Service) Broadcaster() *worker.Broadcaster {
	return s.broadcaster
}

func (s *Service) SeedDemoData(ctx context.Context) error {
	file, err := os.ReadFile(demoSeedPath())
	if err != nil {
		return err
	}
	var seed SeedEnvelope
	if err := json.Unmarshal(file, &seed); err != nil {
		return err
	}
	return s.repo.ResetDemoData(ctx, seed.ToDomain())
}

func (s *Service) LoadDemoSeed(ctx context.Context) (SeedEnvelope, error) {
	file, err := os.ReadFile(demoSeedPath())
	if err != nil {
		return SeedEnvelope{}, err
	}
	var seed SeedEnvelope
	if err := json.Unmarshal(file, &seed); err != nil {
		return SeedEnvelope{}, err
	}
	return seed, nil
}

func (s *Service) LoadDemoData(ctx context.Context) (map[string]int, error) {
	file, err := os.ReadFile(demoSeedPath())
	if err != nil {
		return nil, err
	}
	var seed SeedEnvelope
	if err := json.Unmarshal(file, &seed); err != nil {
		return nil, err
	}
	return s.repo.LoadDemoData(ctx, seed.ToDomain())
}

func (s *Service) ClearDemoData(ctx context.Context) (map[string]int, error) {
	return s.repo.ClearDemoData(ctx)
}

func (s *Service) GetDemoDataStatus(ctx context.Context) (bool, map[string]int, error) {
	return s.repo.GetDemoDataStatus(ctx)
}

func demoSeedPath() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join("testdata", "demo-data.json")
	}
	return filepath.Join(filepath.Dir(filename), "..", "..", "testdata", "demo-data.json")
}
