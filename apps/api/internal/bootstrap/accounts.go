package bootstrap

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
)

var (
	defaultAccountID   = uuid.MustParse("64a3d50f-c71a-4015-9ee1-45572147ce56")
	defaultAccountName = "Lennard Claproth"
)

// Accounts bootstraps the default account through the account feature boundary.
func Accounts(ctx context.Context, commands *account.Commands, queries *account.Queries, logger logging.Logger) {
	if commands == nil || queries == nil {
		panic(fmt.Errorf("bootstrap accounts: account commands/queries are required"))
	}

	if _, err := queries.GetByID(ctx, defaultAccountID); err == nil {
		logger.Info(ctx, "account already exists, skipping bootstrap", "account_id", defaultAccountID.String())
		return
	} else if !errors.Is(err, account.ErrAccountNotFound) {
		panic(fmt.Errorf("bootstrap accounts: fetch by id %s: %w", defaultAccountID, err))
	}

	if _, err := commands.Create(ctx, &defaultAccountID, nil, defaultAccountName); err != nil {
		if errors.Is(err, account.ErrAccountAlreadyExists) {
			logger.Info(ctx, "account already exists by unique constraint, skipping bootstrap", "account_name", defaultAccountName)
			return
		}
		panic(fmt.Errorf("bootstrap accounts: create account %s: %w", defaultAccountName, err))
	}

	logger.Info(ctx, "bootstrapped account", "account_id", defaultAccountID.String(), "name", defaultAccountName)
}
