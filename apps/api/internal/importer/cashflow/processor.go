package cashflow

import (
	"context"
	"fmt"

	cashflowdomain "github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	"github.com/lennardclaproth/my-finances-tracker/internal/files"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

// Processor parses and persists accepted cashflow CSV imports.
type Processor struct {
	vendors  *vendor.Queries
	files    *files.Queries
	parsers  importer.CashflowParserFactory
	commands *cashflowdomain.Commands
}

// NewProcessor constructs the cashflow CSV import processor.
func NewProcessor(
	vendors *vendor.Queries,
	files *files.Queries,
	parsers importer.CashflowParserFactory,
	commands *cashflowdomain.Commands,
) *Processor {
	return &Processor{
		vendors:  vendors,
		files:    files,
		parsers:  parsers,
		commands: commands,
	}
}

// Process parses the cashflow CSV into a batch of transaction data and hands the whole
// batch to the cashflow commands, which persist it in a single bulk insert. Rows are
// stamped with the vendor as their source and the CSV row number for deduplication.
func (p *Processor) Process(ctx context.Context, imp *importer.Import) (importer.ProcessResult, error) {
	accountID, err := imp.RequireAccountID()
	if err != nil {
		return importer.ProcessResult{}, err
	}

	v, err := p.vendors.GetById(ctx, imp.VendorID)
	if err != nil {
		return importer.ProcessResult{}, fmt.Errorf("fetch vendor: %w", err)
	}
	parser, err := p.parsers(v.Name)
	if err != nil {
		return importer.ProcessResult{}, err
	}
	rc, err := p.files.ReadCsv(ctx, imp.Path)
	if err != nil {
		return importer.ProcessResult{}, fmt.Errorf("open csv: %w", err)
	}
	defer rc.Close()

	rows, err := parser.ParseAll(rc)
	if err != nil {
		return importer.ProcessResult{}, fmt.Errorf("parse csv: %w", err)
	}

	source := string(v.Name)
	data := make([]cashflowdomain.TransactionData, 0)
	for rowNumber, row := range rows {
		row.Source = source
		row.RowNumber = rowNumber
		data = append(data, row)
	}
	if len(data) == 0 {
		return importer.ProcessResult{}, nil
	}

	result, err := p.commands.CreateMany(ctx, accountID, &imp.ID, data)
	if err != nil {
		return importer.ProcessResult{TotalRows: len(data), Failed: len(data)}, fmt.Errorf("import cashflow: %w", err)
	}
	return importer.ProcessResult{
		TotalRows:  len(data),
		Imported:   result.Imported,
		Duplicates: result.Duplicates,
	}, nil
}
