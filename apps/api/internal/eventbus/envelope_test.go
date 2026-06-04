package eventbus_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/eventbus"
)

func TestNewEnvelopeRootCorrelatesToSelf(t *testing.T) {
	env := eventbus.NewEnvelope(context.Background(), "topic.root", nil)

	if env.MessageID == uuid.Nil {
		t.Fatal("expected a generated message id")
	}
	if env.Topic != "topic.root" {
		t.Fatalf("unexpected topic %q", env.Topic)
	}
	if env.CorrelationID != env.MessageID.String() {
		t.Fatalf("root correlation should default to message id: got %q want %q", env.CorrelationID, env.MessageID.String())
	}
	if env.CausationID != "" {
		t.Fatalf("root message should have no causation, got %q", env.CausationID)
	}
	if env.OccurredAt.IsZero() {
		t.Fatal("expected OccurredAt to be set")
	}
}

func TestNewEnvelopeInheritsParentChain(t *testing.T) {
	parent := eventbus.Metadata{MessageID: uuid.New(), CorrelationID: "corr-123"}
	ctx := eventbus.ContextWithMetadata(context.Background(), parent)

	env := eventbus.NewEnvelope(ctx, "topic.child", nil)

	if env.CorrelationID != "corr-123" {
		t.Fatalf("expected inherited correlation, got %q", env.CorrelationID)
	}
	if env.CausationID != parent.MessageID.String() {
		t.Fatalf("expected causation = parent message id, got %q", env.CausationID)
	}
}

func TestNewEnvelopeOptionsOverrideParent(t *testing.T) {
	parent := eventbus.Metadata{MessageID: uuid.New(), CorrelationID: "corr-parent"}
	ctx := eventbus.ContextWithMetadata(context.Background(), parent)

	env := eventbus.NewEnvelope(ctx, "topic.child", nil,
		eventbus.WithCorrelationID("corr-explicit"),
		eventbus.WithCausationID("cause-explicit"),
	)

	if env.CorrelationID != "corr-explicit" {
		t.Fatalf("explicit correlation should win, got %q", env.CorrelationID)
	}
	if env.CausationID != "cause-explicit" {
		t.Fatalf("explicit causation should win, got %q", env.CausationID)
	}
}
