package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/service"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleCreateExport(w http.ResponseWriter, r *http.Request) {
	var payload domain.ExportCreateInput
	if !decodeAndValidate(w, r, &payload) {
		return
	}
	ctx, cancel := requestContext(r)
	defer cancel()
	item, err := s.svc.CreateExport(ctx, payload)
	if err != nil {
		respondError(w, http.StatusBadRequest, "create_export_failed", err)
		return
	}
	respondJSON(w, http.StatusCreated, item)
}

func (s *Server) handleListExports(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	items, err := s.svc.ListExports(ctx, service.ExportListParams{
		Status:   r.URL.Query().Get("status"),
		Page:     parseInt(r.URL.Query().Get("page"), 1),
		PageSize: parseInt(r.URL.Query().Get("pageSize"), 24),
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "list_exports_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (s *Server) handleGetExport(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	item, err := s.svc.Repository().GetExportJob(ctx, chi.URLParam(r, "id"))
	if handleNotFound(w, err) {
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "get_export_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (s *Server) handleDeleteExport(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	id := chi.URLParam(r, "id")
	job, err := s.svc.Repository().GetExportJob(ctx, id)
	if handleNotFound(w, err) {
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "load_export_failed", err)
		return
	}

	switch job.Status {
	case domain.ExportStatusProcessing:
		if err := s.svc.Repository().RequestExportCancellation(ctx, id); err != nil {
			respondError(w, http.StatusInternalServerError, "cancel_export_failed", err)
			return
		}
		if err := s.svc.Repository().NotifyExportJobChanged(ctx, id); err != nil {
			s.logger.Warn().Err(err).Str("job_id", id).Msg("failed to notify cancellation request")
		}
	case domain.ExportStatusPending, domain.ExportStatusDone, domain.ExportStatusFailed:
		if err := s.svc.Repository().DeleteExportJob(ctx, id); err != nil {
			respondError(w, http.StatusInternalServerError, "delete_export_failed", err)
			return
		}
	default:
		if err := s.svc.Repository().DeleteExportJob(ctx, id); err != nil {
			respondError(w, http.StatusInternalServerError, "delete_export_failed", err)
			return
		}
	}

	respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleDownloadExport(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	job, err := s.svc.Repository().GetExportJob(ctx, chi.URLParam(r, "id"))
	if handleNotFound(w, err) {
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "download_export_failed", err)
		return
	}

	path, err := s.svc.Repository().GetExportDownloadPath(ctx, job.ID)
	if handleNotFound(w, err) {
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "download_export_failed", err)
		return
	}

	w.Header().Set("Content-Disposition", exportAttachmentDisposition(exportDownloadFilename(job)))
	w.Header().Set("Content-Type", exportDownloadContentType(job.Format))
	http.ServeFile(w, r, filepath.Join(s.svc.Repository().StorageRoot(), path))
}

func exportDownloadFilename(job *domain.ExportJob) string {
	title := strings.TrimSpace(job.PaperTitle)
	if title == "" {
		title = "mathlib-export"
	}

	return fmt.Sprintf("%s-%s%s", title, exportVariantLabel(job.Variant), exportFormatExt(job.Format))
}

func exportVariantLabel(variant domain.ExportVariant) string {
	switch variant {
	case domain.ExportVariantAnswer:
		return "answer"
	case domain.ExportVariantBoth:
		return "both"
	default:
		return "student"
	}
}

func exportFormatExt(format domain.ExportFormat) string {
	switch format {
	case domain.ExportFormatPDF:
		return ".pdf"
	case domain.ExportFormatLatex:
		return ".zip"
	default:
		return ""
	}
}

func exportDownloadContentType(format domain.ExportFormat) string {
	switch format {
	case domain.ExportFormatPDF:
		return "application/pdf"
	case domain.ExportFormatLatex:
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}

func exportAttachmentDisposition(filename string) string {
	fallback := sanitizeASCIIFilename(filename)
	return fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, fallback, url.PathEscape(filename))
}

func sanitizeASCIIFilename(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "mathlib-export"
	}

	var builder strings.Builder
	builder.Grow(len(value))
	lastWasDash := false

	for _, char := range value {
		switch {
		case char >= 'a' && char <= 'z', char >= 'A' && char <= 'Z', char >= '0' && char <= '9', char == '.', char == '_':
			builder.WriteRune(char)
			lastWasDash = false
		case char == '-':
			if lastWasDash {
				continue
			}
			builder.WriteRune(char)
			lastWasDash = true
		default:
			if !lastWasDash {
				builder.WriteByte('-')
				lastWasDash = true
			}
		}
	}

	result := strings.Trim(builder.String(), "-._")
	if result == "" {
		return "mathlib-export"
	}
	return result
}

func (s *Server) handleExportStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx := r.Context()
	initial, err := s.svc.ListExports(ctx, service.ExportListParams{Page: 1, PageSize: 100})
	if err == nil {
		for _, item := range initial.Items {
			if item.Status == domain.ExportStatusPending || item.Status == domain.ExportStatusProcessing {
				_ = streamJSON(ctx, w, item)
			}
		}
	}

	ch := s.svc.Broadcaster().Subscribe(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case raw := <-ch:
			if len(raw) == 0 {
				continue
			}
			if _, err := w.Write([]byte("data: ")); err != nil {
				return
			}
			if _, err := w.Write(raw); err != nil {
				return
			}
			if _, err := w.Write([]byte("\n\n")); err != nil {
				return
			}
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		case <-time.After(15 * time.Second):
			if _, err := w.Write([]byte(": ping\n\n")); err != nil {
				return
			}
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}
