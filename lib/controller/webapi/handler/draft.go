package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/snowmerak/DraftStore/lib/controller/webapi/converter"
	"github.com/snowmerak/DraftStore/lib/controller/webapi/dto"
	"github.com/snowmerak/DraftStore/lib/service/draft"
)

type DraftHandler struct {
	draftService *draft.Service
}

func NewDraftHandler(draftService *draft.Service) *DraftHandler {
	return &DraftHandler{
		draftService: draftService,
	}
}

// CreateDraftBucket handles POST /api/v1/draft/bucket
func (h *DraftHandler) CreateDraftBucket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := h.draftService.CreateDraftBucket(ctx)
	result := converter.ConvertErrorToResult(err)

	response := &dto.CreateDraftBucketResponse{
		Result: result,
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(response)
}

// GetUploadURL handles POST /api/v1/draft/upload-url
func (h *DraftHandler) GetUploadURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dto.GetUploadURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		result := &dto.Result{
			Success:      false,
			ErrorMessage: "Invalid request body",
			ErrorType:    dto.ErrorTypeInternalError,
		}
		response := &dto.GetUploadURLResponse{Result: result}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	url, err := h.draftService.GetUploadURL(ctx, req.ObjectName)
	result := converter.ConvertErrorToResult(err)

	response := &dto.GetUploadURLResponse{
		Result: result,
		Url:    url,
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(response)
}

// GetDownloadURL handles POST /api/v1/draft/download-url
func (h *DraftHandler) GetDownloadURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dto.GetDownloadURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		result := &dto.Result{
			Success:      false,
			ErrorMessage: "Invalid request body",
			ErrorType:    dto.ErrorTypeInternalError,
		}
		response := &dto.GetDownloadURLResponse{Result: result}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	url, err := h.draftService.GetDownloadURL(ctx, req.ObjectName)
	result := converter.ConvertErrorToResult(err)

	response := &dto.GetDownloadURLResponse{
		Result: result,
		Url:    url,
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(response)
}

// ConfirmUpload handles POST /api/v1/draft/confirm
func (h *DraftHandler) ConfirmUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dto.ConfirmUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		result := &dto.Result{
			Success:      false,
			ErrorMessage: "Invalid request body",
			ErrorType:    dto.ErrorTypeInternalError,
		}
		response := &dto.ConfirmUploadResponse{Result: result}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	err := h.draftService.ConfirmUpload(ctx, req.ObjectName)
	result := converter.ConvertErrorToResult(err)

	response := &dto.ConfirmUploadResponse{
		Result: result,
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(response)
}

// RegisterRoutes registers all draft-related routes
func (h *DraftHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/draft", func(r chi.Router) {
		r.Post("/bucket", h.CreateDraftBucket)
		r.Post("/upload-url", h.GetUploadURL)
		r.Post("/download-url", h.GetDownloadURL)
		r.Post("/confirm", h.ConfirmUpload)
	})
}
