package errormap

import (
	"strings"

	draftv1 "github.com/snowmerak/DraftStore/gen/draft/v1"
)

// MapToErrorType maps Go errors to protobuf ErrorType enum
func MapToErrorType(err error) draftv1.ErrorType {
	if err == nil {
		return draftv1.ErrorType_ERROR_TYPE_UNSPECIFIED
	}

	errMsg := strings.ToLower(err.Error())

	switch {
	case isBucketNotFound(errMsg):
		return draftv1.ErrorType_ERROR_TYPE_BUCKET_NOT_FOUND
	case isObjectNotFound(errMsg):
		return draftv1.ErrorType_ERROR_TYPE_OBJECT_NOT_FOUND
	case isAccessDenied(errMsg):
		return draftv1.ErrorType_ERROR_TYPE_ACCESS_DENIED
	case isNetworkError(errMsg):
		return draftv1.ErrorType_ERROR_TYPE_NETWORK_ERROR
	case isStorageQuotaExceeded(errMsg):
		return draftv1.ErrorType_ERROR_TYPE_STORAGE_QUOTA_EXCEEDED
	case isInvalidObjectName(errMsg):
		return draftv1.ErrorType_ERROR_TYPE_INVALID_OBJECT_NAME
	case isBucketAlreadyExists(errMsg):
		return draftv1.ErrorType_ERROR_TYPE_BUCKET_ALREADY_EXISTS
	case isCopyFailed(errMsg):
		return draftv1.ErrorType_ERROR_TYPE_COPY_FAILED
	case isDeleteFailed(errMsg):
		return draftv1.ErrorType_ERROR_TYPE_DELETE_FAILED
	case isPresignedURLFailed(errMsg):
		return draftv1.ErrorType_ERROR_TYPE_PRESIGNED_URL_FAILED
	default:
		return draftv1.ErrorType_ERROR_TYPE_INTERNAL_ERROR
	}
}

// Error pattern matching functions
func isBucketNotFound(errMsg string) bool {
	return strings.Contains(errMsg, "bucket") &&
		(strings.Contains(errMsg, "not found") ||
			strings.Contains(errMsg, "does not exist") ||
			strings.Contains(errMsg, "no such bucket"))
}

func isObjectNotFound(errMsg string) bool {
	return strings.Contains(errMsg, "object") &&
		(strings.Contains(errMsg, "not found") ||
			strings.Contains(errMsg, "does not exist") ||
			strings.Contains(errMsg, "no such key"))
}

func isAccessDenied(errMsg string) bool {
	return strings.Contains(errMsg, "access denied") ||
		strings.Contains(errMsg, "permission") ||
		strings.Contains(errMsg, "forbidden") ||
		strings.Contains(errMsg, "unauthorized")
}

func isNetworkError(errMsg string) bool {
	return strings.Contains(errMsg, "network") ||
		strings.Contains(errMsg, "connection") ||
		strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "dial")
}

func isStorageQuotaExceeded(errMsg string) bool {
	return strings.Contains(errMsg, "quota") ||
		strings.Contains(errMsg, "storage full") ||
		strings.Contains(errMsg, "insufficient storage") ||
		strings.Contains(errMsg, "storage limit")
}

func isInvalidObjectName(errMsg string) bool {
	return strings.Contains(errMsg, "invalid") &&
		(strings.Contains(errMsg, "object") ||
			strings.Contains(errMsg, "key") ||
			strings.Contains(errMsg, "name"))
}

func isBucketAlreadyExists(errMsg string) bool {
	return strings.Contains(errMsg, "bucket") &&
		(strings.Contains(errMsg, "already exists") ||
			strings.Contains(errMsg, "already owned"))
}

func isCopyFailed(errMsg string) bool {
	return strings.Contains(errMsg, "copy") &&
		strings.Contains(errMsg, "failed")
}

func isDeleteFailed(errMsg string) bool {
	return strings.Contains(errMsg, "delete") &&
		strings.Contains(errMsg, "failed")
}

func isPresignedURLFailed(errMsg string) bool {
	return strings.Contains(errMsg, "presigned") ||
		(strings.Contains(errMsg, "url") &&
			(strings.Contains(errMsg, "generate") ||
				strings.Contains(errMsg, "create")))
}
