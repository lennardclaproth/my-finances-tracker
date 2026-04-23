package handlers

import "testing"

func TestParseAssetDateRange_ValidRange(t *testing.T) {
	from, to, err := parseAssetDateRange("2026-03-01", "2026-03-31")
	if err != nil {
		t.Fatalf("expected valid range parse, got error: %v", err)
	}
	if from == nil || to == nil {
		t.Fatalf("expected non-nil from/to, got from=%v to=%v", from, to)
	}
	if got := from.Format("2006-01-02T15:04:05.999999999Z07:00"); got != "2026-03-01T00:00:00Z" {
		t.Fatalf("unexpected from value: %s", got)
	}
	if got := to.Format("2006-01-02T15:04:05.999999999Z07:00"); got != "2026-03-31T23:59:59.999999999Z" {
		t.Fatalf("unexpected to value: %s", got)
	}
}

func TestParseAssetDateRange_InvalidFromReturnsFieldProblem(t *testing.T) {
	_, _, err := parseAssetDateRange("2026/03/01", "")
	if err == nil {
		t.Fatal("expected from parse error, got nil")
	}
	problem := assetDateRangeProblem(err)
	if problem["from"] != "from must be in YYYY-MM-DD format" {
		t.Fatalf("expected from problem, got %+v", problem)
	}
}

func TestParseAssetDateRange_FromAfterToReturnsFieldProblem(t *testing.T) {
	_, _, err := parseAssetDateRange("2026-03-10", "2026-03-01")
	if err == nil {
		t.Fatal("expected from-after-to error, got nil")
	}
	problem := assetDateRangeProblem(err)
	if problem["from"] != "from must be before or equal to to" {
		t.Fatalf("expected from ordering problem, got %+v", problem)
	}
}
