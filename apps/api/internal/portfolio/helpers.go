package portfolio

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// getOrPromoteCycle returns the active cycle for "key".
// If tx now has an ISIN and we previously created a symbol-only cycle, we promote it:
// move activeByKey[symbolKey] -> activeByKey[key].
func getOrPromoteCycle(activeByKey map[string]*Position, key, symbolKey string, tx Transaction) *Position {
	pos := activeByKey[key]
	if pos != nil {
		return pos
	}
	// Promote symbol-only cycle to ISIN key once ISIN is known.
	if tx.ISIN != nil && symbolKey != "" {
		if promoted, ok := activeByKey[symbolKey]; ok {
			delete(activeByKey, symbolKey)
			activeByKey[key] = promoted
			return promoted
		}
	}
	return nil
}

// applyTxToCycle applies a transaction to a position cycle and assigns tx.PositionID.
// It may create a new cycle if needed (returns created=true in that case).
func applyTxToCycle(
	accID uuid.UUID,
	tx *Transaction,
	key, symbolKey string,
	aliases map[string]string,
	activeByKey map[string]*Position,
	pos *Position,
) (*Position, bool, error) {
	ensure := func() (*Position, bool, error) {
		if pos != nil {
			return pos, false, nil
		}
		// Start a new cycle with the best identity we can infer at this point.
		pos, err := newPositionCycle(accID, *tx, symbolKey, aliases)
		if err != nil {
			return nil, false, err
		}
		activeByKey[key] = pos
		return pos, true, nil
	}
	switch tx.Type {
	case TxBuy:
		pos, created, err := ensure()
		if err != nil {
			return nil, false, err
		}
		mergePositionIdentity(pos, *tx)
		if err := pos.ApplyTx(*tx); err != nil {
			return nil, false, fmt.Errorf("apply buy transaction to position: %w", err)
		}
		tx.PositionID = &pos.ID
		return pos, created, nil
	case TxSell:
		// No active cycle to reduce/close. ignore silently (current behavior)
		if pos == nil {
			return nil, false, nil
		}
		mergePositionIdentity(pos, *tx)
		if err := pos.ApplyTx(*tx); err != nil {
			return nil, false, fmt.Errorf("apply sell transaction to position: %w", err)
		}
		tx.PositionID = &pos.ID
		// Close the cycle when quantity reaches zero.
		if pos.Quantity == 0 {
			delete(activeByKey, key)
		}
		return pos, false, nil
	case TxDividend, TxTax, TxFee:
		// These affect cost basis; if we don't have a cycle yet, start one.
		pos, created, err := ensure()
		if err != nil {
			return nil, false, err
		}
		mergePositionIdentity(pos, *tx)
		if err := pos.ApplyTx(*tx); err != nil {
			return nil, false, fmt.Errorf("apply non-trade transaction to position: %w", err)
		}
		tx.PositionID = &pos.ID
		return pos, created, nil
	default:
		// Unknown tx type: ignore or error depending on your domain strictness.
		return pos, false, nil
	}
}

func canonicalPositionKey(tx Transaction, aliases map[string]string) (key, symbolKey string, err error) {
	isin := normalizedID(tx.ISIN)
	symbol := normalizedID(tx.Symbol)
	if isin == "" && symbol == "" {
		return "", "", ErrTransactionISINAndSymbolMissing
	}
	if symbol != "" {
		symbolKey = symbol
	}
	if isin != "" {
		key = isin
		if symbol != "" {
			aliases[symbol] = isin
		}
		return key, symbolKey, nil
	}
	if mapped, ok := aliases[symbol]; ok && mapped != "" {
		key = mapped
	} else {
		key = symbol
	}
	return key, symbolKey, nil
}

func newPositionCycle(accID uuid.UUID, tx Transaction, symbolKey string, aliases map[string]string) (*Position, error) {
	isin := tx.ISIN
	if isin == nil && symbolKey != "" {
		if mapped, ok := aliases[symbolKey]; ok && mapped != "" && mapped != symbolKey {
			v := mapped
			isin = &v
		}
	}
	return NewPosition(accID, isin, tx.Symbol, nil, tx.OccurredAt)
}

func mergePositionIdentity(pos *Position, tx Transaction) {
	if pos == nil {
		return
	}
	if pos.ISIN == nil && tx.ISIN != nil && normalizedID(tx.ISIN) != "" {
		v := normalizedID(tx.ISIN)
		pos.ISIN = &v
	}
	if pos.Symbol == nil && tx.Symbol != nil && normalizedID(tx.Symbol) != "" {
		v := normalizedID(tx.Symbol)
		pos.Symbol = &v
	}
}

func normalizedID(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

// getPositionFromMap tries to find an existing position for the given transaction based on ISIN or symbol.
// If no position exists, it creates a new one and adds it to the map.
func getPositionFromMap(accID uuid.UUID, positions *map[string]*Position, tx Transaction) (*Position, error) {
	txID, err := tx.GetID()
	if err != nil {
		return nil, err
	}
	// Fallback to symbol-based mapping if ISIN is not available
	if pos, ok := (*positions)[txID]; ok {
		return pos, nil
	}
	// No existing position found for this transaction, create a new one
	pos, err := NewPosition(accID, tx.ISIN, tx.Symbol, nil, tx.OccurredAt)
	if err != nil {
		return nil, err
	}
	id, err := pos.Identity()
	if err != nil {
		return nil, err
	}
	(*positions)[id] = pos
	return pos, nil
}

func parseManualType(raw string) (TransactionType, error) {
	normalized := TransactionType(strings.ToUpper(strings.TrimSpace(raw)))
	switch normalized {
	case TxBuy, TxSell, TxDividend, TxTax, TxFee, TxCash:
		return normalized, nil
	default:
		return "", ErrManualInvalidType
	}
}

func parseManualQuantity(txType TransactionType, raw *string) (float64, error) {
	if txType == TxBuy || txType == TxSell {
		if raw == nil || strings.TrimSpace(*raw) == "" {
			return 0, ErrManualQuantityRequired
		}
		quantity, err := parseDecimalString(*raw, ErrManualInvalidQuantity)
		if err != nil {
			return 0, err
		}
		if quantity <= 0 {
			return 0, ErrManualQuantityMustBePositive
		}
		return quantity, nil
	}

	if raw != nil && strings.TrimSpace(*raw) != "" {
		return 0, ErrManualQuantityForbidden
	}
	return 0, nil
}

var decimalSixPattern = regexp.MustCompile(`^-?\d+(\.\d{1,6})?$`)

func parseDecimalString(raw string, invalidErr error) (float64, error) {
	trimmed := strings.TrimSpace(raw)
	if !decimalSixPattern.MatchString(trimmed) {
		return 0, invalidErr
	}
	val, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, invalidErr
	}
	return val, nil
}
