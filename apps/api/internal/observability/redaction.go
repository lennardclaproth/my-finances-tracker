package observability

import (
	"fmt"

	"go.elastic.co/apm/v2"
)

var allowedObservabilityKeys = map[string]struct{}{
	"trace.id":       {},
	"transaction.id": {},
	"span.id":        {},
	"request_id":     {},
	"correlation_id": {},
	"operation":      {},
	"component":      {},
	"outcome":        {},
	"result":         {},
	"error_class":    {},
	"error":          {},
	"method":         {},
	"path":           {},
	"status":         {},
	"bytes":          {},
	"duration_ms":    {},
	"topic":          {},
	"message_id":     {},
	"causation_id":   {},
	"account_id":     {},
	"vendor_id":      {},
	"listing_id":     {},
	"import_id":      {},
	"upload_id":      {},
	"transaction_id": {},
	"row_number":     {},
	"symbol":         {},
	"source":         {},
	"date":           {},
	"provider":       {},
	"topics":         {},
	"keys_count":     {},
	"type":           {},
	"addr":           {},
	"swagger_url":    {},
	"worker_id":      {},
	"updated_count":  {},
	"stage":          {},
}

func IsAllowedKey(key string) bool {
	_, ok := allowedObservabilityKeys[key]
	return ok
}

func FilterFields(fields ...any) []any {
	if len(fields) == 0 {
		return nil
	}

	filtered := make([]any, 0, len(fields))
	for i := 0; i < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok || !IsAllowedKey(key) {
			continue
		}

		filtered = append(filtered, key)
		if i+1 < len(fields) {
			filtered = append(filtered, fields[i+1])
		} else {
			filtered = append(filtered, "")
		}
	}

	return filtered
}

func SetSafeTransactionLabels(tx *apm.Transaction, labels map[string]any) {
	if tx == nil || tx.TransactionData == nil || labels == nil {
		return
	}
	for key, value := range labels {
		if !IsAllowedKey(key) {
			continue
		}
		tx.Context.SetLabel(key, normalizeLabelValue(value))
	}
}

func normalizeLabelValue(value any) any {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return value
	}
}
