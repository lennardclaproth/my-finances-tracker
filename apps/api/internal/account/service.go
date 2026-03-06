package account

import (
	"context"
	"fmt"

	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/bus"
)

type CreateService struct {
	creator   Creator
	publisher bus.Bus
}

func NewCreateService(creator Creator, publisher bus.Bus) *CreateService {
	return &CreateService{
		creator:   creator,
		publisher: publisher,
	}
}

func (s *CreateService) Create(ctx context.Context, acc *Account) error {
	if err := s.creator.Create(ctx, acc); err != nil {
		return err
	}
	if s.publisher == nil {
		return nil
	}
	env, err := bus.NewJSONEnvelopeFromContext(ctx, api.AccountCreated{AccID: acc.ID})
	if err != nil {
		return fmt.Errorf("account create service: encode account created event: %w", err)
	}
	if err := s.publisher.Publish(ctx, env); err != nil {
		return fmt.Errorf("account create service: publish account created event: %w", err)
	}
	return nil
}

var _ Creator = (*CreateService)(nil)
