package cashflow

import (
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	"github.com/lennardclaproth/my-finances-tracker/internal/date"
)

// TransactionFilters contains transaction filters for bulk tag mutations.
type TransactionFilters struct {
	Q           string `json:"q,omitempty"`
	Description string `json:"description,omitempty"`
	Note        string `json:"note,omitempty"`
	Source      string `json:"source,omitempty"`
	Direction   string `json:"direction,omitempty"`
	Tags        string `json:"tags,omitempty"`
	Untagged    *bool  `json:"untagged,omitempty"`
	HideIgnored *bool  `json:"hide_ignored,omitempty"`
	From        string `json:"from,omitempty"`
	To          string `json:"to,omitempty"`
}

func (tf TransactionFilters) ToAppFilters() (cashflow.TransactionFilters, map[string]string) {
	problems := make(map[string]string)
	direction, directionErr := cashflow.ParseDirection(tf.Direction)
	if directionErr != nil {
		problems["filters.direction"] = directionErr.Error()
	}
	from, to, dateErr := date.ParseFromTo(tf.From, tf.To)
	if dateErr != nil {
		problems["filters.date_range"] = dateErr.Error()
	}

	return cashflow.TransactionFilters{
		Query:       tf.Q,
		Description: tf.Description,
		Note:        tf.Note,
		Source:      tf.Source,
		Direction:   direction,
		Tags:        cashflow.SplitTags(tf.Tags),
		Untagged:    tf.Untagged,
		HideIgnored: tf.HideIgnored,
		From:        from,
		To:          to,
	}, problems
}
