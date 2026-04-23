package service

import (
	"context"

	"mathlib/server/internal/domain"
)

func (s *Service) MetaStats(ctx context.Context) (domain.MetaStats, error) {
	return s.repo.MetaStats(ctx)
}

func (s *Service) RecentProblems(ctx context.Context, limit int) ([]domain.Problem, error) {
	items, err := s.ListProblems(ctx, ProblemListParams{Page: 1, PageSize: limit})
	if err != nil {
		return nil, err
	}
	return items.Items, nil
}

func (s *Service) RecentExports(ctx context.Context, limit int) ([]domain.ExportJob, error) {
	items, err := s.ListExports(ctx, ExportListParams{Page: 1, PageSize: limit})
	if err != nil {
		return nil, err
	}
	return items.Items, nil
}
