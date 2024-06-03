package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/Raimguhinov/dav-go/pkg/logger/slogpretty"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

// Logger -.
type Logger struct {
	*slog.Logger
}

// New -.
func New(level, env string) *Logger {
	var lev slog.Level

	switch strings.ToLower(level) {
	case "error":
		lev = slog.LevelError
	case "warn":
		lev = slog.LevelWarn
	case "info":
		lev = slog.LevelInfo
	case "debug":
		lev = slog.LevelDebug
	default:
		lev = slog.LevelInfo
	}

	var logger *slog.Logger

	switch env {
	case envLocal:
		logger = setupPrettySlog(lev)
	case envDev:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lev}),
		)
	case envProd:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return &Logger{logger}
}

func setupPrettySlog(level slog.Level) *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: level,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}

func (l *Logger) Printf(msg string, args ...interface{}) {
	l.Info(strings.TrimSpace(strings.ReplaceAll(msg, "%v", "")), slog.Any("args", args))
}

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func Query(q string) slog.Attr {
	return slog.Attr{
		Key:   "query",
		Value: slog.StringValue(slogpretty.PrettySQL(q)),
	}
}
