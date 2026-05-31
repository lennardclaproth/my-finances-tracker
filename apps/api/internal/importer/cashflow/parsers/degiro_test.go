package parsers

import (
	"io"
	"strings"
	"testing"

	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

func TestDegiroParser_ParseAll_HappyPath(t *testing.T) {
	input := `Date,Time,Value date,Product,ISIN,Description,FX,Change,,Balance,,Order Id
08-02-2026,14:02,08-02-2026,,,SEPA Instant Terugstorting,,EUR,"-480,00",EUR,"7,43",
19-01-2026,12:30,19-01-2026,,,Degiro Cash Sweep Transfer,,EUR,"641,02",EUR,"487,43",
19-01-2026,11:30,19-01-2026,VANECK AEX UCITS ETF,NL0009272749,DEGIRO Transactiekosten en/of kosten van derden,,EUR,"-3,00",EUR,"487,43",6c44d01d-ca3a-4786-a432-b1fe0f19be4b
19-01-2026,11:30,19-01-2026,VANGUARD S&P 500 UCITS ETF USD DIS,IE00B3XXRP09,"Koop 3 @ 111,962 EUR",,EUR,"-335,89",EUR,"792,56",47509942-dc4e-4d27-8806-903428c8f88a
`
	p := NewDegiroParser()
	seq, err := p.ParseAll(io.NopCloser(strings.NewReader(input)))
	if err != nil {
		t.Fatalf("ParseAll returned error: %v", err)
	}

	var got []cashflow.TransactionData
	for _, txd := range seq {
		got = append(got, txd)
	}

	if len(got) != 4 {
		t.Fatalf("expected 4 parsed rows, got %d", len(got))
	}

	if got[0].Description != "SEPA Instant Terugstorting" {
		t.Fatalf("unexpected row 1 description: %s", got[0].Description)
	}
	if got[0].Direction != cashflow.CashOut {
		t.Fatalf("expected row 1 direction out, got %s", got[0].Direction)
	}
	price, err := money.NewPrice(480.00)
	if err != nil {
		t.Fatalf("failed to build expected price: %v", err)
	}
	if got[0].Amount != price {
		t.Fatalf("expected row 1 amount 480, got %s", got[0].Amount)
	}

	if got[1].Direction != cashflow.CashIn {
		t.Fatalf("expected row 2 direction in, got %s", got[1].Direction)
	}
	price, err = money.NewPrice(641.02)
	if got[1].Amount != price {
		t.Fatalf("expected row 2 amount 641.02, got %s", got[1].Amount)
	}

	if got[2].AccountType == nil || *got[2].AccountType != cashflow.AccountTypeBrokerage {
		t.Fatalf("expected brokerage account type")
	}
	if got[2].Source != "DEGIRO" {
		t.Fatalf("expected source DEGIRO, got %s", got[2].Source)
	}
	if !strings.Contains(got[2].Note, "ISIN:NL0009272749") {
		t.Fatalf("expected note to contain ISIN, got %s", got[2].Note)
	}
	if !strings.Contains(got[2].Note, "OrderID:6c44d01d-ca3a-4786-a432-b1fe0f19be4b") {
		t.Fatalf("expected note to contain order id, got %s", got[2].Note)
	}
}

func TestDegiroParser_ParseAll_MissingRequiredHeader(t *testing.T) {
	input := `Date,Time,Value date,Product,ISIN,Description,FX,,Balance,,Order Id
08-02-2026,14:02,08-02-2026,,,SEPA Instant Terugstorting,,EUR,"7,43",
`

	p := NewDegiroParser()
	_, err := p.ParseAll(io.NopCloser(strings.NewReader(input)))
	if err == nil {
		t.Fatalf("expected error for missing required header")
	}
	if !strings.Contains(err.Error(), "missing required header: Change") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDegiroParser_ParseAll_FallbackToDateWhenValueDateEmpty(t *testing.T) {
	input := `Date,Time,Value date,Product,ISIN,Description,FX,Change,,Balance,,Order Id
08-02-2026,14:02,,,NL0009272749,SEPA Instant Terugstorting,, "-10,00",,EUR,"7,43",
`

	p := NewDegiroParser()
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
	if got[0].Date.Format("2006-01-02") != "2026-02-08" {
		t.Fatalf("expected parsed fallback date 2026-02-08, got %s", got[0].Date.Format("2006-01-02"))
	}
}

func TestCreateCsvParser_Degiro(t *testing.T) {
	p, err := CreateCsvParser(vendor.VendorDEGIRO)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p == nil {
		t.Fatalf("expected parser instance")
	}
}
