package grpc

import (
	"context"
	"fmt"

	draftv1 "github.com/snowmerak/DraftStore/gen/draft/v1"
	"github.com/snowmerak/DraftStore/lib/service/draft"
	"github.com/snowmerak/DraftStore/lib/util/errormap"
	"github.com/snowmerak/DraftStore/lib/util/logger"
)

type Server struct {
	draftv1.UnimplementedDraftServiceServer
	draftService *draft.Service
	address      string
}

type ServerOptions struct {
	DraftService *draft.Service
	Address      string
}

func NewServer(option ServerOptions) *Server {
	log := logger.GetServiceLogger("grpc-controller")

	server := &Server{
		draftService: option.DraftService,
		address:      option.Address,
	}

	log.Info().
		Str("address", server.address).
		Msg("gRPC server controller initialized")

	return server
}

// CreateDraftBucket creates the necessary buckets for draft operations
func (s *Server) CreateDraftBucket(ctx context.Context, req *draftv1.CreateDraftBucketRequest) (*draftv1.CreateDraftBucketResponse, error) {
	log := logger.GetHandlerLogger("grpc", "CreateDraftBucket", "/draft.v1.DraftService/CreateDraftBucket")

	log.Info().Msg("Handling CreateDraftBucket request")

	err := s.draftService.CreateDraftBucket(ctx)
	if err != nil {
		log.Error().
			Err(err).
			Msg("CreateDraftBucket operation failed")
		return &draftv1.CreateDraftBucketResponse{
			Result: &draftv1.Result{
				Success:      false,
				ErrorMessage: err.Error(),
				ErrorType:    errormap.MapToErrorType(err),
			},
		}, nil
	}

	log.Info().Msg("CreateDraftBucket operation completed successfully")
	return &draftv1.CreateDraftBucketResponse{
		Result: &draftv1.Result{
			Success: true,
		},
	}, nil
}

// GetUploadURL generates a presigned URL for uploading files to the draft bucket
func (s *Server) GetUploadURL(ctx context.Context, req *draftv1.GetUploadURLRequest) (*draftv1.GetUploadURLResponse, error) {
	log := logger.GetHandlerLogger("grpc", "GetUploadURL", "/draft.v1.DraftService/GetUploadURL").With().
		Str("object_name", req.ObjectName).
		Logger()

	log.Info().Msg("Handling GetUploadURL request")

	url, err := s.draftService.GetUploadURL(ctx, req.ObjectName)
	if err != nil {
		log.Error().
			Err(err).
			Msg("GetUploadURL operation failed")
		return &draftv1.GetUploadURLResponse{
			Result: &draftv1.Result{
				Success:      false,
				ErrorMessage: err.Error(),
				ErrorType:    errormap.MapToErrorType(err),
			},
		}, nil
	}

	log.Info().
		Str("url_length", fmt.Sprintf("%d", len(url))).
		Msg("GetUploadURL operation completed successfully")
	return &draftv1.GetUploadURLResponse{
		Result: &draftv1.Result{
			Success: true,
		},
		Url: url,
	}, nil
}

// GetDownloadURL generates a presigned URL for downloading files from the main bucket
func (s *Server) GetDownloadURL(ctx context.Context, req *draftv1.GetDownloadURLRequest) (*draftv1.GetDownloadURLResponse, error) {
	log := logger.GetHandlerLogger("grpc", "GetDownloadURL", "/draft.v1.DraftService/GetDownloadURL").With().
		Str("object_name", req.ObjectName).
		Logger()

	log.Info().Msg("Handling GetDownloadURL request")

	url, err := s.draftService.GetDownloadURL(ctx, req.ObjectName)
	if err != nil {
		log.Error().
			Err(err).
			Msg("GetDownloadURL operation failed")
		return &draftv1.GetDownloadURLResponse{
			Result: &draftv1.Result{
				Success:      false,
				ErrorMessage: err.Error(),
				ErrorType:    errormap.MapToErrorType(err),
			},
		}, nil
	}

	log.Info().
		Str("url_length", fmt.Sprintf("%d", len(url))).
		Msg("GetDownloadURL operation completed successfully")
	return &draftv1.GetDownloadURLResponse{
		Result: &draftv1.Result{
			Success: true,
		},
		Url: url,
	}, nil
}

// ConfirmUpload moves a file from draft bucket to main bucket
func (s *Server) ConfirmUpload(ctx context.Context, req *draftv1.ConfirmUploadRequest) (*draftv1.ConfirmUploadResponse, error) {
	log := logger.GetHandlerLogger("grpc", "ConfirmUpload", "/draft.v1.DraftService/ConfirmUpload").With().
		Str("object_name", req.ObjectName).
		Logger()

	log.Info().Msg("Handling ConfirmUpload request")

	err := s.draftService.ConfirmUpload(ctx, req.ObjectName)
	if err != nil {
		log.Error().
			Err(err).
			Msg("ConfirmUpload operation failed")
		return &draftv1.ConfirmUploadResponse{
			Result: &draftv1.Result{
				Success:      false,
				ErrorMessage: err.Error(),
				ErrorType:    errormap.MapToErrorType(err),
			},
		}, nil
	}

	log.Info().Msg("ConfirmUpload operation completed successfully")
	return &draftv1.ConfirmUploadResponse{
		Result: &draftv1.Result{
			Success: true,
		},
	}, nil
}
