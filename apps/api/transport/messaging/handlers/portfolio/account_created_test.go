package portfolio

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/eventbus"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

// recordingPortfolioStore records the account id of the created projection.
type recordingPortfolioStore struct{ createdAccountID uuid.UUID }

func (s *recordingPortfolioStore) CreateAccount(_ context.Context, acc *portfolio.Account) error {
	s.createdAccountID = acc.AccountID
	return nil
}
func (s *recordingPortfolioStore) CreateTransaction(_ context.Context, _ *portfolio.Transaction) error {
	return nil
}
func (s *recordingPortfolioStore) CreateTransactions(_ context.Context, _ []*portfolio.Transaction) (int, error) {
	return 0, nil
}

func TestAccountCreatedCreatesPortfolioProjection(t *testing.T) {
	store := &recordingPortfolioStore{}
	commands := portfolio.NewCommands(store, marketdata.Queries{}, vendor.Queries{})
	handler := NewAccountCreatedHandler(commands, nil)

	accID := uuid.New()
	if err := handler.Handle(context.Background(), account.Created{AccID: accID}, eventbus.Metadata{}); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if store.createdAccountID != accID {
		t.Fatalf("expected portfolio projection for %s, got %s", accID, store.createdAccountID)
	}
}
