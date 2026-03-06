package parsers

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
)

type BrandNewDayParser struct{}

func NewBrandNewDayParser() *BrandNewDayParser {
	return &BrandNewDayParser{}
}

func (p *BrandNewDayParser) ParseAll(rc io.ReadCloser) (ParseResult, error) {
	raw, err := io.ReadAll(rc)
	_ = rc.Close()
	if err != nil {
		return ParseResult{}, err
	}

	text := decodeText(raw)
	if strings.TrimSpace(text) == "" {
		return ParseResult{}, fmt.Errorf("empty file")
	}

	reader := csv.NewReader(strings.NewReader(text))
	reader.Comma = detectDelimiter(text)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	header, err := reader.Read()
	if err != nil {
		return ParseResult{}, err
	}
	headerMap, err := parseHeader(header)
	if err != nil {
		return ParseResult{}, err
	}

	result := ParseResult{
		Rows:      make([]DailyRow, 0),
		RowErrors: make([]marketdata.DailyUploadRowError, 0),
		TotalRows: 0,
	}

	seenDates := map[string]struct{}{}
	rowNumber := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.TotalRows++
			result.RowErrors = append(result.RowErrors, marketdata.DailyUploadRowError{
				RowNumber: rowNumber,
				Reason:    "invalid CSV row",
			})
			rowNumber++
			continue
		}
		if isEmptyRecord(record) {
			rowNumber++
			continue
		}
		result.TotalRows++

		dateRaw := readValue(record, headerMap, "date")
		navRaw := readValue(record, headerMap, "nav")
		if strings.TrimSpace(navRaw) == "" || strings.TrimSpace(navRaw) == "-" {
			result.RowErrors = append(result.RowErrors, marketdata.DailyUploadRowError{
				RowNumber: rowNumber,
				Reason:    "NAV is required",
			})
			rowNumber++
			continue
		}

		parsedDate, err := time.Parse("2/1/2006", strings.TrimSpace(dateRaw))
		if err != nil {
			result.RowErrors = append(result.RowErrors, marketdata.DailyUploadRowError{
				RowNumber: rowNumber,
				Reason:    "invalid Date format, expected dd/mm/yyyy",
			})
			rowNumber++
			continue
		}

		dateKey := parsedDate.Format("2006-01-02")
		if _, exists := seenDates[dateKey]; exists {
			result.RowErrors = append(result.RowErrors, marketdata.DailyUploadRowError{
				RowNumber: rowNumber,
				Reason:    "duplicate date in file",
			})
			rowNumber++
			continue
		}

		nav, err := parseDecimalValue(navRaw)
		if err != nil {
			result.RowErrors = append(result.RowErrors, marketdata.DailyUploadRowError{
				RowNumber: rowNumber,
				Reason:    "invalid NAV value",
			})
			rowNumber++
			continue
		}
		if nav < 0 {
			result.RowErrors = append(result.RowErrors, marketdata.DailyUploadRowError{
				RowNumber: rowNumber,
				Reason:    "NAV cannot be negative",
			})
			rowNumber++
			continue
		}

		seenDates[dateKey] = struct{}{}
		result.Rows = append(result.Rows, DailyRow{
			RowNumber: rowNumber,
			Date:      parsedDate,
			Open:      nav,
			High:      nav,
			Low:       nav,
			Close:     nav,
			Volume:    0,
		})
		rowNumber++
	}

	return result, nil
}

func parseHeader(header []string) (map[string]int, error) {
	headerMap := make(map[string]int, len(header))
	for i, h := range header {
		headerMap[normalizeHeader(h)] = i
	}
	required := []string{"date", "nav", "ask", "bid", "dividend"}
	for _, key := range required {
		if _, ok := headerMap[key]; !ok {
			return nil, fmt.Errorf("missing required header: %s", key)
		}
	}
	return headerMap, nil
}

func normalizeHeader(raw string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimPrefix(raw, "\ufeff")))
}

func readValue(record []string, headerMap map[string]int, key string) string {
	idx, ok := headerMap[key]
	if !ok {
		return ""
	}
	if idx < 0 || idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}

func parseDecimalValue(raw string) (float64, error) {
	s := strings.TrimSpace(raw)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "'", "")
	if s == "" {
		return 0, fmt.Errorf("empty number")
	}

	lastComma := strings.LastIndex(s, ",")
	lastDot := strings.LastIndex(s, ".")
	switch {
	case lastComma >= 0 && lastDot >= 0:
		if lastComma > lastDot {
			s = strings.ReplaceAll(s, ".", "")
			s = strings.ReplaceAll(s, ",", ".")
		} else {
			s = strings.ReplaceAll(s, ",", "")
		}
	case lastComma >= 0:
		s = strings.ReplaceAll(s, ",", ".")
	}

	return strconv.ParseFloat(s, 64)
}

func detectDelimiter(text string) rune {
	firstLine := text
	if idx := strings.IndexAny(text, "\r\n"); idx >= 0 {
		firstLine = text[:idx]
	}
	if strings.Count(firstLine, "\t") >= strings.Count(firstLine, ";") {
		return '\t'
	}
	return ';'
}

func isEmptyRecord(record []string) bool {
	for _, v := range record {
		if strings.TrimSpace(v) != "" {
			return false
		}
	}
	return true
}

func decodeText(raw []byte) string {
	if utf8.Valid(raw) {
		return string(raw)
	}
	var b strings.Builder
	b.Grow(len(raw))
	for _, byt := range raw {
		b.WriteRune(rune(byt))
	}
	return b.String()
}
