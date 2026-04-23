package api

import (
	"encoding/json"
	"net/http"

	"mathlib/server/internal/domain"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	conditions := make([]domain.SearchCondition, 0)
	if raw := r.URL.Query().Get("conditions"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &conditions); err != nil {
			respondError(w, http.StatusBadRequest, "invalid_conditions", err)
			return
		}
	}
	items, err := s.svc.Search(ctx, r.URL.Query().Get("keyword"), r.URL.Query().Get("formula"), conditions)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "search_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (s *Server) handleSearchHistory(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	items, err := s.svc.Repository().ListSearchHistory(ctx, parseInt(r.URL.Query().Get("limit"), 20))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "list_search_history_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (s *Server) handleDeleteSearchHistory(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	if err := s.svc.Repository().DeleteSearchHistory(ctx, chi.URLParam(r, "id")); err != nil {
		respondError(w, http.StatusInternalServerError, "delete_search_history_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleSavedSearches(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	items, err := s.svc.Repository().ListSavedSearches(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "list_saved_searches_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (s *Server) handleCreateSavedSearch(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name    string         `json:"name"`
		Query   string         `json:"query"`
		Filters map[string]any `json:"filters"`
	}
	if err := decodeJSON(r, &payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_json", err)
		return
	}
	ctx, cancel := requestContext(r)
	defer cancel()
	item, err := s.svc.Repository().CreateSavedSearch(ctx, payload.Name, payload.Query, payload.Filters)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "create_saved_search_failed", err)
		return
	}
	respondJSON(w, http.StatusCreated, item)
}

func (s *Server) handleDeleteSavedSearch(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	if err := s.svc.Repository().DeleteSavedSearch(ctx, chi.URLParam(r, "id")); err != nil {
		respondError(w, http.StatusInternalServerError, "delete_saved_search_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
