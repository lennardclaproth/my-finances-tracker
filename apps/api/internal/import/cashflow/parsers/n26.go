package parsers

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"iter"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

const (
	n26FieldBookingDate      = "booking_date"
	n26FieldValueDate        = "value_date"
	n26FieldPartnerName      = "partner_name"
	n26FieldPartnerIBAN      = "partner_iban"
	n26FieldType             = "type"
	n26FieldPaymentRef       = "payment_reference"
	n26FieldAmountEUR        = "amount_eur"
	n26FieldOriginalAmount   = "original_amount"
	n26FieldOriginalCurrency = "original_currency"
	n26FieldExchangeRate     = "exchange_rate"
)

var n26HeaderAliasToField = map[string]string{
	"bookingdate":        n26FieldBookingDate,
	"buchungsdatum":      n26FieldBookingDate,
	"buchungstag":        n26FieldBookingDate,
	"valuedate":          n26FieldValueDate,
	"wertstellungsdatum": n26FieldValueDate,
	"wertstellung":       n26FieldValueDate,
	"partnername":        n26FieldPartnerName,
	"empfaengername":     n26FieldPartnerName,
	"beguenstigter":      n26FieldPartnerName,
	"partneriban":        n26FieldPartnerIBAN,
	"empfaengeriban":     n26FieldPartnerIBAN,
	"beguenstigteriban":  n26FieldPartnerIBAN,
	"type":               n26FieldType,
	"typ":                n26FieldType,
	"paymentreference":   n26FieldPaymentRef,
	"reference":          n26FieldPaymentRef,
	"referenz":           n26FieldPaymentRef,
	"verwendungszweck":   n26FieldPaymentRef,
	"amounteur":          n26FieldAmountEUR,
	"betrageur":          n26FieldAmountEUR,
	"betrageuro":         n26FieldAmountEUR,
	"originalamount":     n26FieldOriginalAmount,
	"originalbetrag":     n26FieldOriginalAmount,
	"originalcurrency":   n26FieldOriginalCurrency,
	"originalwaehrung":   n26FieldOriginalCurrency,
	"exchangerate":       n26FieldExchangeRate,
	"wechselkurs":        n26FieldExchangeRate,
}

type N26Parser struct {
	headerToColumn map[string]int
	accountType    *cashflow.AccountType
}

func NewN26Parser() *N26Parser {
	acct := cashflow.AccountTypeChecking
	return &N26Parser{
		headerToColumn: make(map[string]int),
		accountType:    &acct,
	}
}

func (p *N26Parser) ParseAll(rc io.ReadCloser) (iter.Seq2[int, cashflow.TransactionData], error) {
	raw, err := io.ReadAll(rc)
	closeErr := rc.Close()
	if err != nil {
		if closeErr != nil {
			return nil, errors.Join(err, closeErr)
		}
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}

	decoded := decodeN26Text(raw)
	csvReader := csv.NewReader(strings.NewReader(decoded))
	csvReader.Comma = detectN26Delimiter(decoded)
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true

	header, err := csvReader.Read()
	if err != nil {
		return nil, err
	}
	if err := p.parseHeader(header); err != nil {
		return nil, err
	}

	seq := func(yield func(int, cashflow.TransactionData) bool) {
		rowNumber := 1 // first data row after header
		for {
			record, err := csvReader.Read()
			if err == io.EOF {
				return
			}
			if err != nil {
				rowNumber++
				continue
			}
			td, err := p.ParseRow(record, rowNumber, uuid.Nil)
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

func (p *N26Parser) parseHeader(headers []string) error {
	p.headerToColumn = make(map[string]int, len(headers))

	for i, h := range headers {
		alias := normalizeN26Header(h)
		field, ok := n26HeaderAliasToField[alias]
		if !ok {
			continue
		}
		if _, exists := p.headerToColumn[field]; !exists {
			p.headerToColumn[field] = i
		}
	}

	if _, ok := p.headerToColumn[n26FieldAmountEUR]; !ok {
		return fmt.Errorf("missing required header: Amount (EUR)")
	}
	if _, hasBooking := p.headerToColumn[n26FieldBookingDate]; !hasBooking {
		if _, hasValue := p.headerToColumn[n26FieldValueDate]; !hasValue {
			return fmt.Errorf("missing required header: Booking Date")
		}
	}

	return nil
}

func (p *N26Parser) ParseRow(record []string, rowNumber int, importID uuid.UUID) (cashflow.TransactionData, error) {
	if len(p.headerToColumn) == 0 {
		return cashflow.TransactionData{}, fmt.Errorf("header map not initialized")
	}

	dateRaw := p.value(record, n26FieldBookingDate)
	if dateRaw == "" {
		dateRaw = p.value(record, n26FieldValueDate)
	}
	if dateRaw == "" {
		return cashflow.TransactionData{}, fmt.Errorf("missing date")
	}
	parsedDate, err := parseN26Date(dateRaw)
	if err != nil {
		return cashflow.TransactionData{}, fmt.Errorf("invalid date: %w", err)
	}

	amount, err := parseN26Amount(p.value(record, n26FieldAmountEUR))
	if err != nil {
		return cashflow.TransactionData{}, fmt.Errorf("invalid amount: %w", err)
	}

	direction := cashflow.CashIn
	if amount < 0 {
		direction = cashflow.CashOut
	}

	paymentRef := p.value(record, n26FieldPaymentRef)
	description := p.value(record, n26FieldPartnerName)
	if description == "" {
		description = paymentRef
	}
	if description == "" {
		description = "N26 transaction"
	}

	var noteParts []string
	if paymentRef != "" {
		noteParts = append(noteParts, "Reference:"+paymentRef)
	}
	if partnerIBAN := p.value(record, n26FieldPartnerIBAN); partnerIBAN != "" {
		noteParts = append(noteParts, "IBAN:"+partnerIBAN)
	}
	if txType := p.value(record, n26FieldType); txType != "" {
		noteParts = append(noteParts, "Type:"+txType)
	}
	if originalAmount := p.value(record, n26FieldOriginalAmount); originalAmount != "" {
		noteParts = append(noteParts, "OriginalAmount:"+originalAmount)
	}
	if originalCurrency := p.value(record, n26FieldOriginalCurrency); originalCurrency != "" {
		noteParts = append(noteParts, "OriginalCurrency:"+originalCurrency)
	}
	if exchangeRate := p.value(record, n26FieldExchangeRate); exchangeRate != "" {
		noteParts = append(noteParts, "ExchangeRate:"+exchangeRate)
	}
	price, _ := money.NewPrice(amount)
	return cashflow.TransactionData{
		Description: description,
		Note:        strings.Join(noteParts, " | "),
		Source:      "N26",
		Direction:   direction,
		Amount:      price,
		Date:        parsedDate,
		AccountType: p.accountType,
	}, nil
}

func (p *N26Parser) value(record []string, key string) string {
	idx, ok := p.headerToColumn[key]
	if !ok || idx < 0 || idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}

func parseN26Date(raw string) (time.Time, error) {
	clean := strings.TrimSpace(raw)
	layouts := []string{
		"2006-01-02",
		"02-01-2006",
		"02.01.2006",
		"2006/01/02",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, clean)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported date format: %s", raw)
}

func parseN26Amount(raw string) (float64, error) {
	normalized := normalizeN26Amount(raw)
	if normalized == "" {
		return 0, fmt.Errorf("empty amount")
	}
	return strconv.ParseFloat(normalized, 64)
}

func normalizeN26Amount(raw string) string {
	clean := strings.TrimSpace(raw)
	clean = strings.ReplaceAll(clean, "EUR", "")
	clean = strings.ReplaceAll(clean, "eur", "")
	clean = strings.ReplaceAll(clean, string(rune(0x20AC)), "") // euro sign
	clean = strings.ReplaceAll(clean, "'", "")
	clean = strings.Map(func(r rune) rune {
		if r == rune(0x2212) { // unicode minus
			return '-'
		}
		if r == rune(0x00A0) { // non-breaking space
			return -1
		}
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			return -1
		}
		return r
	}, clean)

	lastComma := strings.LastIndex(clean, ",")
	lastDot := strings.LastIndex(clean, ".")

	switch {
	case lastComma >= 0 && lastDot >= 0:
		if lastComma > lastDot {
			// 1.234,56
			clean = strings.ReplaceAll(clean, ".", "")
			clean = strings.ReplaceAll(clean, ",", ".")
		} else {
			// 1,234.56
			clean = strings.ReplaceAll(clean, ",", "")
		}
	case lastComma >= 0:
		// 1234,56
		clean = strings.ReplaceAll(clean, ",", ".")
	}

	return clean
}

func normalizeN26Header(raw string) string {
	s := strings.TrimSpace(strings.TrimPrefix(raw, "\ufeff"))
	s = strings.ToLower(s)

	var b strings.Builder
	for _, r := range s {
		switch r {
		case rune(0x00E4):
			b.WriteString("ae")
			continue
		case rune(0x00F6):
			b.WriteString("oe")
			continue
		case rune(0x00FC):
			b.WriteString("ue")
			continue
		case rune(0x00DF):
			b.WriteString("ss")
			continue
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func detectN26Delimiter(csvText string) rune {
	firstLine := csvText
	if idx := strings.IndexAny(csvText, "\r\n"); idx >= 0 {
		firstLine = csvText[:idx]
	}
	if strings.Count(firstLine, ";") > strings.Count(firstLine, ",") {
		return ';'
	}
	return ','
}

func decodeN26Text(raw []byte) string {
	if utf8.Valid(raw) {
		return string(raw)
	}
	if decoded, ok := decodeWindows1252(raw); ok {
		return decoded
	}
	return decodeLatin1(raw)
}

func decodeWindows1252(raw []byte) (string, bool) {
	var b strings.Builder
	b.Grow(len(raw))
	for _, byt := range raw {
		r, ok := windows1252Rune(byt)
		if !ok {
			return "", false
		}
		b.WriteRune(r)
	}
	return b.String(), true
}

func windows1252Rune(byt byte) (rune, bool) {
	switch byt {
	case 0x80:
		return rune(0x20AC), true
	case 0x82:
		return rune(0x201A), true
	case 0x83:
		return rune(0x0192), true
	case 0x84:
		return rune(0x201E), true
	case 0x85:
		return rune(0x2026), true
	case 0x86:
		return rune(0x2020), true
	case 0x87:
		return rune(0x2021), true
	case 0x88:
		return rune(0x02C6), true
	case 0x89:
		return rune(0x2030), true
	case 0x8A:
		return rune(0x0160), true
	case 0x8B:
		return rune(0x2039), true
	case 0x8C:
		return rune(0x0152), true
	case 0x8E:
		return rune(0x017D), true
	case 0x91:
		return rune(0x2018), true
	case 0x92:
		return rune(0x2019), true
	case 0x93:
		return rune(0x201C), true
	case 0x94:
		return rune(0x201D), true
	case 0x95:
		return rune(0x2022), true
	case 0x96:
		return rune(0x2013), true
	case 0x97:
		return rune(0x2014), true
	case 0x98:
		return rune(0x02DC), true
	case 0x99:
		return rune(0x2122), true
	case 0x9A:
		return rune(0x0161), true
	case 0x9B:
		return rune(0x203A), true
	case 0x9C:
		return rune(0x0153), true
	case 0x9E:
		return rune(0x017E), true
	case 0x9F:
		return rune(0x0178), true
	case 0x81, 0x8D, 0x8F, 0x90, 0x9D:
		return 0, false
	default:
		return rune(byt), true
	}
}

func decodeLatin1(raw []byte) string {
	var b strings.Builder
	b.Grow(len(raw))
	for _, byt := range raw {
		b.WriteRune(rune(byt))
	}
	return b.String()
}
