package api

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/service"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleListImages(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	result, err := s.svc.ListImages(ctx, service.ImageListParams{
		Keyword:  r.URL.Query().Get("keyword"),
		TagIDs:   parseCommaList(r.URL.Query().Get("tagIds")),
		MIME:     r.URL.Query().Get("mime"),
		Page:     parseInt(r.URL.Query().Get("page"), 1),
		PageSize: parseInt(r.URL.Query().Get("pageSize"), 24),
		Deleted:  r.URL.Query().Get("deleted") == "true",
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "list_images_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (s *Server) handleGetImage(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	detail, err := s.svc.Repository().GetImageDetail(ctx, chi.URLParam(r, "id"), true)
	if handleNotFound(w, err) {
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "get_image_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"image":          detail.Image,
		"linkedProblems": detail.LinkedProblems,
		"tags":           detail.Tags,
	})
}

func (s *Server) handleUploadImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(25 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_multipart", err)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "file_required", err)
		return
	}
	defer file.Close()

	body, err := io.ReadAll(file)
	if err != nil {
		respondError(w, http.StatusBadRequest, "read_file_failed", err)
		return
	}
	description := strings.TrimSpace(r.FormValue("description"))
	var descriptionPtr *string
	if description != "" {
		descriptionPtr = &description
	}
	tagIDs := parseCommaList(r.FormValue("tagIds"))

	ctx, cancel := requestContext(r)
	defer cancel()
	item, err := s.svc.SaveUploadedImage(ctx, header.Filename, body, descriptionPtr, tagIDs)
	if err != nil {
		respondError(w, http.StatusBadRequest, "upload_image_failed", err)
		return
	}
	respondJSON(w, http.StatusCreated, item)
}

func (s *Server) handleUpdateImage(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		TagIDs      []string `json:"tagIds"`
		Description *string  `json:"description"`
	}
	if err := decodeJSON(r, &payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_json", err)
		return
	}
	ctx, cancel := requestContext(r)
	defer cancel()
	item, err := s.svc.Repository().UpdateImageMetadata(ctx, chi.URLParam(r, "id"), payload.TagIDs, payload.Description)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "update_image_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (s *Server) handleDeleteImage(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	if err := s.svc.Repository().SoftDeleteImage(ctx, chi.URLParam(r, "id")); err != nil {
		respondError(w, http.StatusInternalServerError, "delete_image_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleRestoreImage(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	if err := s.svc.Repository().RestoreImage(ctx, chi.URLParam(r, "id")); err != nil {
		respondError(w, http.StatusInternalServerError, "restore_image_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleHardDeleteImage(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	if err := s.svc.Repository().HardDeleteImage(ctx, chi.URLParam(r, "id")); err != nil {
		respondError(w, http.StatusInternalServerError, "hard_delete_image_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleEditImage(w http.ResponseWriter, r *http.Request) {
	var payload domain.ImageEditInput
	if err := decodeJSON(r, &payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_json", err)
		return
	}
	ctx, cancel := requestContext(r)
	defer cancel()
	problemID := strings.TrimSpace(r.URL.Query().Get("problemId"))
	var problemIDPtr *string
	if problemID != "" {
		problemIDPtr = &problemID
	}
	item, err := s.svc.EditImage(ctx, chi.URLParam(r, "id"), payload, problemIDPtr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "edit_image_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (s *Server) handleImageFile(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	info, err := s.svc.Repository().GetImageStorageInfo(ctx, chi.URLParam(r, "id"))
	if handleNotFound(w, err) {
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "load_image_file_failed", err)
		return
	}
	http.ServeFile(w, r, filepath.Join(s.svc.Repository().StorageRoot(), info.StoragePath))
}

func (s *Server) handleImageThumbnail(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	info, err := s.svc.Repository().GetImageStorageInfo(ctx, chi.URLParam(r, "id"))
	if handleNotFound(w, err) {
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "load_image_thumbnail_failed", err)
		return
	}
	http.ServeFile(w, r, filepath.Join(s.svc.Repository().StorageRoot(), info.ThumbnailPath))
}

func (s *Server) handleBatchDeleteImages(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ImageIDs []string `json:"imageIds"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_json", err)
		return
	}
	if len(req.ImageIDs) == 0 {
		respondError(w, http.StatusBadRequest, "empty_image_ids", nil)
		return
	}
	ctx, cancel := requestContext(r)
	defer cancel()
	deleted, err := s.svc.Repository().BatchDeleteImages(ctx, req.ImageIDs)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "batch_delete_images_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"ok": true, "deleted": deleted})
}
