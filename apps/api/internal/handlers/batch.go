package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"xians/api/internal/models"
	"xians/api/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// maxProxyUploadBytes caps the legacy proxy upload endpoint. Vercel rejects
// request bodies over 4.5 MB, so anything larger must use the presigned R2
// URL issued by /api/batches/init.
const maxProxyUploadBytes = 4_500_000

type BatchHandler struct {
	uploadSvc  *services.UploadService
	gallerySvc *services.GalleryService
}

func NewBatchHandler(uploadSvc *services.UploadService, gallerySvc *services.GalleryService) *BatchHandler {
	return &BatchHandler{uploadSvc: uploadSvc, gallerySvc: gallerySvc}
}

func (h *BatchHandler) Init(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Files []models.FileManifest `json:"files"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = forwarded
	}

	result, err := h.uploadSvc.InitBatch(r.Context(), req.Files, ip)
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("init batch failed")

		switch {
		case errors.Is(err, services.ErrTooManyFiles),
			errors.Is(err, services.ErrFileTooLarge),
			errors.Is(err, services.ErrBatchTooLarge),
			errors.Is(err, services.ErrUnsupportedType),
			errors.Is(err, services.ErrEmptyFile):
			writeError(w, http.StatusUnprocessableEntity, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to initialize upload")
		}
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *BatchHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	batchID := chi.URLParam(r, "id")
	fileID := chi.URLParam(r, "fileId")

	// Vercel caps request bodies at 4.5 MB; large files must go straight to
	// the presigned R2 URL returned by /api/batches/init.
	r.Body = http.MaxBytesReader(w, r.Body, maxProxyUploadBytes)

	if err := h.uploadSvc.UploadFile(r.Context(), batchID, fileID, r.Body); err != nil {
		log.Ctx(r.Context()).Error().Err(err).Str("batchId", batchID).Str("fileId", fileID).Msg("upload file failed")

		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "file too large for proxy upload; use the presigned upload URL")
			return
		}

		switch {
		case errors.Is(err, services.ErrBatchNotFound):
			writeError(w, http.StatusNotFound, "batch not found")
		case errors.Is(err, services.ErrBatchNotPending):
			writeError(w, http.StatusConflict, "batch is not in pending state")
		case errors.Is(err, services.ErrFileNotFound):
			writeError(w, http.StatusNotFound, "file not found in batch")
		default:
			log.Error().Err(err).Str("batchId", batchID).Str("fileId", fileID).Msg("upload failed")
			writeError(w, http.StatusInternalServerError, "failed to upload file")
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BatchHandler) Complete(w http.ResponseWriter, r *http.Request) {
	batchID := chi.URLParam(r, "id")

	var req models.CompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.uploadSvc.CompleteBatch(r.Context(), batchID, req)
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("complete batch failed")

		switch {
		case errors.Is(err, services.ErrBatchNotFound):
			writeError(w, http.StatusNotFound, "batch not found")
		case errors.Is(err, services.ErrBatchNotPending):
			writeError(w, http.StatusConflict, "batch is not in pending state")
		case errors.Is(err, services.ErrUploadNotVerified):
			writeError(w, http.StatusFailedDependency, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to complete batch")
		}
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *BatchHandler) GetGallery(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	hasAccess := hasGalleryAccess(r, slug)

	result, err := h.gallerySvc.GetGallery(r.Context(), slug, hasAccess)
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Str("slug", slug).Msg("get gallery failed")

		switch {
		case errors.Is(err, services.ErrBatchNotFound):
			writeError(w, http.StatusNotFound, "gallery not found")
		case errors.Is(err, services.ErrGalleryNotActive):
			writeError(w, http.StatusGone, "gallery is no longer available")
		default:
			writeError(w, http.StatusInternalServerError, "failed to get gallery")
		}
		return
	}

	if result.RequiresPassword {
		writeJSON(w, http.StatusUnauthorized, map[string]bool{"requiresPassword": true})
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=60, s-maxage=60")
	writeJSON(w, http.StatusOK, result)
}

func (h *BatchHandler) Unlock(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	_, err := h.gallerySvc.Unlock(r.Context(), slug, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidPassword):
			writeError(w, http.StatusUnauthorized, "invalid password")
		case errors.Is(err, services.ErrPasswordRequired):
			writeError(w, http.StatusBadRequest, "this gallery does not require a password")
		default:
			log.Ctx(r.Context()).Error().Err(err).Str("slug", slug).Msg("unlock failed")
			writeError(w, http.StatusInternalServerError, "failed to unlock gallery")
		}
		return
	}

	setGalleryCookie(w, slug, r)

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *BatchHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	fileID := chi.URLParam(r, "fileId")

	hasAccess := hasGalleryAccess(r, slug)
	url, err := h.gallerySvc.GetFileDownloadURL(r.Context(), slug, fileID, hasAccess)
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Str("slug", slug).Str("fileId", fileID).Msg("download file failed")

		switch {
		case errors.Is(err, services.ErrBatchNotFound):
			writeError(w, http.StatusNotFound, "file not found")
		case errors.Is(err, services.ErrGalleryNotActive):
			writeError(w, http.StatusGone, "gallery is no longer available")
		case errors.Is(err, services.ErrPasswordRequired), errors.Is(err, services.ErrInvalidPassword):
			writeError(w, http.StatusUnauthorized, "password required")
		case errors.Is(err, services.ErrDownloadLimitMet):
			writeError(w, http.StatusGone, "download limit reached")
		default:
			writeError(w, http.StatusInternalServerError, "failed to get download URL")
		}
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func (h *BatchHandler) ViewFile(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	fileID := chi.URLParam(r, "fileId")

	hasAccess := hasGalleryAccess(r, slug)
	url, err := h.gallerySvc.GetFileViewURL(r.Context(), slug, fileID, hasAccess)
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Str("slug", slug).Str("fileId", fileID).Msg("view file failed")

		switch {
		case errors.Is(err, services.ErrBatchNotFound):
			writeError(w, http.StatusNotFound, "file not found")
		case errors.Is(err, services.ErrGalleryNotActive):
			writeError(w, http.StatusGone, "gallery is no longer available")
		case errors.Is(err, services.ErrPasswordRequired), errors.Is(err, services.ErrInvalidPassword):
			writeError(w, http.StatusUnauthorized, "password required")
		default:
			writeError(w, http.StatusInternalServerError, "failed to get file")
		}
		return
	}

	// Redirect to the presigned R2 URL: media streams straight from
	// Cloudflare with Range/seek support instead of proxying through this
	// server (which would hit Vercel's request duration limits).
	w.Header().Set("Cache-Control", "private, max-age=0, must-revalidate")
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *BatchHandler) DownloadZip(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	hasAccess := hasGalleryAccess(r, slug)
	if err := h.gallerySvc.StreamZip(r.Context(), slug, w, hasAccess); err != nil {
		log.Ctx(r.Context()).Error().Err(err).Str("slug", slug).Msg("zip download failed")

		switch {
		case errors.Is(err, services.ErrBatchNotFound):
			writeError(w, http.StatusNotFound, "gallery not found")
		case errors.Is(err, services.ErrGalleryNotActive):
			writeError(w, http.StatusGone, "gallery is no longer available")
		case errors.Is(err, services.ErrPasswordRequired), errors.Is(err, services.ErrInvalidPassword):
			writeError(w, http.StatusUnauthorized, "password required")
		default:
			writeError(w, http.StatusInternalServerError, "failed to generate zip")
		}
	}
}

func (h *BatchHandler) Report(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	if err := h.gallerySvc.Report(r.Context(), slug); err != nil {
		log.Ctx(r.Context()).Error().Err(err).Str("slug", slug).Msg("report failed")

		if errors.Is(err, services.ErrBatchNotFound) {
			writeError(w, http.StatusNotFound, "gallery not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to report gallery")
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BatchHandler) Delete(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	ownerToken := r.Header.Get("X-Owner-Token")

	if ownerToken == "" {
		writeError(w, http.StatusUnauthorized, "owner token required")
		return
	}

	if err := h.gallerySvc.Delete(r.Context(), slug, ownerToken); err != nil {
		log.Ctx(r.Context()).Error().Err(err).Str("slug", slug).Msg("delete failed")

		switch {
		case errors.Is(err, services.ErrBatchNotFound):
			writeError(w, http.StatusNotFound, "gallery not found")
		case errors.Is(err, services.ErrInvalidOwnerToken):
			writeError(w, http.StatusForbidden, "invalid owner token")
		default:
			writeError(w, http.StatusInternalServerError, "failed to delete gallery")
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BatchHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
