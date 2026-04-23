package logging

import "context"

// Logger defines the minimal structured logging API used by application code.
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...any)
	Info(ctx context.Context, msg string, fields ...any)
	Warn(ctx context.Context, msg string, fields ...any)
	Error(ctx context.Context, msg string, err error, fields ...any)
}
