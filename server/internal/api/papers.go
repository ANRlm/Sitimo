package api

import (
	"net/http"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/service"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleListPapers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	items, err := s.svc.ListPapers(ctx, service.PaperListParams{
		Keyword:  r.URL.Query().Get("keyword"),
		Page:     parseInt(r.URL.Query().Get("page"), 1),
		PageSize: parseInt(r.URL.Query().Get("pageSize"), 24),
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "list_papers_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (s *Server) handleGetPaper(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	item, err := s.svc.Repository().GetPaperDetail(ctx, chi.URLParam(r, "id"), true)
	if handleNotFound(w, err) {
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "get_paper_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (s *Server) handleCreatePaper(w http.ResponseWriter, r *http.Request) {
	var payload domain.PaperWriteInput
	if !decodeAndValidate(w, r, &payload) {
		return
	}
	ctx, cancel := requestContext(r)
	defer cancel()
	item, err := s.svc.CreatePaper(ctx, payload)
	if err != nil {
		respondError(w, http.StatusBadRequest, "create_paper_failed", err)
		return
	}
	respondJSON(w, http.StatusCreated, item)
}

func (s *Server) handleUpdatePaper(w http.ResponseWriter, r *http.Request) {
	var payload domain.PaperWriteInput
	if !decodeAndValidate(w, r, &payload) {
		return
	}
	ctx, cancel := requestContext(r)
	defer cancel()
	item, err := s.svc.UpdatePaper(ctx, chi.URLParam(r, "id"), payload)
	if err != nil {
		respondError(w, http.StatusBadRequest, "update_paper_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (s *Server) handleUpdatePaperItems(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Items []domain.PaperItem `json:"items"`
	}
	if err := decodeJSON(r, &payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_json", err)
		return
	}
	ctx, cancel := requestContext(r)
	defer cancel()
	existing, err := s.svc.Repository().GetPaperDetail(ctx, chi.URLParam(r, "id"), true)
	if handleNotFound(w, err) {
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "load_paper_failed", err)
		return
	}
	input := domain.PaperWriteInput{
		Title:        existing.Title,
		Subtitle:     existing.Subtitle,
		SchoolName:   existing.SchoolName,
		ExamName:     existing.ExamName,
		Subject:      existing.Subject,
		Duration:     existing.Duration,
		TotalScore:   existing.TotalScore,
		Description:  existing.Description,
		Status:       existing.Status,
		Instructions: existing.Instructions,
		FooterText:   existing.FooterText,
		Items:        payload.Items,
		Layout:       existing.Layout,
	}
	item, err := s.svc.UpdatePaper(ctx, chi.URLParam(r, "id"), input)
	if err != nil {
		respondError(w, http.StatusBadRequest, "update_paper_items_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (s *Server) handleDeletePaper(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	if err := s.svc.Repository().DeletePaper(ctx, chi.URLParam(r, "id")); err != nil {
		respondError(w, http.StatusInternalServerError, "delete_paper_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleDuplicatePaper(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	item, err := s.svc.Repository().DuplicatePaper(ctx, chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "duplicate_paper_failed", err)
		return
	}
	respondJSON(w, http.StatusCreated, item)
}
