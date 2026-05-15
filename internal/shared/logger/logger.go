package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func Init(appName string) {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	Log = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", appName).
		Logger()
}
