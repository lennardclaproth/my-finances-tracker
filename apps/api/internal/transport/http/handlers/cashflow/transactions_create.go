package cashflow

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/transport/http"
)

// CreateManualCashflowTransactionRequest represents one manual cashflow transaction row.
type CreateManualCashflowTransactionRequest struct {
	Date        string `json:"date"`
	Amount      string `json:"amount"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Note        string `json:"note"`
	Tag         string `json:"tag"`
	Vendor      string `json:"vendor,omitempty"`
}

// CreateTransactionsRequest creates one or more manual cashflow transactions.
type CreateTransactionsRequest struct {
	AccountID    uuid.UUID                                `json:"account_id"`
	Transactions []CreateManualCashflowTransactionRequest `json:"transactions"`
}

func (r CreateTransactionsRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	if len(r.Transactions) == 0 {
		problems["transactions"] = "transactions is required"
		return len(problems) == 0, problems
	}
	for i, row := range r.Transactions {
		prefix := fmt.Sprintf("transactions[%d]", i)
		if strings.TrimSpace(row.Date) == "" {
			problems[prefix+".date"] = "date is required"
		}
		if strings.TrimSpace(row.Amount) == "" {
			problems[prefix+".amount"] = "amount is required"
		}
		if strings.TrimSpace(row.Type) == "" {
			problems[prefix+".type"] = "type is required"
		}
		if strings.TrimSpace(row.Description) == "" {
			problems[prefix+".description"] = "description is required"
		}
		if strings.TrimSpace(row.Note) == "" {
			problems[prefix+".note"] = "note is required"
		}
		if strings.TrimSpace(row.Tag) == "" {
			problems[prefix+".tag"] = "tag is required"
		}
	}
	return len(problems) == 0, problems
}

// CreateTransactionResponse represents a cashflow transaction in API responses.
type CreateTransactionResponse struct {
	ID          uuid.UUID `json:"id"`
	Description string    `json:"description"`
	Note        string    `json:"note"`
	Source      string    `json:"source"`
	AmountCents int64     `json:"amountCents"`
	Direction   string    `json:"direction"`
	Date        time.Time `json:"date"`
	Tag         string    `json:"tag"`
	Ignored     bool      `json:"ignored"`
}

// TransactionsResponse returns the manual cashflow create result.
type TransactionsResponse struct {
	CreatedCount int                         `json:"created_count"`
	Data         []CreateTransactionResponse `json:"data"`
}

// CreateTransactions creates manual cashflow transactions in bulk.
//
// @Summary     Create manual cashflow transactions
// @Description Creates one or more manual cashflow transactions for an account.
// @Tags        Transactions
// @Accept      application/json
// @Produce     application/json
// @Param       payload body CreateTransactionsRequest true "Manual cashflow transactions payload"
// @Success     201 {object} TransactionsResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     404 {object} map[string]string "Not found"
// @Failure     409 {object} map[string]string "Conflict"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/transactions/manual [post]
func CreateTransactions(
	log logging.Logger,
	commands *cashflow.Commands,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.JSONDecode[CreateTransactionsRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
			return
		}
		isValid, problems := req.isValid()
		if !isValid {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		transactions, err := toManualCashflowCreateTransactions(req.Transactions)
		if err != nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"transaction": err.Error()})
			return
		}

		created, err := commands.CreateMany(r.Context(), req.AccountID, nil, transactions)
		if err != nil {
			switch {
			case errors.Is(err, cashflow.ErrAccountNotFound):
				_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"account_id": err.Error()})
			case isManualCashflowValidationError(err):
				_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"transaction": err.Error()})
			case errors.Is(err, cashflow.ErrDuplicateTransaction):
				_ = httpx.JSONEncode(w, http.StatusConflict, map[string]string{"transaction": "duplicate transaction"})
			default:
				log.Error(r.Context(), "manual cashflow create: failed to create transactions", err)
				_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to create manual cashflow transactions"})
			}
			return
		}
		if created == nil {
			log.Error(r.Context(), "manual cashflow create: empty result", errors.New("empty result"))
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to create manual cashflow transactions"})
			return
		}

		data := make([]CreateTransactionResponse, 0, len(created))
		for _, tx := range created {
			data = append(data, CreateTransactionResponse{
				ID:          tx.ID,
				Description: tx.Description,
				Note:        tx.Note,
				Source:      tx.Source,
				AmountCents: int64(tx.AmountCents),
				Direction:   string(tx.Direction),
				Date:        tx.Date,
				Tag:         tx.Tag,
				Ignored:     tx.Ignored,
			})
		}

		_ = httpx.JSONEncode(w, http.StatusCreated, TransactionsResponse{
			CreatedCount: len(data),
			Data:         data,
		})
	})
}

func toManualCashflowCreateTransactions(
	transactions []CreateManualCashflowTransactionRequest,
) ([]cashflow.TransactionData, error) {
	out := make([]cashflow.TransactionData, 0, len(transactions))
	for _, tx := range transactions {
		row, err := cashflow.NewTransactionData(
			tx.Date,
			tx.Amount,
			tx.Type,
			tx.Description,
			tx.Note,
			tx.Tag,
			tx.Vendor,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, nil
}

func isManualCashflowValidationError(err error) bool {
	return errors.Is(err, cashflow.ErrTransactionsRequired) ||
		errors.Is(err, cashflow.ErrTransactionLimitExceeded) ||
		errors.Is(err, cashflow.ErrManualCashflowInvalidDate) ||
		errors.Is(err, cashflow.ErrManualCashflowInvalidAmount) ||
		errors.Is(err, cashflow.ErrManualCashflowInvalidType) ||
		errors.Is(err, cashflow.ErrManualCashflowDescriptionRequired) ||
		errors.Is(err, cashflow.ErrManualCashflowNoteRequired) ||
		errors.Is(err, cashflow.ErrManualCashflowTagRequired)
}
