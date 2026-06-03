package portfolio

import (
	"context"
	"fmt"

	"github.com/lennardclaproth/my-finances-tracker/internal/files"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
	portfoliodomain "github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

// Processor parses and persists accepted portfolio CSV imports.
type Processor struct {
	vendors  *vendor.Queries
	files    *files.Queries
	parsers  importer.PortfolioParserFactory
	commands *portfoliodomain.Commands
}

// NewProcessor constructs the portfolio CSV import processor.
func NewProcessor(
	vendors *vendor.Queries,
	files *files.Queries,
	parsers importer.PortfolioParserFactory,
	commands *portfoliodomain.Commands,
) *Processor {
	return &Processor{
		vendors:  vendors,
		files:    files,
		parsers:  parsers,
		commands: commands,
	}
}

// Process parses the portfolio CSV into a batch of transaction data and hands the whole
// batch to the portfolio commands, which persist it in a single bulk insert. It requires
// a brokerage vendor and stamps each row with its CSV row number for deduplication.
func (p *Processor) Process(ctx context.Context, imp *importer.Import) (importer.ProcessResult, error) {
	accountID, err := imp.RequireAccountID()
	if err != nil {
		return importer.ProcessResult{}, err
	}

	v, err := p.vendors.GetById(ctx, imp.VendorID)
	if err != nil {
		return importer.ProcessResult{}, fmt.Errorf("fetch vendor: %w", err)
	}
	if v.Type != vendor.VendorTypeBrokerage {
		return importer.ProcessResult{}, importer.ErrVendorNotBrokerage
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

	data := make([]portfoliodomain.TransactionData, 0)
	for rowNumber, row := range rows {
		row.RowNumber = rowNumber
		data = append(data, row)
	}
	if len(data) == 0 {
		return importer.ProcessResult{}, nil
	}

	result, err := p.commands.CreateMany(ctx, imp.ID, &accountID, data)
	if err != nil {
		return importer.ProcessResult{TotalRows: len(data), Failed: len(data)}, fmt.Errorf("import portfolio: %w", err)
	}
	return importer.ProcessResult{
		TotalRows:  len(data),
		Imported:   result.Imported,
		Duplicates: result.Duplicates,
	}, nil
}
