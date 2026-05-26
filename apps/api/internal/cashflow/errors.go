package cashflow

import "fmt"

// Transaction Errors
var (
	ErrTransactionsRequired              = fmt.Errorf("transactions are required")
	ErrTransactionLimitExceeded          = fmt.Errorf("transactions exceeds maximum batch size")
	ErrManualCashflowInvalidDate         = fmt.Errorf("date must be in YYYY-MM-DD format")
	ErrManualCashflowInvalidAmount       = fmt.Errorf("amount must be a positive decimal string with up to 6 decimals")
	ErrManualCashflowInvalidType         = fmt.Errorf("type must be one of: in, out, income, expense")
	ErrManualCashflowDescriptionRequired = fmt.Errorf("description is required")
	ErrManualCashflowNoteRequired        = fmt.Errorf("note is required")
	ErrManualCashflowTagRequired         = fmt.Errorf("tag is required")
)

// Account Errors
var (
	ErrAccountNotFound = fmt.Errorf("account not found")
)
