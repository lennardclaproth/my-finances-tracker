package cashflow

import "fmt"

// Transaction Errors
var (
	ErrTransactionsRequired     = fmt.Errorf("transactions are required")
	ErrTransactionLimitExceeded = fmt.Errorf("transactions exceeds maximum batch size")
)

// Account Errors
var (
	ErrAccountNotFound = fmt.Errorf("account not found")
)
