package transactions

import (
	"time"

	"github.com/lennardclaproth/my-finances-tracker/internal/domain"
)

type TransactionData struct {
	Description string
	Note        string
	Source      string
	Direction   domain.CashFlowDirection
	Amount      float64
	Date        time.Time
}