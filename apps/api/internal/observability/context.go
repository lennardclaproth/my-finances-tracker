package observability

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

const (
	HeaderRequestID     = "X-Request-ID"
	HeaderCorrelationID = "X-Correlation-ID"
	HeaderTraceparent   = "traceparent"
	HeaderTracestate    = "tracestate"
)

type contextKey string

const (
	requestIDContextKey     contextKey = "request_id"
	correlationIDContextKey contextKey = "correlation_id"
)

func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDContextKey, requestID)
}

func ContextWithCorrelationID(ctx context.Context, correlationID string) context.Context {
	correlationID = strings.TrimSpace(correlationID)
	if correlationID == "" {
		return ctx
	}
	return context.WithValue(ctx, correlationIDContextKey, correlationID)
}

func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	requestID, _ := ctx.Value(requestIDContextKey).(string)
	return strings.TrimSpace(requestID)
}

func CorrelationIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	correlationID, _ := ctx.Value(correlationIDContextKey).(string)
	return strings.TrimSpace(correlationID)
}

func EnsureRequestAndCorrelationIDs(ctx context.Context, requestID, correlationID string) (context.Context, string, string) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		requestID = RequestIDFromContext(ctx)
	}
	if requestID == "" {
		requestID = uuid.NewString()
	}

	correlationID = strings.TrimSpace(correlationID)
	if correlationID == "" {
		correlationID = CorrelationIDFromContext(ctx)
	}
	if correlationID == "" {
		correlationID = requestID
	}

	ctx = ContextWithRequestID(ctx, requestID)
	ctx = ContextWithCorrelationID(ctx, correlationID)
	return ctx, requestID, correlationID
}

func CorrelationUUIDFromContext(ctx context.Context) uuid.UUID {
	correlationID := CorrelationIDFromContext(ctx)
	if correlationID == "" {
		return uuid.Nil
	}
	id, err := uuid.Parse(correlationID)
	if err == nil {
		return id
	}
	return uuid.NewSHA1(uuid.Nil, []byte(correlationID))
}
