package api

import (
	"net/http"
)

func (s *Server) handleMetaGrades(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	values, err := s.svc.Repository().ListProblemGrades(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "meta_grades_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, values)
}

func (s *Server) handleMetaStats(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	stats, err := s.svc.MetaStats(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "meta_stats_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, stats)
}

func (s *Server) handleRecentProblems(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	items, err := s.svc.RecentProblems(ctx, parseInt(r.URL.Query().Get("limit"), 5))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "recent_problems_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (s *Server) handleRecentExports(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	items, err := s.svc.RecentExports(ctx, parseInt(r.URL.Query().Get("limit"), 5))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "recent_exports_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, items)
}
