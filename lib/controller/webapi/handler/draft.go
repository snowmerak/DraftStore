package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/snowmerak/DraftStore/lib/controller/webapi/converter"
	"github.com/snowmerak/DraftStore/lib/controller/webapi/dto"
	"github.com/snowmerak/DraftStore/lib/service/draft"
	"github.com/snowmerak/DraftStore/lib/util/logger"
)

type DraftHandler struct {
	draftService *draft.Service
}

func NewDraftHandler(draftService *draft.Service) *DraftHandler {
	log := logger.GetServiceLogger("webapi-handler")

	handler := &DraftHandler{
		draftService: draftService,
	}

	log.Info().Msg("WebAPI draft handler initialized")
	return handler
}

// CreateDraftBucket handles POST /api/v1/draft/bucket
func (h *DraftHandler) CreateDraftBucket(w http.ResponseWriter, r *http.Request) {
	log := logger.GetHandlerLogger("http", "POST", "/api/v1/draft/bucket")
	ctx := r.Context()

	log.Info().Msg("Handling CreateDraftBucket request")

	err := h.draftService.CreateDraftBucket(ctx)
	result := converter.ConvertErrorToResult(err)

	response := &dto.CreateDraftBucketResponse{
		Result: result,
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		log.Error().
			Err(err).
			Msg("CreateDraftBucket operation failed")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		log.Info().Msg("CreateDraftBucket operation completed successfully")
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(response)
}

// GetUploadURL handles POST /api/v1/draft/upload-url
func (h *DraftHandler) GetUploadURL(w http.ResponseWriter, r *http.Request) {
	log := logger.GetHandlerLogger("http", "POST", "/api/v1/draft/upload-url")
	ctx := r.Context()

	var req dto.GetUploadURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to decode request body")
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

	log.Info().
		Str("object_name", req.ObjectName).
		Msg("Handling GetUploadURL request")

	url, err := h.draftService.GetUploadURL(ctx, req.ObjectName)
	result := converter.ConvertErrorToResult(err)

	response := &dto.GetUploadURLResponse{
		Result: result,
		Url:    url,
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		log.Error().
			Err(err).
			Str("object_name", req.ObjectName).
			Msg("GetUploadURL operation failed")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		log.Info().
			Str("object_name", req.ObjectName).
			Msg("GetUploadURL operation completed successfully")
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(response)
}

// GetDownloadURL handles POST /api/v1/draft/download-url
func (h *DraftHandler) GetDownloadURL(w http.ResponseWriter, r *http.Request) {
	log := logger.GetHandlerLogger("http", "POST", "/api/v1/draft/download-url")
	ctx := r.Context()

	var req dto.GetDownloadURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to decode request body")
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

	log.Info().
		Str("object_name", req.ObjectName).
		Msg("Handling GetDownloadURL request")

	url, err := h.draftService.GetDownloadURL(ctx, req.ObjectName)
	result := converter.ConvertErrorToResult(err)

	response := &dto.GetDownloadURLResponse{
		Result: result,
		Url:    url,
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		log.Error().
			Err(err).
			Str("object_name", req.ObjectName).
			Msg("GetDownloadURL operation failed")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		log.Info().
			Str("object_name", req.ObjectName).
			Msg("GetDownloadURL operation completed successfully")
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(response)
}

// ConfirmUpload handles POST /api/v1/draft/confirm
func (h *DraftHandler) ConfirmUpload(w http.ResponseWriter, r *http.Request) {
	log := logger.GetHandlerLogger("http", "POST", "/api/v1/draft/confirm")
	ctx := r.Context()

	var req dto.ConfirmUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to decode request body")
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

	log.Info().
		Str("object_name", req.ObjectName).
		Msg("Handling ConfirmUpload request")

	err := h.draftService.ConfirmUpload(ctx, req.ObjectName)
	result := converter.ConvertErrorToResult(err)

	response := &dto.ConfirmUploadResponse{
		Result: result,
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		log.Error().
			Err(err).
			Str("object_name", req.ObjectName).
			Msg("ConfirmUpload operation failed")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		log.Info().
			Str("object_name", req.ObjectName).
			Msg("ConfirmUpload operation completed successfully")
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
