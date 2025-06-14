package converter

import (
	draftv1 "github.com/snowmerak/DraftStore/gen/draft/v1"
	"github.com/snowmerak/DraftStore/lib/util/errormap"
)

// ConvertErrorToResult converts a Go error to a protobuf Result
func ConvertErrorToResult(err error) *draftv1.Result {
	if err == nil {
		return &draftv1.Result{Success: true}
	}

	return &draftv1.Result{
		Success:      false,
		ErrorMessage: err.Error(),
		ErrorType:    errormap.MapToErrorType(err),
	}
}
