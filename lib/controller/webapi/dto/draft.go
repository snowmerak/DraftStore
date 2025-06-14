package dto

import (
	draftv1 "github.com/snowmerak/DraftStore/gen/draft/v1"
)

// Re-export protobuf types for API usage
type (
	ErrorType                 = draftv1.ErrorType
	Result                    = draftv1.Result
	CreateDraftBucketRequest  = draftv1.CreateDraftBucketRequest
	CreateDraftBucketResponse = draftv1.CreateDraftBucketResponse
	GetUploadURLRequest       = draftv1.GetUploadURLRequest
	GetUploadURLResponse      = draftv1.GetUploadURLResponse
	GetDownloadURLRequest     = draftv1.GetDownloadURLRequest
	GetDownloadURLResponse    = draftv1.GetDownloadURLResponse
	ConfirmUploadRequest      = draftv1.ConfirmUploadRequest
	ConfirmUploadResponse     = draftv1.ConfirmUploadResponse
)

// Error type constants for easier access
const (
	ErrorTypeUnspecified          = draftv1.ErrorType_ERROR_TYPE_UNSPECIFIED
	ErrorTypeBucketNotFound       = draftv1.ErrorType_ERROR_TYPE_BUCKET_NOT_FOUND
	ErrorTypeObjectNotFound       = draftv1.ErrorType_ERROR_TYPE_OBJECT_NOT_FOUND
	ErrorTypeAccessDenied         = draftv1.ErrorType_ERROR_TYPE_ACCESS_DENIED
	ErrorTypeNetworkError         = draftv1.ErrorType_ERROR_TYPE_NETWORK_ERROR
	ErrorTypeStorageQuotaExceeded = draftv1.ErrorType_ERROR_TYPE_STORAGE_QUOTA_EXCEEDED
	ErrorTypeInvalidObjectName    = draftv1.ErrorType_ERROR_TYPE_INVALID_OBJECT_NAME
	ErrorTypeBucketAlreadyExists  = draftv1.ErrorType_ERROR_TYPE_BUCKET_ALREADY_EXISTS
	ErrorTypeCopyFailed           = draftv1.ErrorType_ERROR_TYPE_COPY_FAILED
	ErrorTypeDeleteFailed         = draftv1.ErrorType_ERROR_TYPE_DELETE_FAILED
	ErrorTypePresignedURLFailed   = draftv1.ErrorType_ERROR_TYPE_PRESIGNED_URL_FAILED
	ErrorTypeInternalError        = draftv1.ErrorType_ERROR_TYPE_INTERNAL_ERROR
)
