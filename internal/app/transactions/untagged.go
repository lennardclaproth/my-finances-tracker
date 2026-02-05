package transactions

import (
	"context"
	"fmt"

	"github.com/lennardclaproth/my-finances-tracker/api/contracts"
)

type UntaggedHandler struct {
	uf UntaggedTransactionFetcher
}

func NewUntaggedHandler(uf UntaggedTransactionFetcher) *UntaggedHandler {
	return &UntaggedHandler{uf: uf}
}

func (h *UntaggedHandler) Handle(ctx context.Context, page, pageSize int) ([]contracts.Transaction, error) {
	transactions, err := h.uf.FetchUntagged(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("fetching untagged transactions: %w", err)
	}
	res := []contracts.Transaction{}
	for _, tx := range transactions {
		res = append(res, contracts.Transaction{
			ID:          tx.ID,
			Description: tx.Description,
			Note:        tx.Note,
			Source:      tx.Source,
			AmountCents: tx.AmountCents,
			Date:        tx.Date,
		})
	}
	return res, err
}
