package eod

import (
	"context"
	"fmt"

	"github.com/lennardclaproth/my-finances-tracker/internal/files"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
)

// Processor parses and persists accepted EOD CSV imports.
type Processor struct {
	listings *marketdata.Queries
	files    *files.Queries
	parsers  importer.EODParserFactory
	commands *marketdata.Commands
}

// NewProcessor constructs the EOD CSV import processor.
func NewProcessor(
	listings *marketdata.Queries,
	files *files.Queries,
	parsers importer.EODParserFactory,
	commands *marketdata.Commands,
) *Processor {
	return &Processor{
		listings: listings,
		files:    files,
		parsers:  parsers,
		commands: commands,
	}
}

// Process parses the EOD CSV into a batch of datapoints and hands the whole batch to the
// marketdata commands, which persist it in a single bulk insert. Parsing is all-or-nothing:
// any row-level parse error fails the import.
func (p *Processor) Process(ctx context.Context, imp *importer.Import) (importer.ProcessResult, error) {
	listingID, err := imp.RequireListingID()
	if err != nil {
		return importer.ProcessResult{}, err
	}

	listing, err := p.listings.Listing(ctx, listingID)
	if err != nil {
		return importer.ProcessResult{}, fmt.Errorf("fetch listing: %w", err)
	}
	if listing == nil {
		return importer.ProcessResult{}, importer.ErrImportListingNotFound
	}

	parser, err := p.parsers(listing.Source)
	if err != nil {
		return importer.ProcessResult{}, err
	}
	rc, err := p.files.ReadCsv(ctx, imp.Path)
	if err != nil {
		return importer.ProcessResult{}, fmt.Errorf("open csv: %w", err)
	}
	defer rc.Close()

	parsed, err := parser.ParseAll(rc)
	if err != nil {
		return importer.ProcessResult{}, fmt.Errorf("parse csv: %w", err)
	}
	if len(parsed.RowErrors) > 0 {
		return importer.ProcessResult{
			TotalRows: parsed.TotalRows,
			Failed:    len(parsed.RowErrors),
		}, fmt.Errorf("parse csv: %d row errors", len(parsed.RowErrors))
	}

	inputs := make([]marketdata.EODInput, 0, len(parsed.Rows))
	for _, row := range parsed.Rows {
		inputs = append(inputs, marketdata.EODInput{
			Date:   row.Date,
			Open:   row.Open,
			Close:  row.Close,
			High:   row.High,
			Low:    row.Low,
			Volume: row.Volume,
		})
	}
	if len(inputs) == 0 {
		return importer.ProcessResult{TotalRows: parsed.TotalRows}, nil
	}

	result, err := p.commands.CreateEODs(ctx, listing.ID, listing.Symbol, inputs)
	if err != nil {
		return importer.ProcessResult{TotalRows: parsed.TotalRows, Failed: len(inputs)}, fmt.Errorf("import eod: %w", err)
	}
	return importer.ProcessResult{
		TotalRows:  parsed.TotalRows,
		Imported:   result.Imported,
		Duplicates: result.Duplicates,
	}, nil
}
