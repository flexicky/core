package logger

import (
	"log/slog"
	"os"
)

// Environment names understood by the application.
const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
)

// New builds the application logger for the given environment.
// Local and dev environments log at debug level, everything else at info level.
func New(env string) *slog.Logger {
	level := slog.LevelInfo

	switch env {
	case EnvLocal, EnvDev:
		level = slog.LevelDebug
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
