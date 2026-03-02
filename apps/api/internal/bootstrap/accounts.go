package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
)

var (
	defaultAccountID   = uuid.MustParse("64a3d50f-c71a-4015-9ee1-45572147ce56")
	defaultAccountName = "Lennard Claproth"
)

func Accounts(ctx context.Context, creator account.Creator, fetcher account.Fetcher, logger logging.Logger) {
	if creator == nil || fetcher == nil {
		panic(fmt.Errorf("bootstrap accounts: creator/fetcher is required"))
	}

	if _, err := fetcher.FetchByID(ctx, defaultAccountID); err == nil {
		logger.Info(ctx, "account already exists, skipping bootstrap", "account_id", defaultAccountID.String())
		return
	} else if !errors.Is(err, account.ErrAccountNotFound) {
		panic(fmt.Errorf("bootstrap accounts: fetch by id %s: %w", defaultAccountID, err))
	}

	acc := &account.Account{
		ID:        defaultAccountID,
		Name:      defaultAccountName,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := creator.Create(ctx, acc); err != nil {
		if errors.Is(err, account.ErrAccountAlreadyExists) {
			logger.Info(ctx, "account already exists by unique constraint, skipping bootstrap", "account_name", defaultAccountName)
			return
		}
		panic(fmt.Errorf("bootstrap accounts: create account %s: %w", defaultAccountName, err))
	}

	logger.Info(ctx, "bootstrapped account", "account_id", defaultAccountID.String(), "name", defaultAccountName)
}
