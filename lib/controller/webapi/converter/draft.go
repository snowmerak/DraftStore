package converter

import (
	"strings"

	draftv1 "github.com/snowmerak/DraftStore/gen/draft/v1"
)

// ConvertErrorToResult converts a Go error to a protobuf Result
func ConvertErrorToResult(err error) *draftv1.Result {
	if err == nil {
		return &draftv1.Result{Success: true}
	}

	return &draftv1.Result{
		Success:      false,
		ErrorMessage: err.Error(),
		ErrorType:    mapErrorType(err),
	}
}

// mapErrorType maps Go errors to protobuf ErrorType
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
