package eventbus

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Envelope is the transport-level wrapper for a published message. It carries
// the payload together with identity and correlation metadata.
type Envelope struct {
	MessageID     uuid.UUID
	Topic         string
	CorrelationID string
	CausationID   string
	OccurredAt    time.Time
	Headers       map[string]string
	Payload       any
}

// NewEnvelope builds an Envelope for a message to be published. It assigns a
// fresh message identity and timestamp, applies the publish options, and
// derives the correlation/causation chain: explicit options take precedence,
// otherwise the values are inherited from the parent message carried in ctx
// (see ContextWithMetadata). A root message with no parent correlates to its
// own ID.
func NewEnvelope(ctx context.Context, topic string, payload any, opts ...PublishOption) Envelope {
	options := publishOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	messageID := uuid.New()
	correlationID := options.correlationID
	causationID := options.causationID

	if parent, ok := MetadataFromContext(ctx); ok {
		if correlationID == "" {
			correlationID = parent.CorrelationID
		}
		if causationID == "" {
			causationID = parent.MessageID.String()
		}
	}
	if correlationID == "" {
		correlationID = messageID.String()
	}

	return Envelope{
		MessageID:     messageID,
		Topic:         topic,
		CorrelationID: correlationID,
		CausationID:   causationID,
		OccurredAt:    time.Now().UTC(),
		Headers:       cloneHeaders(options.headers),
		Payload:       payload,
	}
}
