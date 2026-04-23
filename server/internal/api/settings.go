package api

import (
	"net/http"

	"mathlib/server/internal/domain"
)

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	settings, err := s.svc.Repository().GetSettings(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "get_settings_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

func (s *Server) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var payload domain.SettingsPayload
	if err := decodeJSON(r, &payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_json", err)
		return
	}
	ctx, cancel := requestContext(r)
	defer cancel()
	if err := s.svc.Repository().UpsertSettings(ctx, payload); err != nil {
		respondError(w, http.StatusInternalServerError, "update_settings_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleResetDemoData(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	if err := s.svc.SeedDemoData(ctx); err != nil {
		respondError(w, http.StatusInternalServerError, "reset_demo_data_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleExportAll(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	seed, err := s.svc.Repository().ExportAll(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "export_all_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, seed)
}

func (s *Server) handleImportAll(w http.ResponseWriter, r *http.Request) {
	var payload domain.DemoSeed
	if err := decodeJSON(r, &payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_json", err)
		return
	}
	ctx, cancel := requestContext(r)
	defer cancel()
	if err := s.svc.Repository().ImportAll(ctx, payload); err != nil {
		respondError(w, http.StatusInternalServerError, "import_all_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleSweepOrphans(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	result, err := s.svc.SweepOrphanImages(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "sweep_orphans_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (s *Server) handleLoadDemoData(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	stats, err := s.svc.LoadDemoData(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "load_demo_data_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, stats)
}

func (s *Server) handleClearDemoData(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	stats, err := s.svc.ClearDemoData(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "clear_demo_data_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, stats)
}

func (s *Server) handleDemoDataStatus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	loaded, stats, err := s.svc.GetDemoDataStatus(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "get_demo_data_status_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"loaded": loaded,
		"stats":  stats,
	})
}
