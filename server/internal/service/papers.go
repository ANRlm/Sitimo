package service

import (
	"context"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/store"
)

func (s *Service) ListPapers(ctx context.Context, params PaperListParams) (domain.Paginated[domain.PaperDetail], error) {
	return s.repo.ListPapers(ctx, store.PaperListOptions{
		Keyword:  params.Keyword,
		Page:     params.Page,
		PageSize: params.PageSize,
	})
}

func (s *Service) CreatePaper(ctx context.Context, input domain.PaperWriteInput) (*domain.PaperDetail, error) {
	return s.repo.CreatePaper(ctx, input)
}

func (s *Service) UpdatePaper(ctx context.Context, id string, input domain.PaperWriteInput) (*domain.PaperDetail, error) {
	return s.repo.UpdatePaper(ctx, id, input)
}
