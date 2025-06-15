package logger

import (
	"os" // Added import for os.Stderr
	"time"

	"github.com/rs/zerolog"
)

// Log is the global logger instance.
var Log zerolog.Logger

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel) // Set the global log level to Info
	zerolog.TimeFieldFormat = time.RFC3339    // Set the time format to RFC3339
	zerolog.TimestampFieldName = "time"

	Log = zerolog.New(os.Stderr).With().
		Timestamp(). // Adds timestamp field with global "time" and RFC3339 format
		Caller().    // Adds caller field
		Logger()
}
