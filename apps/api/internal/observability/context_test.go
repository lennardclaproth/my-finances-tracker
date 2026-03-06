package observability

import (
	"context"
	"testing"
)

func TestEnsureRequestAndCorrelationIDs_GeneratesWhenMissing(t *testing.T) {
	t.Parallel()

	ctx, requestID, correlationID := EnsureRequestAndCorrelationIDs(context.Background(), "", "")
	if requestID == "" {
		t.Fatalf("expected request id to be generated")
	}
	if correlationID == "" {
		t.Fatalf("expected correlation id to be generated")
	}
	if requestID != correlationID {
		t.Fatalf("expected correlation id to default to request id")
	}
	if got := RequestIDFromContext(ctx); got != requestID {
		t.Fatalf("expected request id in context, got %q", got)
	}
	if got := CorrelationIDFromContext(ctx); got != correlationID {
		t.Fatalf("expected correlation id in context, got %q", got)
	}
}

func TestEnsureRequestAndCorrelationIDs_PreservesProvidedValues(t *testing.T) {
	t.Parallel()

	ctx, requestID, correlationID := EnsureRequestAndCorrelationIDs(context.Background(), "req-123", "corr-456")
	if requestID != "req-123" {
		t.Fatalf("expected request id req-123, got %q", requestID)
	}
	if correlationID != "corr-456" {
		t.Fatalf("expected correlation id corr-456, got %q", correlationID)
	}
	if got := RequestIDFromContext(ctx); got != "req-123" {
		t.Fatalf("expected request id in context, got %q", got)
	}
	if got := CorrelationIDFromContext(ctx); got != "corr-456" {
		t.Fatalf("expected correlation id in context, got %q", got)
	}
}
