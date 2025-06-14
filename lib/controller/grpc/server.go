package grpc

import (
	"context"

	draftv1 "github.com/snowmerak/DraftStore/gen/draft/v1"
	"github.com/snowmerak/DraftStore/lib/service/draft"
	"github.com/snowmerak/DraftStore/lib/util/errormap"
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
	return &Server{
		draftService: option.DraftService,
		address:      option.Address,
	}
}

// CreateDraftBucket creates the necessary buckets for draft operations
func (s *Server) CreateDraftBucket(ctx context.Context, req *draftv1.CreateDraftBucketRequest) (*draftv1.CreateDraftBucketResponse, error) {
	err := s.draftService.CreateDraftBucket(ctx)
	if err != nil {
		return &draftv1.CreateDraftBucketResponse{
			Result: &draftv1.Result{
				Success:      false,
				ErrorMessage: err.Error(),
				ErrorType:    errormap.MapToErrorType(err),
			},
		}, nil
	}

	return &draftv1.CreateDraftBucketResponse{
		Result: &draftv1.Result{
			Success: true,
		},
	}, nil
}

// GetUploadURL generates a presigned URL for uploading files to the draft bucket
func (s *Server) GetUploadURL(ctx context.Context, req *draftv1.GetUploadURLRequest) (*draftv1.GetUploadURLResponse, error) {
	url, err := s.draftService.GetUploadURL(ctx, req.ObjectName)
	if err != nil {
		return &draftv1.GetUploadURLResponse{
			Result: &draftv1.Result{
				Success:      false,
				ErrorMessage: err.Error(),
				ErrorType:    errormap.MapToErrorType(err),
			},
		}, nil
	}

	return &draftv1.GetUploadURLResponse{
		Result: &draftv1.Result{
			Success: true,
		},
		Url: url,
	}, nil
}

// GetDownloadURL generates a presigned URL for downloading files from the main bucket
func (s *Server) GetDownloadURL(ctx context.Context, req *draftv1.GetDownloadURLRequest) (*draftv1.GetDownloadURLResponse, error) {
	url, err := s.draftService.GetDownloadURL(ctx, req.ObjectName)
	if err != nil {
		return &draftv1.GetDownloadURLResponse{
			Result: &draftv1.Result{
				Success:      false,
				ErrorMessage: err.Error(),
				ErrorType:    errormap.MapToErrorType(err),
			},
		}, nil
	}

	return &draftv1.GetDownloadURLResponse{
		Result: &draftv1.Result{
			Success: true,
		},
		Url: url,
	}, nil
}

// ConfirmUpload moves a file from draft bucket to main bucket
func (s *Server) ConfirmUpload(ctx context.Context, req *draftv1.ConfirmUploadRequest) (*draftv1.ConfirmUploadResponse, error) {
	err := s.draftService.ConfirmUpload(ctx, req.ObjectName)
	if err != nil {
		return &draftv1.ConfirmUploadResponse{
			Result: &draftv1.Result{
				Success:      false,
				ErrorMessage: err.Error(),
				ErrorType:    errormap.MapToErrorType(err),
			},
		}, nil
	}

	return &draftv1.ConfirmUploadResponse{
		Result: &draftv1.Result{
			Success: true,
		},
	}, nil
}
