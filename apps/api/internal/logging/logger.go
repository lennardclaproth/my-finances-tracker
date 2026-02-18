package logging

import "context"

type Logger interface {
	Info(ctx context.Context, msg string, fields ...any)
	Error(ctx context.Context, msg string, err error, fields ...any)
}
