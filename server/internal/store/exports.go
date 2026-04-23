package store

import (
	"context"
	"time"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/store/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *Repository) CreateExportJobRecord(ctx context.Context, input domain.ExportCreateInput, paperTitle string) (*domain.ExportJob, error) {
	id := newID()
	now := time.Now().UTC()
	record, err := r.queries.InsertExportJob(ctx, sqlc.InsertExportJobParams{
		ID:         id,
		PaperID:    input.PaperID,
		PaperTitle: paperTitle,
		Format:     string(input.Format),
		Variant:    string(input.Variant),
		Status:     string(domain.ExportStatusPending),
		Progress:   0,
		CreatedAt:  pgTimestamptzFromTime(now),
	})
	if err != nil {
		return nil, err
	}
	item := r.exportJobFromRecord(record)
	return &item, nil
}

func (r *Repository) UpdateExportJobState(ctx context.Context, id string, status domain.ExportStatus, progress int, downloadPath *string, errorMessage *string) error {
	current, err := r.queries.GetExportJobByID(ctx, id)
	if err != nil {
		return err
	}

	startedAt := current.StartedAt
	completedAt := current.CompletedAt
	download := current.DownloadPath
	errMessage := current.ErrorMessage
	cancelRequestedAt := current.CancelRequestedAt

	if status == domain.ExportStatusProcessing {
		if !startedAt.Valid {
			startedAt = pgTimestamptzFromTime(time.Now().UTC())
		}
	}
	if status == domain.ExportStatusDone || status == domain.ExportStatusFailed {
		completedAt = pgTimestamptzFromTime(time.Now().UTC())
		cancelRequestedAt = pgtype.Timestamptz{}
	}
	if downloadPath != nil {
		download = pgTextFromPtr(downloadPath)
	}
	if errorMessage != nil {
		errMessage = pgTextFromPtr(errorMessage)
	}

	return r.queries.UpdateExportJobStatus(ctx, sqlc.UpdateExportJobStatusParams{
		ID:                id,
		Status:            string(status),
		Progress:          int32(progress),
		StartedAt:         startedAt,
		CompletedAt:       completedAt,
		ErrorMessage:      errMessage,
		DownloadPath:      download,
		CancelRequestedAt: cancelRequestedAt,
	})
}

func (r *Repository) RequestExportCancellation(ctx context.Context, id string) error {
	return r.queries.RequestExportCancellation(ctx, id)
}

func (r *Repository) DeleteExportJob(ctx context.Context, id string) error {
	return r.queries.DeleteExportJob(ctx, id)
}

func (r *Repository) GetExportDownloadPath(ctx context.Context, id string) (string, error) {
	record, err := r.queries.GetExportJobByID(ctx, id)
	if err != nil {
		return "", err
	}
	if !record.DownloadPath.Valid {
		return "", pgx.ErrNoRows
	}
	return record.DownloadPath.String, nil
}
