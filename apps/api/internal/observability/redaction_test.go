package observability

import (
	"testing"

	"go.elastic.co/apm/v2"
	"go.elastic.co/apm/v2/apmtest"
	"go.elastic.co/apm/v2/model"
)

func TestFilterFields_DropsDisallowedKeys(t *testing.T) {
	t.Parallel()

	fields := FilterFields(
		"operation", "http.get.accounts",
		"description", "should-not-pass",
		"account_id", "acc-1",
	)

	if len(fields) != 4 {
		t.Fatalf("expected two allowed key-value pairs, got %d values", len(fields))
	}
	if fields[0] != "operation" || fields[2] != "account_id" {
		t.Fatalf("unexpected filtered keys: %#v", fields)
	}
}

func TestSetSafeTransactionLabels_AllowlistOnly(t *testing.T) {
	t.Parallel()

	recordingTracer := apmtest.NewRecordingTracer()
	defer recordingTracer.Close()

	previousDefaultTracer := apm.DefaultTracer()
	apm.SetDefaultTracer(recordingTracer.Tracer)
	defer apm.SetDefaultTracer(previousDefaultTracer)

	tx := apm.DefaultTracer().StartTransaction("test.tx", "custom")
	SetSafeTransactionLabels(tx, map[string]any{
		"account_id":     "acc-1",
		"description":    "do-not-store",
		"correlation_id": "corr-1",
	})
	tx.End()
	recordingTracer.Flush(nil)

	payloads := recordingTracer.Payloads()
	if len(payloads.Transactions) != 1 {
		t.Fatalf("expected one transaction payload, got %d", len(payloads.Transactions))
	}
	tags := payloads.Transactions[0].Context.Tags
	if !hasTag(tags, "account_id") {
		t.Fatalf("expected allowlisted account_id label")
	}
	if !hasTag(tags, "correlation_id") {
		t.Fatalf("expected allowlisted correlation_id label")
	}
	if hasTag(tags, "description") {
		t.Fatalf("expected non-allowlisted description label to be removed")
	}
}

func hasTag(tags model.IfaceMap, key string) bool {
	for _, item := range tags {
		if item.Key == key {
			return true
		}
	}
	return false
}
