package bus

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/observability"
	"go.elastic.co/apm/v2"
)

type Bus interface {
	Publish(ctx context.Context, env Envelope) error
	Subscribe(eventType string, h Handler) (Subscription, error)
	Close() error
}

type Handler func(context.Context, Envelope) error

type TypedHandler[T any] func(context.Context, Envelope, T) error

func DecodeHandler[T any](reg *CodecRegistry, h TypedHandler[T]) Handler {
	return func(ctx context.Context, env Envelope) (err error) {
		ctx = observability.ContextWithPropagationHeaders(ctx, env.Headers)
		if observability.CorrelationIDFromContext(ctx) == "" && env.CorrelationID != uuid.Nil {
			ctx = observability.ContextWithCorrelationID(ctx, env.CorrelationID.String())
		}
		if observability.RequestIDFromContext(ctx) == "" && env.MessageID != uuid.Nil {
			ctx = observability.ContextWithRequestID(ctx, env.MessageID.String())
		}

		tx, txCtx, txErr := observability.StartTransactionFromHeaders(
			ctx,
			observability.BusConsumeOperation(env.Topic),
			"messaging",
			env.Headers,
		)
		if txErr != nil {
			apm.CaptureError(ctx, txErr).Send()
		}
		ctx = txCtx
		tx.Result = "success"
		tx.Outcome = "success"
		observability.SetSafeTransactionLabels(tx, map[string]any{
			"operation":      observability.BusConsumeOperation(env.Topic),
			"component":      "bus",
			"topic":          env.Topic,
			"message_id":     env.MessageID.String(),
			"correlation_id": env.CorrelationID.String(),
			"causation_id":   env.CausationID.String(),
		})
		defer func() {
			if err != nil {
				tx.Result = "error"
				tx.Outcome = "failure"
				apm.CaptureError(ctx, err).Send()
			}
			tx.End()
		}()

		codec, err := reg.Get(env.Codec)
		if err != nil {
			return err
		}

		var msg T
		if err := codec.Decode(env.Payload, &msg); err != nil {
			return fmt.Errorf("decode %s as %T: %w", env.Codec, msg, err)
		}

		return h(ctx, env, msg)
	}
}

type Subscription interface {
	Close() error
}
