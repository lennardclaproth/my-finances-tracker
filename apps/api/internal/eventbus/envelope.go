package eventbus

import (
	"time"

	"github.com/google/uuid"
)

type Envelope struct {
	MessageID     uuid.UUID
	Topic         string
	CorrelationID string
	CausationID   string
	OccurredAt    time.Time
	Headers       map[string]string
	Payload       any
}
