package eventbus

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type publishOptions struct {
	correlationID string
	causationID   string
	headers       map[string]string
}

type Metadata struct {
	MessageID     uuid.UUID
	Topic         string
	CorrelationID string
	CausationID   string
	OccurredAt    time.Time
	Headers       map[string]string
}

var ErrHandlerNil = fmt.Errorf("eventbus: handler cannot be nil")

type PublishOption func(*publishOptions)

// WithCorrelationID sets the correlation ID for the message to be
// published. It overrides any existing correlation ID in the context or
// message headers.
func WithCorrelationID(id string) PublishOption {
	return func(o *publishOptions) {
		o.correlationID = id
	}
}

// WithCausationID sets the causation ID for the message to be
// published. It overrides any existing causation ID in the context or
// message headers.
func WithCausationID(id string) PublishOption {
	return func(o *publishOptions) {
		o.causationID = id
	}
}

// WithHeader adds a custom header to the message to be published.
func WithHeader(key, value string) PublishOption {
	return func(o *publishOptions) {
		if o.headers == nil {
			o.headers = make(map[string]string)
		}
		o.headers[key] = value
	}
}

type Bus interface {
	Publish(ctx context.Context, topic string, payload any, opts ...PublishOption) error
	Subscribe(topic string, h HandlerFunc) (Subscription, error)
	Close() error
}

type Handler[T any] func(ctx context.Context, payload T, meta Metadata) error

type HandlerFunc func(ctx context.Context, envelope Envelope) error

type Subscription interface {
	Close() error
}

func Subscribe[T any](
	bus Bus,
	topic string,
	handler Handler[T],
) (Subscription, error) {
	if handler == nil {
		return nil, ErrHandlerNil
	}

	return bus.Subscribe(topic, func(ctx context.Context, envelope Envelope) error {
		payload, ok := envelope.Payload.(T)
		if !ok {
			return fmt.Errorf(
				"eventbus: handler for topic %q expected payload %T, got %T",
				topic,
				*new(T),
				envelope.Payload,
			)
		}

		return handler(ctx, payload, MetadataFromEnvelope(envelope))
	})
}

func MetadataFromEnvelope(e Envelope) Metadata {
	return Metadata{
		MessageID:     e.MessageID,
		Topic:         e.Topic,
		CorrelationID: e.CorrelationID,
		CausationID:   e.CausationID,
		OccurredAt:    e.OccurredAt,
		Headers:       cloneHeaders(e.Headers),
	}
}

func cloneHeaders(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return map[string]string{}
	}

	cloned := make(map[string]string, len(headers))
	for k, v := range headers {
		cloned[k] = v
	}

	return cloned
}
