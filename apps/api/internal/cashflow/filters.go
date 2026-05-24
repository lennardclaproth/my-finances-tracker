package cashflow

import "time"

type TransactionFilters struct {
	Query       string
	Description string
	Note        string
	Source      string

	Sort        *CashFlowDirection
	Tags        []string
	Untagged    *bool
	HideIgnored *bool

	From *time.Time
	To   *time.Time
}
