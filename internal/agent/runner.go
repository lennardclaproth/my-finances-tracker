package agent

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/transaction"
)

type Runner struct {
	c                 *Client
	defaultTagAgentID uuid.UUID
}

func NewRunner(baseUrl string, defaultTagAgentID uuid.UUID) *Runner {
	client := NewClient(baseUrl)
	return &Runner{c: client, defaultTagAgentID: defaultTagAgentID}
}

func (r *Runner) RunTagAgent(ctx context.Context, tx *transaction.Transaction) error {
	msg := fmt.Sprintf(`
	Please tag the following transaction with the most appropriate tag based on its details. If no suitable tag is found, please use "unk". **Make sure to save the tag via the tool**\n
	## Transaction details: \n
	- ID: %s \n
	- Amount: %.2f \n
	- Date: %s \n
	- Description: %s \n
	- Note: %s \n
	`, tx.ID, float64(tx.AmountCents)/100, tx.Date.Format("2006-01-02"), tx.Description, tx.Note)
	return r.c.CallAgent(ctx, r.defaultTagAgentID, msg)
}
