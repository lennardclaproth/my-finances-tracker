package logging

import (
	"context"
	"log/slog"
	"os"
)

type SlogLogger struct {
	logger *slog.Logger
}

func NewSlogLogger(level slog.Leveler) *SlogLogger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	return &SlogLogger{logger: slog.New(handler)}
}

func (l *SlogLogger) Info(ctx context.Context, msg string, args ...any) {
	l.logger.InfoContext(ctx, msg, args...)
}

func (l *SlogLogger) Error(ctx context.Context, msg string, err error, args ...any) {
	fields := append(args, "error", err.Error())
	l.logger.ErrorContext(ctx, msg, fields...)
}
