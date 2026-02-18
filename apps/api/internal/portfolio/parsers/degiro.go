package parsers

import (
	"encoding/csv"
	"fmt"
	"io"
	"iter"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
)

var tradePattern = regexp.MustCompile(`(?i)(koop|buy|verkoop|sell)\s+([0-9]+(?:[.,][0-9]+)?)\s*@\s*([0-9]+(?:[.,][0-9]+)?)`)

type DegiroParser struct {
	headerToColumn map[string]int
}

func NewDegiroParser() *DegiroParser {
	return &DegiroParser{
		headerToColumn: make(map[string]int),
	}
}

func (p *DegiroParser) ParseAll(rc io.ReadCloser) (iter.Seq2[int, portfolio.TransactionData], error) {
	reader := csv.NewReader(rc)
	reader.Comma = ','
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return nil, err
	}
	if err := p.parseHeader(header); err != nil {
		return nil, err
	}

	seq := func(yield func(int, portfolio.TransactionData) bool) {
		defer rc.Close()
		rowNumber := 1
		for {
			record, err := reader.Read()
			if err == io.EOF {
				return
			}
			if err != nil {
				rowNumber++
				continue
			}
			td, err := p.ParseRow(record)
			if err != nil {
				rowNumber++
				continue
			}
			if !yield(rowNumber, td) {
				return
			}
			rowNumber++
		}
	}
	return seq, nil
}

func (p *DegiroParser) parseHeader(headers []string) error {
	p.headerToColumn = make(map[string]int, len(headers))
	for i, h := range headers {
		p.headerToColumn[strings.TrimSpace(h)] = i
	}
	required := []string{"Date", "Value date", "Product", "ISIN", "Description", "Change", "Order Id"}
	for _, key := range required {
		if _, ok := p.headerToColumn[key]; !ok {
			return fmt.Errorf("missing required header: %s", key)
		}
	}
	return nil
}

func (p *DegiroParser) ParseRow(record []string) (portfolio.TransactionData, error) {
	dateRaw := p.value(record, "Date")
	if dateRaw == "" {
		return portfolio.TransactionData{}, fmt.Errorf("missing Date")
	}
	occurredAt, err := time.Parse("02-01-2006", dateRaw)
	if err != nil {
		return portfolio.TransactionData{}, fmt.Errorf("invalid Date: %w", err)
	}

	var valueDate *time.Time
	valueDateRaw := p.value(record, "Value date")
	if valueDateRaw != "" {
		parsed, err := time.Parse("02-01-2006", valueDateRaw)
		if err != nil {
			return portfolio.TransactionData{}, fmt.Errorf("invalid Value date: %w", err)
		}
		valueDate = &parsed
	}

	amount, err := parseDegiroAmount(p.value(record, "Change"))
	if err != nil {
		if fallback, fbErr := p.amountFromAdjacentColumn(record); fbErr == nil {
			amount = fallback
		} else {
			return portfolio.TransactionData{}, fmt.Errorf("invalid amount: %w", err)
		}
	}

	description := p.value(record, "Description")
	txType := classifyType(description)
	quantity, price := parseTrade(description)
	product := p.value(record, "Product")
	isin := p.value(record, "ISIN")
	return portfolio.TransactionData{
		Source:     "DEGIRO",
		OccurredAt: occurredAt,
		ValueDate:  valueDate,
		ISIN:       &isin,
		Symbol:     &product,
		Type:       txType,
		Quantity:   quantity,
		Price:      price,
		Amount:     amount,
		RawRef:     p.value(record, "Order Id"),
	}, nil
}

func classifyType(description string) portfolio.TransactionType {
	d := strings.ToLower(strings.TrimSpace(description))
	switch {
	case strings.Contains(d, "dividendbelasting"), strings.Contains(d, "withholding tax"), strings.Contains(d, "tax"):
		return portfolio.TxTax
	case strings.Contains(d, "dividend"):
		return portfolio.TxDividend
	case strings.Contains(d, "transactiekosten"), strings.Contains(d, "kosten"), strings.Contains(d, "fee"):
		return portfolio.TxFee
	case strings.Contains(d, "koop"), strings.Contains(d, "buy"):
		return portfolio.TxBuy
	case strings.Contains(d, "verkoop"), strings.Contains(d, "sell"):
		return portfolio.TxSell
	default:
		return portfolio.TxCash
	}
}

func parseTrade(description string) (quantity float64, price float64) {
	matches := tradePattern.FindStringSubmatch(description)
	if len(matches) != 4 {
		return 0, 0
	}
	q, err := parseEUFloat(matches[2])
	if err != nil {
		return 0, 0
	}
	p, err := parseEUFloat(matches[3])
	if err != nil {
		return 0, 0
	}
	return q, p
}

func (p *DegiroParser) value(record []string, key string) string {
	idx := p.headerToColumn[key]
	if idx < 0 || idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}

func parseDegiroAmount(raw string) (float64, error) {
	return parseEUFloat(raw)
}

func parseEUFloat(raw string) (float64, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, fmt.Errorf("empty amount")
	}
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")
	return strconv.ParseFloat(s, 64)
}

func (p *DegiroParser) amountFromAdjacentColumn(record []string) (float64, error) {
	idx := p.headerToColumn["Change"]
	if idx+1 >= len(record) {
		return 0, fmt.Errorf("no adjacent amount column")
	}
	return parseDegiroAmount(record[idx+1])
}
