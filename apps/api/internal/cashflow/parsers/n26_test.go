package parsers

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

func TestN26Parser_ParseAll_EnglishHeaders(t *testing.T) {
	input := `"Booking Date","Value Date","Partner Name","Partner Iban",Type,"Payment Reference","Account Name","Amount (EUR)","Original Amount","Original Currency","Exchange Rate"
2025-06-21,2025-06-20,"DM Drogerie Markt",,Presentment,,"Lennard & Greta",-1.85,1.85,EUR,1
2025-06-29,2025-06-29,,DE36100101787510095616,"Credit Transfer","PAYMENT RETURN OBAZ0A22CXJTCAWCG","Lennard & Greta",0.01,,,
`

	p := NewN26Parser()
	seq, err := p.ParseAll(io.NopCloser(strings.NewReader(input)))
	if err != nil {
		t.Fatalf("ParseAll returned error: %v", err)
	}

	var got []cashflow.TransactionData
	for _, txd := range seq {
		got = append(got, txd)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 parsed rows, got %d", len(got))
	}

	if got[0].Description != "DM Drogerie Markt" {
		t.Fatalf("unexpected row 1 description: %s", got[0].Description)
	}
	if got[0].Direction != cashflow.CashOut {
		t.Fatalf("expected row 1 direction out, got %s", got[0].Direction)
	}
	price, _ := money.NewPrice(-1.85)
	if got[0].Amount != price {
		t.Fatalf("expected row 1 amount -1.85, got %f", got[0].Amount)
	}
	if got[0].Date.Format("2006-01-02") != "2025-06-21" {
		t.Fatalf("expected booking date 2025-06-21, got %s", got[0].Date.Format("2006-01-02"))
	}
	if got[0].Source != "N26" {
		t.Fatalf("expected source N26, got %s", got[0].Source)
	}
	if got[0].AccountType == nil || *got[0].AccountType != cashflow.AccountTypeChecking {
		t.Fatalf("expected checking account type")
	}

	if got[1].Description != "PAYMENT RETURN OBAZ0A22CXJTCAWCG" {
		t.Fatalf("expected fallback description from payment reference, got %s", got[1].Description)
	}
	if got[1].Direction != cashflow.CashIn {
		t.Fatalf("expected row 2 direction in, got %s", got[1].Direction)
	}
	if !strings.Contains(got[1].Note, "IBAN:DE36100101787510095616") {
		t.Fatalf("expected note to include IBAN, got %s", got[1].Note)
	}
	if !strings.Contains(got[1].Note, "Type:Credit Transfer") {
		t.Fatalf("expected note to include type, got %s", got[1].Note)
	}
}

func TestN26Parser_ParseAll_GermanHeadersAndSemicolon(t *testing.T) {
	input := `"Buchungsdatum";"Wertstellungsdatum";"Partnername";"Partner IBAN";"Typ";"Verwendungszweck";"Kontoname";"Betrag (EUR)";"Originalbetrag";"Originalwaehrung";"Wechselkurs"
2025-08-13;2025-08-13;"GC RE OCTOPUS ENERGY GERMANY GMBH";"FR7617118820080000002438012";"Direct Debit";"REF-3C40BA8A18F8401A3C";"Lennard & Greta";-79,88;;;
`

	p := NewN26Parser()
	seq, err := p.ParseAll(io.NopCloser(strings.NewReader(input)))
	if err != nil {
		t.Fatalf("ParseAll returned error: %v", err)
	}

	var got []cashflow.TransactionData
	for _, txd := range seq {
		got = append(got, txd)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 parsed row, got %d", len(got))
	}
	if got[0].Direction != cashflow.CashOut {
		t.Fatalf("expected direction out, got %s", got[0].Direction)
	}
	price, _ := money.NewPrice(-79.88)
	if got[0].Amount !=  price{
		t.Fatalf("expected amount -79.88, got %f", got[0].Amount)
	}
	if got[0].Date.Format("2006-01-02") != "2025-08-13" {
		t.Fatalf("expected booking date 2025-08-13, got %s", got[0].Date.Format("2006-01-02"))
	}
	if !strings.Contains(got[0].Note, "Reference:REF-3C40BA8A18F8401A3C") {
		t.Fatalf("expected note to include reference, got %s", got[0].Note)
	}
}

func TestN26Parser_ParseAll_EncodingFallback_Windows1252(t *testing.T) {
	raw := []byte("\"Booking Date\",\"Value Date\",\"Partner Name\",\"Partner Iban\",Type,\"Payment Reference\",\"Account Name\",\"Amount (EUR)\",\"Original Amount\",\"Original Currency\",\"Exchange Rate\"\n" +
		"2025-06-18,2025-06-18,,,\"Credit Transfer\",\"Danke f\xFCr den Import\",\"Lennard & Greta\",50.00,,,\n")

	p := NewN26Parser()
	seq, err := p.ParseAll(io.NopCloser(bytes.NewReader(raw)))
	if err != nil {
		t.Fatalf("ParseAll returned error: %v", err)
	}

	var got []cashflow.TransactionData
	for _, txd := range seq {
		got = append(got, txd)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 parsed row, got %d", len(got))
	}
	if !strings.Contains(got[0].Note, "Danke f\u00fcr den Import") {
		t.Fatalf("expected decoded umlaut in note, got %s", got[0].Note)
	}
}

func TestCreateCsvParser_N26(t *testing.T) {
	p, err := CreateCsvParser(vendor.VendorN26)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p == nil {
		t.Fatalf("expected parser instance")
	}
}
