package bus

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type EnvelopeOption func(*Envelope)

type Message interface {
	MessageTopic() string
}

type Envelope struct {
	MessageID     uuid.UUID
	Topic         string
	Subject       string
	CorrelationID uuid.UUID
	CausationID   uuid.UUID
	OccurredAt    time.Time
	PublishedAt   time.Time
	Headers       map[string]string
	Codec         CodecType
	Payload       []byte
}

func WithSubject(s string) EnvelopeOption {
	return func(e *Envelope) { e.Subject = s }
}

func WithCorrelation(id uuid.UUID) EnvelopeOption {
	return func(e *Envelope) { e.CorrelationID = id }
}

func WithHeader(k, v string) EnvelopeOption {
	return func(e *Envelope) {
		if e.Headers == nil {
			e.Headers = map[string]string{}
		}
		e.Headers[k] = v
	}
}

func NewJSONEnvelope[T Message](payload T, opts ...EnvelopeOption) (Envelope, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, err
	}
	now := time.Now()

	env := Envelope{
		MessageID:   uuid.New(),
		Topic:       payload.MessageTopic(),
		OccurredAt:  now,
		PublishedAt: now,
		Codec:       CodecJSON,
		Headers:     map[string]string{},
		Payload:     b,
	}
	for _, opt := range opts {
		opt(&env)
	}
	return env, nil
}
