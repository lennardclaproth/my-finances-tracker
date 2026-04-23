package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

const (
	manualCashflowSourcePrefix  = "manual"
	manualCashflowBatchMaxSize  = 100
	manualCashflowImportMessage = "manual cashflow transaction creation"
)

var (
	// ErrManualCashflowAccountNotFound indicates the account does not exist.
	ErrManualCashflowAccountNotFound = fmt.Errorf("account not found")
	// ErrManualCashflowTransactionsRequired indicates no transactions were supplied.
	ErrManualCashflowTransactionsRequired = fmt.Errorf("transactions is required")
	// ErrManualCashflowTransactionLimitExceeded indicates the request is above the supported batch size.
	ErrManualCashflowTransactionLimitExceeded = fmt.Errorf("transactions exceeds maximum batch size of 100")
	// ErrManualCashflowInvalidDate indicates date format is invalid.
	ErrManualCashflowInvalidDate = fmt.Errorf("date must be in YYYY-MM-DD format")
	// ErrManualCashflowInvalidType indicates direction/type is invalid.
	ErrManualCashflowInvalidType = fmt.Errorf("type must be one of: in, out")
	// ErrManualCashflowInvalidAmount indicates amount format is invalid.
	ErrManualCashflowInvalidAmount = fmt.Errorf("amount must be a positive decimal string with up to 6 decimals")
	// ErrManualCashflowDescriptionRequired indicates description is missing.
	ErrManualCashflowDescriptionRequired = fmt.Errorf("description is required")
	// ErrManualCashflowNoteRequired indicates note is missing.
	ErrManualCashflowNoteRequired = fmt.Errorf("note is required")
	// ErrManualCashflowTagRequired indicates tag is missing.
	ErrManualCashflowTagRequired = fmt.Errorf("tag is required")
	// ErrManualCashflowVendorUnavailable indicates no import vendor could be resolved for manual rows.
	ErrManualCashflowVendorUnavailable = fmt.Errorf("manual import vendor unavailable")
)

var manualCashflowDecimalPattern = regexp.MustCompile(`^\d+(\.\d{1,6})?$`)

// ManualCashflowCreateTransactionInput describes one create row from transport.
type ManualCashflowCreateTransactionInput struct {
	Date        string
	Amount      string
	Type        string
	Description string
	Note        string
	Tag         string
	Vendor      string
}

// ManualCashflowCreateInput describes a bulk create request for one account.
type ManualCashflowCreateInput struct {
	AccountID    uuid.UUID
	Transactions []ManualCashflowCreateTransactionInput
}

// ManualCashflowCreateResult contains persisted rows.
type ManualCashflowCreateResult struct {
	Transactions []*cashflow.Transaction
}

type manualCashflowAccountFetcher interface {
	FetchByID(ctx context.Context, id uuid.UUID) (*account.Account, error)
}

type manualCashflowAccountProjectionCreator interface {
	Create(ctx context.Context, acc *cashflow.Account) error
}

type manualCashflowImportAccountCreator interface {
	Create(ctx context.Context, acc *importer.Account) error
}

type manualCashflowVendorStore interface {
	FetchByName(ctx context.Context, name vendor.VendorID) (*vendor.Vendor, error)
	ListActive(ctx context.Context) ([]*vendor.Vendor, error)
}

type manualCashflowImportCreator interface {
	Create(ctx context.Context, imp *importer.Import) error
}

type manualCashflowTransactionCreator interface {
	Create(ctx context.Context, tx *cashflow.Transaction) error
}

// ManualCreateService creates manual cashflow transactions in bulk.
type ManualCreateService struct {
	accounts     manualCashflowAccountFetcher
	cashflowAccs manualCashflowAccountProjectionCreator
	importAccs   manualCashflowImportAccountCreator
	vendors      manualCashflowVendorStore
	imports      manualCashflowImportCreator
	transactions manualCashflowTransactionCreator
}

// NewManualCreateService constructs a manual cashflow transaction create service.
func NewManualCreateService(
	accounts manualCashflowAccountFetcher,
	cashflowAccs manualCashflowAccountProjectionCreator,
	importAccs manualCashflowImportAccountCreator,
	vendors manualCashflowVendorStore,
	imports manualCashflowImportCreator,
	transactions manualCashflowTransactionCreator,
) *ManualCreateService {
	return &ManualCreateService{
		accounts:     accounts,
		cashflowAccs: cashflowAccs,
		importAccs:   importAccs,
		vendors:      vendors,
		imports:      imports,
		transactions: transactions,
	}
}

// CreateMany validates and persists manual cashflow transactions for one account.
func (s *ManualCreateService) CreateMany(ctx context.Context, input ManualCashflowCreateInput) (*ManualCashflowCreateResult, error) {
	if len(input.Transactions) == 0 {
		return nil, ErrManualCashflowTransactionsRequired
	}
	if len(input.Transactions) > manualCashflowBatchMaxSize {
		return nil, ErrManualCashflowTransactionLimitExceeded
	}

	if _, err := s.accounts.FetchByID(ctx, input.AccountID); err != nil {
		if errors.Is(err, account.ErrAccountNotFound) {
			return nil, ErrManualCashflowAccountNotFound
		}
		return nil, fmt.Errorf("manual cashflow create: fetch account: %w", err)
	}

	if err := s.cashflowAccs.Create(ctx, cashflow.NewAccount(input.AccountID)); err != nil {
		return nil, fmt.Errorf("manual cashflow create: ensure cashflow account projection: %w", err)
	}
	if err := s.importAccs.Create(ctx, importer.NewAccount(input.AccountID)); err != nil {
		return nil, fmt.Errorf("manual cashflow create: ensure import account projection: %w", err)
	}

	normalized, err := normalizeManualCashflowRows(input.Transactions)
	if err != nil {
		return nil, err
	}

	impVendor, err := s.resolveManualImportVendor(ctx)
	if err != nil {
		return nil, err
	}
	imp := importer.NewImport(*impVendor, manualCashflowImportPath(input.AccountID), input.AccountID)
	imp.MarkCompleted(0, len(normalized), len(normalized), 0)
	imp.StatusMsg = manualCashflowImportMessage

	if err := s.imports.Create(ctx, imp); err != nil {
		return nil, fmt.Errorf("manual cashflow create: create import row: %w", err)
	}

	accID := input.AccountID
	out := make([]*cashflow.Transaction, 0, len(normalized))
	for i, row := range normalized {
		rowNumber := manualCashflowRowNumber(i)
		tx, err := cashflow.NewTransaction(
			row.Description,
			row.Note,
			row.Source,
			row.Direction,
			row.Amount,
			row.Date,
			rowNumber,
			imp.ID,
			nil,
			&accID,
		)
		if err != nil {
			return nil, fmt.Errorf("manual cashflow create: build transaction: %w", err)
		}
		tx.Tag = row.Tag

		if err := s.transactions.Create(ctx, tx); err != nil {
			return nil, err
		}
		out = append(out, tx)
	}

	return &ManualCashflowCreateResult{Transactions: out}, nil
}

type normalizedManualCashflowRow struct {
	Date        time.Time
	Amount      float64
	Direction   cashflow.CashFlowDirection
	Description string
	Note        string
	Tag         string
	Source      string
}

func normalizeManualCashflowRows(rows []ManualCashflowCreateTransactionInput) ([]normalizedManualCashflowRow, error) {
	out := make([]normalizedManualCashflowRow, 0, len(rows))
	for idx, row := range rows {
		norm, err := normalizeManualCashflowRow(row)
		if err != nil {
			return nil, fmt.Errorf("%w (transaction index %d)", err, idx)
		}
		out = append(out, norm)
	}
	return out, nil
}

func normalizeManualCashflowRow(row ManualCashflowCreateTransactionInput) (normalizedManualCashflowRow, error) {
	parsedDate, err := time.Parse("2006-01-02", strings.TrimSpace(row.Date))
	if err != nil {
		return normalizedManualCashflowRow{}, ErrManualCashflowInvalidDate
	}

	direction, err := parseManualCashflowDirection(row.Type)
	if err != nil {
		return normalizedManualCashflowRow{}, err
	}

	amount, err := parseManualCashflowAmount(row.Amount)
	if err != nil {
		return normalizedManualCashflowRow{}, err
	}

	description := strings.TrimSpace(row.Description)
	if description == "" {
		return normalizedManualCashflowRow{}, ErrManualCashflowDescriptionRequired
	}
	note := strings.TrimSpace(row.Note)
	if note == "" {
		return normalizedManualCashflowRow{}, ErrManualCashflowNoteRequired
	}
	tag := strings.TrimSpace(row.Tag)
	if tag == "" {
		return normalizedManualCashflowRow{}, ErrManualCashflowTagRequired
	}

	return normalizedManualCashflowRow{
		Date:        parsedDate.UTC(),
		Amount:      amount,
		Direction:   direction,
		Description: description,
		Note:        note,
		Tag:         tag,
		Source:      manualCashflowSource(row.Vendor),
	}, nil
}

func parseManualCashflowDirection(raw string) (cashflow.CashFlowDirection, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	switch normalized {
	case "in", "income":
		return cashflow.CashIn, nil
	case "out", "expense":
		return cashflow.CashOut, nil
	default:
		return "", ErrManualCashflowInvalidType
	}
}

func parseManualCashflowAmount(raw string) (float64, error) {
	trimmed := strings.TrimSpace(raw)
	if !manualCashflowDecimalPattern.MatchString(trimmed) {
		return 0, ErrManualCashflowInvalidAmount
	}
	amount, err := strconv.ParseFloat(trimmed, 64)
	if err != nil || amount <= 0 {
		return 0, ErrManualCashflowInvalidAmount
	}
	return amount, nil
}

func manualCashflowSource(vendorName string) string {
	trimmedVendor := strings.TrimSpace(vendorName)
	if trimmedVendor == "" {
		return manualCashflowSourcePrefix
	}
	return manualCashflowSourcePrefix + ":" + trimmedVendor
}

func manualCashflowImportPath(accountID uuid.UUID) string {
	return fmt.Sprintf("manual://cashflow/%s/%d", accountID.String(), time.Now().UTC().UnixNano())
}

func manualCashflowRowNumber(idx int) int {
	const maxRowNumber = 2_147_483_647
	row := int(time.Now().UnixNano()%2_000_000_000) + idx + 1
	if row <= 0 {
		return idx + 1
	}
	if row > maxRowNumber {
		return (row % maxRowNumber) + 1
	}
	return row
}

func (s *ManualCreateService) resolveManualImportVendor(ctx context.Context) (*vendor.Vendor, error) {
	ingVendor, err := s.vendors.FetchByName(ctx, vendor.VendorING)
	switch {
	case err == nil && ingVendor != nil && ingVendor.Active:
		return ingVendor, nil
	case err != nil && !errors.Is(err, vendor.ErrVendorNotFound):
		return nil, fmt.Errorf("manual cashflow create: fetch ING vendor: %w", err)
	}

	active, err := s.vendors.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("manual cashflow create: list active vendors: %w", err)
	}
	if len(active) == 0 || active[0] == nil {
		return nil, ErrManualCashflowVendorUnavailable
	}
	return active[0], nil
}
