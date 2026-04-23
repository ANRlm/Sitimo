package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"mathlib/server/internal/api"
	"mathlib/server/internal/config"
	exportsvc "mathlib/server/internal/export"
	"mathlib/server/internal/service"
	"mathlib/server/internal/store"
	"mathlib/server/internal/worker"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	if err := cfg.Validate(); err != nil {
		logger.Fatal().Err(err).Msg("invalid configuration")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	repo, err := store.NewRepository(ctx, cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create repository")
	}
	defer repo.Close()

	broadcaster := worker.NewBroadcaster()
	listener := worker.NewListener(repo, broadcaster, logger)
	listener.Start(ctx)
	exporter := exportsvc.NewManager(repo, broadcaster, logger)
	exporter.Start(ctx)

	svc := service.New(cfg, repo, logger, broadcaster, exporter)

	command := "serve"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "seed":
		if err := svc.SeedDemoData(ctx); err != nil {
			logger.Fatal().Err(err).Msg("failed to seed demo data")
		}
		logger.Info().Msg("demo data seeded")
	case "serve":
		if result, err := svc.SweepOrphanImages(ctx); err != nil {
			logger.Warn().Err(err).Msg("failed to sweep orphan images on startup")
		} else if result.Deleted > 0 {
			logger.Info().Int("deleted", result.Deleted).Int64("bytes_freed", result.BytesFreed).Msg("swept orphan images")
		}
		handler := api.New(cfg, svc, logger)
		server := &http.Server{
			Addr:    ":" + cfg.Port,
			Handler: handler,
		}

		go func() {
			<-ctx.Done()
			_ = server.Shutdown(context.Background())
		}()

		logger.Info().Str("addr", server.Addr).Msg("mathlib server listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("server stopped unexpectedly")
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		os.Exit(1)
	}
}
