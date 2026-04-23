package logging

import (
	"context"
	"log/slog"
	"os"

	"github.com/lennardclaproth/my-finances-tracker/internal/observability"
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

func (l *SlogLogger) Debug(ctx context.Context, msg string, args ...any) {
	fields := observability.AppendContextFields(ctx, args...)
	fields = observability.FilterFields(fields...)
	l.logger.DebugContext(ctx, msg, fields...)
}

func (l *SlogLogger) Info(ctx context.Context, msg string, args ...any) {
	fields := observability.AppendContextFields(ctx, args...)
	fields = observability.FilterFields(fields...)
	l.logger.InfoContext(ctx, msg, fields...)
}

func (l *SlogLogger) Warn(ctx context.Context, msg string, args ...any) {
	fields := observability.AppendContextFields(ctx, args...)
	fields = observability.FilterFields(fields...)
	l.logger.WarnContext(ctx, msg, fields...)
}

func (l *SlogLogger) Error(ctx context.Context, msg string, err error, args ...any) {
	fields := append(args, "error", err.Error())
	fields = observability.AppendContextFields(ctx, fields...)
	fields = observability.FilterFields(fields...)
	l.logger.ErrorContext(ctx, msg, fields...)
}
