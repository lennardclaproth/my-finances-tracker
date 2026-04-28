package messaging

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/observability"
	"go.elastic.co/apm/module/apmhttp/v2"
	"go.elastic.co/apm/v2"
)

type testBusMessage struct {
	Value string `json:"value"`
}

func (m testBusMessage) MessageTopic() string {
	return "test.bus.topic"
}

func TestNewJSONEnvelopeFromContext_PropagatesHeaders(t *testing.T) {
	t.Parallel()

	tx := apm.DefaultTracer().StartTransaction("producer", "request")
	defer tx.End()

	ctx := apm.ContextWithTransaction(context.Background(), tx)
	ctx = observability.ContextWithRequestID(ctx, "req-123")
	ctx = observability.ContextWithCorrelationID(ctx, "corr-456")

	env, err := NewJSONEnvelopeFromContext(ctx, testBusMessage{Value: "hello"})
	if err != nil {
		t.Fatalf("expected no error creating envelope, got %v", err)
	}

	if got := env.Headers[observability.HeaderRequestID]; got != "req-123" {
		t.Fatalf("unexpected request_id header: %q", got)
	}
	if got := env.Headers[observability.HeaderCorrelationID]; got != "corr-456" {
		t.Fatalf("unexpected correlation_id header: %q", got)
	}
	traceparent := env.Headers[observability.HeaderTraceparent]
	if traceparent == "" {
		t.Fatalf("expected traceparent header")
	}
	parsed, err := apmhttp.ParseTraceparentHeader(traceparent)
	if err != nil {
		t.Fatalf("failed to parse traceparent: %v", err)
	}
	if parsed.Trace.String() != tx.TraceContext().Trace.String() {
		t.Fatalf("expected propagated trace id %s, got %s", tx.TraceContext().Trace.String(), parsed.Trace.String())
	}
	if env.CorrelationID == uuid.Nil {
		t.Fatalf("expected non-nil envelope correlation id")
	}
}

func TestDecodeHandler_ContinuesTraceAndCorrelation(t *testing.T) {
	t.Parallel()

	parentTx := apm.DefaultTracer().StartTransaction("producer", "request")
	defer parentTx.End()

	parentCtx := apm.ContextWithTransaction(context.Background(), parentTx)
	parentCtx = observability.ContextWithRequestID(parentCtx, "req-a")
	parentCtx = observability.ContextWithCorrelationID(parentCtx, "corr-b")

	env, err := NewJSONEnvelopeFromContext(parentCtx, testBusMessage{Value: "hello"})
	if err != nil {
		t.Fatalf("expected no error creating envelope, got %v", err)
	}

	reg := NewRegistry(JSONCodec{})
	var gotTraceID string
	var gotParentID string
	var gotRequestID string
	var gotCorrelationID string
	handler := DecodeHandler[testBusMessage](reg, func(ctx context.Context, _ Envelope, _ testBusMessage) error {
		tx := apm.TransactionFromContext(ctx)
		if tx == nil {
			t.Fatalf("expected transaction in consumer context")
		}
		gotTraceID = tx.TraceContext().Trace.String()
		gotParentID = tx.ParentID().String()
		gotRequestID = observability.RequestIDFromContext(ctx)
		gotCorrelationID = observability.CorrelationIDFromContext(ctx)
		return nil
	})

	if err := handler(context.Background(), env); err != nil {
		t.Fatalf("expected handler to succeed, got %v", err)
	}

	if gotTraceID != parentTx.TraceContext().Trace.String() {
		t.Fatalf("expected child transaction to continue trace id %s, got %s", parentTx.TraceContext().Trace.String(), gotTraceID)
	}
	if gotParentID != parentTx.TraceContext().Span.String() {
		t.Fatalf("expected parent id %s, got %s", parentTx.TraceContext().Span.String(), gotParentID)
	}
	if gotRequestID != "req-a" {
		t.Fatalf("expected propagated request id req-a, got %q", gotRequestID)
	}
	if gotCorrelationID != "corr-b" {
		t.Fatalf("expected propagated correlation id corr-b, got %q", gotCorrelationID)
	}
}
