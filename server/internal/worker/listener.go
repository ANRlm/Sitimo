package worker

import (
	"context"
	"encoding/json"
	"time"

	"mathlib/server/internal/store"

	"github.com/rs/zerolog"
)

type Listener struct {
	repo        *store.Repository
	broadcaster *Broadcaster
	logger      zerolog.Logger
	channel     string
}

func NewListener(repo *store.Repository, broadcaster *Broadcaster, logger zerolog.Logger) *Listener {
	return &Listener{
		repo:        repo,
		broadcaster: broadcaster,
		logger:      logger,
		channel:     "export_channel",
	}
}

func (l *Listener) Start(ctx context.Context) {
	go l.run(ctx)
}

func (l *Listener) run(ctx context.Context) {
	for ctx.Err() == nil {
		conn, err := l.repo.DB().Acquire(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			l.logger.Error().Err(err).Msg("failed to acquire listener connection")
			sleepContext(ctx, time.Second)
			continue
		}

		if _, err := conn.Exec(ctx, "LISTEN "+l.channel); err != nil {
			conn.Release()
			if ctx.Err() != nil {
				return
			}
			l.logger.Error().Err(err).Msg("failed to listen for export notifications")
			sleepContext(ctx, time.Second)
			continue
		}

		waitConn := conn.Conn()
		for ctx.Err() == nil {
			notification, err := waitConn.WaitForNotification(ctx)
			if err != nil {
				break
			}
			l.publishPayload(ctx, []byte(notification.Payload))
		}

		conn.Release()
		sleepContext(ctx, 250*time.Millisecond)
	}
}

func (l *Listener) publishPayload(ctx context.Context, raw []byte) {
	var payload struct {
		JobID string `json:"jobId"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil || payload.JobID == "" {
		l.broadcaster.PublishRaw(raw)
		return
	}

	job, err := l.repo.GetExportJob(ctx, payload.JobID)
	if err != nil {
		l.logger.Warn().Err(err).Str("job_id", payload.JobID).Msg("failed to load export job from notification")
		l.broadcaster.PublishRaw(raw)
		return
	}
	l.broadcaster.Publish(job)
}

func sleepContext(ctx context.Context, duration time.Duration) {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}
