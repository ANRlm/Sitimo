package service

import (
	"context"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/store"
)

func (s *Service) ListExports(ctx context.Context, params ExportListParams) (domain.Paginated[domain.ExportJob], error) {
	return s.repo.ListExportJobs(ctx, store.ExportListOptions{
		Status:   params.Status,
		Page:     params.Page,
		PageSize: params.PageSize,
	})
}

func (s *Service) CreateExport(ctx context.Context, input domain.ExportCreateInput) (*domain.ExportJob, error) {
	paper, err := s.repo.GetPaperDetail(ctx, input.PaperID, true)
	if err != nil {
		return nil, err
	}
	job, err := s.repo.CreateExportJobRecord(ctx, input, paper.Title)
	if err != nil {
		return nil, err
	}
	s.exporter.Enqueue(job.ID)
	return job, nil
}
