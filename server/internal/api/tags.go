package api

import (
	"net/http"

	"mathlib/server/internal/domain"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleListTags(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	items, err := s.svc.Repository().ListTags(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "list_tags_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (s *Server) handleCreateTag(w http.ResponseWriter, r *http.Request) {
	var payload domain.Tag
	if !decodeAndValidate(w, r, &payload) {
		return
	}
	ctx, cancel := requestContext(r)
	defer cancel()
	item, err := s.svc.Repository().CreateTag(ctx, payload)
	if err != nil {
		respondError(w, http.StatusBadRequest, "create_tag_failed", err)
		return
	}
	respondJSON(w, http.StatusCreated, item)
}

func (s *Server) handleUpdateTag(w http.ResponseWriter, r *http.Request) {
	var payload domain.Tag
	if !decodeAndValidate(w, r, &payload) {
		return
	}
	ctx, cancel := requestContext(r)
	defer cancel()
	item, err := s.svc.Repository().UpdateTag(ctx, chi.URLParam(r, "id"), payload)
	if err != nil {
		respondError(w, http.StatusBadRequest, "update_tag_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (s *Server) handleDeleteTag(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	if err := s.svc.Repository().DeleteTag(ctx, chi.URLParam(r, "id")); err != nil {
		respondError(w, http.StatusInternalServerError, "delete_tag_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleMergeTag(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		TargetID string `json:"targetId"`
	}
	if err := decodeJSON(r, &payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_json", err)
		return
	}
	ctx, cancel := requestContext(r)
	defer cancel()
	if err := s.svc.Repository().MergeTag(ctx, chi.URLParam(r, "id"), payload.TargetID); err != nil {
		respondError(w, http.StatusInternalServerError, "merge_tag_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
