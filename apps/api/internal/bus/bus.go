package bus

import (
	"context"
	"fmt"
)

type Bus interface {
	Publish(ctx context.Context, env Envelope) error
	Subscribe(eventType string, h Handler) (Subscription, error)
	Close() error
}

type Handler func(context.Context, Envelope) error

type TypedHandler[T any] func(context.Context, Envelope, T) error

func DecodeHandler[T any](reg *CodecRegistry, h TypedHandler[T]) Handler {
	return func(ctx context.Context, env Envelope) error {
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
