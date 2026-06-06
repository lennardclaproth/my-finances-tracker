package account

import "errors"

var (
	// ErrAccountNotFound indicates that the requested account does not exist.
	ErrAccountNotFound = errors.New("account not found")
	// ErrAccountAlreadyExists indicates that an account with the same unique key already exists.
	ErrAccountAlreadyExists = errors.New("account already exists")
	// ErrAccountNameRequired indicates that account name validation failed.
	ErrAccountNameRequired = errors.New("account name is required")
)
