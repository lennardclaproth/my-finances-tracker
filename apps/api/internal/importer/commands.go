package importer

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/eventbus"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

type importCommandStore interface {
	Create(ctx context.Context, imp *Import) error
	FetchByID(ctx context.Context, id uuid.UUID) (*Import, error)
	UpdateState(ctx context.Context, imp *Import) error
}

// CommandOption configures optional command dependencies.
type CommandOption func(*Commands)

// WithProcessors configures the type-specific processors used by Process.
func WithProcessors(cashflow, portfolio, eod Processor) CommandOption {
	return func(c *Commands) {
		c.cashflowProcessor = cashflow
		c.portfolioProcessor = portfolio
		c.eodProcessor = eod
	}
}

// Commands exposes CSV import write-side use cases.
type Commands struct {
	imports    importCommandStore
	files      ImportFileWriter
	remover    FileRemover
	vendors    vendor.Queries
	accounts   account.Queries
	marketdata marketdata.Queries
	bus        eventbus.Bus

	cashflowProcessor  Processor
	portfolioProcessor Processor
	eodProcessor       Processor
}

// NewCommands creates CSV import write-side use cases.
func NewCommands(
	imports importCommandStore,
	files ImportFileWriter,
	remover FileRemover,
	vendors vendor.Queries,
	accounts account.Queries,
	marketdata marketdata.Queries,
	bus eventbus.Bus,
	options ...CommandOption,
) *Commands {
	commands := &Commands{
		imports:    imports,
		files:      files,
		remover:    remover,
		vendors:    vendors,
		accounts:   accounts,
		marketdata: marketdata,
		bus:        bus,
	}
	for _, option := range options {
		option(commands)
	}
	return commands
}

// CashflowCSVImportCommand contains the inputs needed to accept a cashflow CSV import.
type CashflowCSVImportCommand struct {
	File      io.Reader
	VendorID  uuid.UUID
	AccountID uuid.UUID
}

// PortfolioCSVImportCommand contains the inputs needed to accept a portfolio CSV import.
type PortfolioCSVImportCommand struct {
	File      io.Reader
	VendorID  uuid.UUID
	AccountID uuid.UUID
}

// EODCSVImportCommand contains the inputs needed to accept a daily market-data CSV import.
type EODCSVImportCommand struct {
	File      io.Reader
	ListingID uuid.UUID
}

// ImportCashflowCSV validates and persists a pending cashflow CSV import.
func (c *Commands) ImportCashflowCSV(ctx context.Context, cmd CashflowCSVImportCommand) (uuid.UUID, error) {
	if cmd.File == nil {
		return uuid.Nil, ErrImportFileRequired
	}
	if cmd.VendorID == uuid.Nil {
		return uuid.Nil, ErrVendorIDRequired
	}
	if err := c.validateAccount(ctx, cmd.AccountID); err != nil {
		return uuid.Nil, err
	}
	v, err := c.validateVendor(ctx, cmd.VendorID)
	if err != nil {
		return uuid.Nil, err
	}

	return c.accept(ctx, cmd.File, ImportTypeCashflow,
		WithImportVendorID(v.ID),
		WithImportAccountID(cmd.AccountID),
		WithImportSource(string(v.Name)),
	)
}

// ImportPortfolioCSV validates and persists a pending portfolio CSV import.
func (c *Commands) ImportPortfolioCSV(ctx context.Context, cmd PortfolioCSVImportCommand) (uuid.UUID, error) {
	if cmd.File == nil {
		return uuid.Nil, ErrImportFileRequired
	}
	if cmd.VendorID == uuid.Nil {
		return uuid.Nil, ErrVendorIDRequired
	}
	if err := c.validateAccount(ctx, cmd.AccountID); err != nil {
		return uuid.Nil, err
	}
	v, err := c.validateVendor(ctx, cmd.VendorID)
	if err != nil {
		return uuid.Nil, err
	}
	if v.Type != vendor.VendorTypeBrokerage {
		return uuid.Nil, ErrVendorNotBrokerage
	}

	return c.accept(ctx, cmd.File, ImportTypePortfolio,
		WithImportVendorID(v.ID),
		WithImportAccountID(cmd.AccountID),
		WithImportSource(string(v.Name)),
	)
}

// ImportEODCSV validates and persists a pending daily market-data CSV import.
func (c *Commands) ImportEODCSV(ctx context.Context, cmd EODCSVImportCommand) (uuid.UUID, error) {
	if cmd.File == nil {
		return uuid.Nil, ErrImportFileRequired
	}
	listing, err := c.validateManualListing(ctx, cmd.ListingID)
	if err != nil {
		return uuid.Nil, err
	}

	return c.accept(ctx, cmd.File, ImportTypeEOD,
		WithImportListingID(listing.ID),
		WithImportSource(string(listing.Source)),
	)
}

// Process loads one import, updates its lifecycle state, and delegates type-specific work to a processor.
func (c *Commands) Process(ctx context.Context, importID uuid.UUID) error {
	if importID == uuid.Nil {
		return fmt.Errorf("import id is required")
	}
	if c.imports == nil {
		return fmt.Errorf("import commands are not configured")
	}

	imp, err := c.imports.FetchByID(ctx, importID)
	if err != nil {
		if errors.Is(err, ErrNoImportsPending) {
			return nil
		}
		return err
	}
	if imp == nil || imp.Status != ImportStatusPending {
		return nil
	}

	processor := c.processorFor(imp.Type)
	imp.MarkProcessing()
	if err := c.imports.UpdateState(ctx, imp); err != nil {
		return err
	}

	if processor == nil {
		reason := fmt.Errorf("%w: %s", ErrImportProcessorUnavailable, imp.Type)
		return c.markFailed(ctx, imp, ProcessResult{}, reason)
	}

	result, err := processor.Process(ctx, imp)
	if err != nil {
		return c.markFailed(ctx, imp, result, err)
	}

	imp.MarkCompleted(result.Duplicates, result.TotalRows, result.Imported, result.Failed)
	if err := c.imports.UpdateState(ctx, imp); err != nil {
		return err
	}
	c.publishCompleted(ctx, imp)
	return nil
}

func (c *Commands) validateAccount(ctx context.Context, accountID uuid.UUID) error {
	if accountID == uuid.Nil {
		return ErrAccountIDRequired
	}
	exists, err := c.accounts.Exists(ctx, accountID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrAccountNotExists
	}
	return nil
}

func (c *Commands) validateVendor(ctx context.Context, vendorID uuid.UUID) (*vendor.Vendor, error) {
	v, err := c.vendors.GetById(ctx, vendorID)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, vendor.ErrVendorNotFound
	}
	if v.ImportDisabled {
		return nil, ErrVendorImportDisabled
	}
	return v, nil
}

func (c *Commands) validateManualListing(ctx context.Context, listingID uuid.UUID) (*marketdata.Listing, error) {
	listing, err := c.marketdata.Listing(ctx, listingID)
	if err != nil {
		return nil, err
	}
	if listing == nil {
		return nil, ErrImportListingNotFound
	}
	if !listing.Active {
		return nil, ErrImportListingInactive
	}

	providerName, err := marketdata.ProviderNameFromSource(listing.Source)
	if err != nil {
		return nil, ErrImportProviderUnavailable
	}
	provider, err := c.marketdata.Provider(ctx, providerName)
	if err != nil {
		if errors.Is(err, marketdata.ErrProviderNotFound) {
			return nil, ErrImportProviderUnavailable
		}
		return nil, err
	}
	if provider == nil {
		return nil, ErrImportProviderUnavailable
	}
	if !provider.IsManualIngestion() {
		return nil, ErrImportProviderNotManual
	}
	return listing, nil
}

func (c *Commands) accept(ctx context.Context, file io.Reader, importType ImportType, options ...ImportOption) (uuid.UUID, error) {
	if c.imports == nil || c.files == nil {
		return uuid.Nil, fmt.Errorf("import commands are not configured")
	}
	path, err := c.files.WriteCsv(file)
	if err != nil {
		return uuid.Nil, fmt.Errorf("accept import: write csv: %w", err)
	}

	imp, err := NewTypedImport(importType, path, options...)
	if err != nil {
		c.cleanup(path)
		return uuid.Nil, err
	}
	if err := c.imports.Create(ctx, imp); err != nil {
		c.cleanup(path)
		return uuid.Nil, fmt.Errorf("accept import: create import: %w", err)
	}
	c.publishAccepted(ctx, imp)
	return imp.ID, nil
}

func (c *Commands) cleanup(path string) {
	if c.remover == nil {
		return
	}
	_ = c.remover.Remove(path)
}

func (c *Commands) publishAccepted(ctx context.Context, imp *Import) {
	if c.bus == nil || imp == nil {
		return
	}
	_ = c.bus.Publish(ctx, TopicAccepted, Accepted{
		ImportID: imp.ID,
		Type:     imp.Type,
	})
}

func (c *Commands) processorFor(importType ImportType) Processor {
	switch importType {
	case "", ImportTypeCashflow:
		return c.cashflowProcessor
	case ImportTypePortfolio:
		return c.portfolioProcessor
	case ImportTypeEOD:
		return c.eodProcessor
	default:
		return nil
	}
}

func (c *Commands) markFailed(ctx context.Context, imp *Import, result ProcessResult, reason error) error {
	imp.Duplicates = result.Duplicates
	imp.TotalRows = result.TotalRows
	imp.Imported = result.Imported
	imp.Failed = result.Failed
	imp.MarkFailed(reason.Error())
	if err := c.imports.UpdateState(ctx, imp); err != nil {
		return err
	}
	c.publishFailed(ctx, imp, reason)
	return reason
}

func (c *Commands) publishCompleted(ctx context.Context, imp *Import) {
	if c.bus == nil || imp == nil {
		return
	}
	_ = c.bus.Publish(ctx, TopicCompleted, Completed{
		ImportID:  imp.ID,
		Type:      imp.Type,
		AccountID: imp.AccountID,
		ListingID: imp.ListingID,
	})
}

func (c *Commands) publishFailed(ctx context.Context, imp *Import, reason error) {
	if c.bus == nil || imp == nil {
		return
	}
	_ = c.bus.Publish(ctx, TopicFailed, Failed{
		ImportID:  imp.ID,
		Type:      imp.Type,
		AccountID: imp.AccountID,
		ListingID: imp.ListingID,
		Reason:    reason.Error(),
	})
}
