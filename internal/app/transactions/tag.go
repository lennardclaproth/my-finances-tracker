package transactions

import (
	"context"

	"github.com/google/uuid"
)

type TagHandler struct {
	tagger TransactionTagger
}

func NewTagHandler(tagger TransactionTagger) *TagHandler {
	return &TagHandler{tagger: tagger}
}

func (h *TagHandler) Handle(ctx context.Context, id uuid.UUID, tag string) error {
	if err := h.tagger.Tag(ctx, id, tag); err != nil {
		return err
	}
	return nil
}
