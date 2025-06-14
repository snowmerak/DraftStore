package grpc

import (
	"context"
	"strings"

	draftv1 "github.com/snowmerak/DraftStore/gen/draft/v1"
	"github.com/snowmerak/DraftStore/lib/service/draft"
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
				ErrorType:    mapErrorType(err),
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
				ErrorType:    mapErrorType(err),
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
				ErrorType:    mapErrorType(err),
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
				ErrorType:    mapErrorType(err),
			},
		}, nil
	}

	return &draftv1.ConfirmUploadResponse{
		Result: &draftv1.Result{
			Success: true,
		},
	}, nil
}

// mapErrorType maps Go errors to protobuf ErrorType enum
func mapErrorType(err error) draftv1.ErrorType {
	if err == nil {
		return draftv1.ErrorType_ERROR_TYPE_UNSPECIFIED
	}

	errMsg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errMsg, "bucket") && (strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "does not exist")):
		return draftv1.ErrorType_ERROR_TYPE_BUCKET_NOT_FOUND
	case strings.Contains(errMsg, "object") && (strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "does not exist")):
		return draftv1.ErrorType_ERROR_TYPE_OBJECT_NOT_FOUND
	case strings.Contains(errMsg, "access denied") || strings.Contains(errMsg, "permission"):
		return draftv1.ErrorType_ERROR_TYPE_ACCESS_DENIED
	case strings.Contains(errMsg, "network") || strings.Contains(errMsg, "connection"):
		return draftv1.ErrorType_ERROR_TYPE_NETWORK_ERROR
	case strings.Contains(errMsg, "quota") || strings.Contains(errMsg, "storage full"):
		return draftv1.ErrorType_ERROR_TYPE_STORAGE_QUOTA_EXCEEDED
	case strings.Contains(errMsg, "invalid") && strings.Contains(errMsg, "object"):
		return draftv1.ErrorType_ERROR_TYPE_INVALID_OBJECT_NAME
	case strings.Contains(errMsg, "bucket") && strings.Contains(errMsg, "already exists"):
		return draftv1.ErrorType_ERROR_TYPE_BUCKET_ALREADY_EXISTS
	case strings.Contains(errMsg, "copy") && strings.Contains(errMsg, "failed"):
		return draftv1.ErrorType_ERROR_TYPE_COPY_FAILED
	case strings.Contains(errMsg, "delete") && strings.Contains(errMsg, "failed"):
		return draftv1.ErrorType_ERROR_TYPE_DELETE_FAILED
	case strings.Contains(errMsg, "presigned") || strings.Contains(errMsg, "url"):
		return draftv1.ErrorType_ERROR_TYPE_PRESIGNED_URL_FAILED
	default:
		return draftv1.ErrorType_ERROR_TYPE_INTERNAL_ERROR
	}
}
