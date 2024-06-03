package logger

import (
	"context"
	"log/slog"

	"github.com/Raimguhinov/dav-go/pkg/logger/slogpretty"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/tracelog"
)

const queryLog = "Query"

func NewTracer(l *Logger) pgx.QueryTracer {
	return &tracelog.TraceLog{
		Logger:   &Logger{l.Logger},
		LogLevel: tracelog.LogLevelTrace,
	}
}

func (l *Logger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	if msg == queryLog {
		logger := l.Logger
		attrs := make([]slog.Attr, 0, len(data))
		for k, v := range data {
			if k == "sql" {
				// remove \n, \t from msg
				attrs = append(attrs, slog.String(k, slogpretty.PrettySQL(v.(string))))
				continue
			}
		}
		logger.LogAttrs(ctx, translateLevel(level), "pgx."+msg, attrs...)
	}
}

func translateLevel(level tracelog.LogLevel) slog.Level {
	switch level {
	case tracelog.LogLevelTrace:
		return slog.LevelDebug
	case tracelog.LogLevelDebug:
		return slog.LevelDebug
	case tracelog.LogLevelInfo:
		return slog.LevelDebug
	case tracelog.LogLevelWarn:
		return slog.LevelWarn
	case tracelog.LogLevelError:
		return slog.LevelError
	case tracelog.LogLevelNone:
		return slog.LevelError
	default:
		return slog.LevelError
	}
}
