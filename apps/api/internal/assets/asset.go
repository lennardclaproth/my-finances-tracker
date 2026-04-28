package assets

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

var (
	ErrAssetNameEmpty      = errors.New("asset name cannot be empty")
)

// Asset is one concrete asset tracked inside a class.
type Asset struct {
	ID           uuid.UUID   `db:"id"`
	ClassID      uuid.UUID   `db:"class_id"`
	Class        Class       `db:"-"`
	Mutations    []Mutation  `db:"-"`
	AccountID    uuid.UUID   `db:"account_id"`
	Name         string      `db:"name"`
	CurrentWorth money.Price `db:"current_worth"`
	Archived     bool        `db:"archived"`
	CreatedAt    time.Time   `db:"created_at"`
	UpdatedAt    time.Time   `db:"updated_at"`
}

func NewAsset(
	accID, classID uuid.UUID,
	name string,
	initialWorth money.Price,
	note string,
) (*Asset, error) {
	n := strings.TrimSpace(name)
	if name == "" {
		return nil, ErrAssetNameEmpty
	}

	return &Asset{
		ID:           uuid.New(),
		ClassID:      classID,
		AccountID:    accID,
		Name:         n,
		CurrentWorth: initialWorth,
		Archived:     false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}
