package memorybus_test

import (
	"context"
	"testing"
	"time"

	"github.com/lennardclaproth/my-finances-tracker/internal/eventbus"
	memorybus "github.com/lennardclaproth/my-finances-tracker/internal/eventbus/memory"
)

type ping struct{ N int }

func TestMemoryBusPublishDeliversToSubscriber(t *testing.T) {
	bus := memorybus.NewMemoryBus(memorybus.WithWorkers(2), memorybus.WithQueueSize(8))
	defer bus.Close()

	got := make(chan eventbus.Envelope, 1)
	sub, err := bus.Subscribe("topic.ping", func(_ context.Context, env eventbus.Envelope) error {
		got <- env
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer sub.Close()

	if err := bus.Publish(context.Background(), "topic.ping", ping{N: 7}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	select {
	case env := <-got:
		p, ok := env.Payload.(ping)
		if !ok {
			t.Fatalf("unexpected payload type %T", env.Payload)
		}
		if p.N != 7 {
			t.Fatalf("unexpected payload %+v", p)
		}
		if env.CorrelationID != env.MessageID.String() {
			t.Fatalf("root event correlation should equal message id, got %q", env.CorrelationID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for delivery")
	}
}

// TestMemoryBusPropagatesCausationChain verifies that an event published from
// within a handler inherits the parent's correlation ID and is marked as caused
// by the parent message.
func TestMemoryBusPropagatesCausationChain(t *testing.T) {
	bus := memorybus.NewMemoryBus(memorybus.WithWorkers(2), memorybus.WithQueueSize(8))
	defer bus.Close()

	child := make(chan eventbus.Envelope, 1)
	childSub, err := bus.Subscribe("topic.child", func(_ context.Context, env eventbus.Envelope) error {
		child <- env
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe child: %v", err)
	}
	defer childSub.Close()

	parentSub, err := bus.Subscribe("topic.parent", func(ctx context.Context, _ eventbus.Envelope) error {
		return bus.Publish(ctx, "topic.child", ping{N: 1})
	})
	if err != nil {
		t.Fatalf("subscribe parent: %v", err)
	}
	defer parentSub.Close()

	if err := bus.Publish(context.Background(), "topic.parent", ping{N: 0}); err != nil {
		t.Fatalf("publish parent: %v", err)
	}

	select {
	case env := <-child:
		if env.CausationID == "" {
			t.Fatal("expected child to be caused by the parent message")
		}
		if env.CorrelationID == env.MessageID.String() {
			t.Fatal("child should inherit the parent's correlation, not correlate to itself")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for child delivery")
	}
}
