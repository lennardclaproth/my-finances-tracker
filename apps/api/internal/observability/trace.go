package observability

import (
	"context"
	"fmt"
	"strings"

	"go.elastic.co/apm/module/apmhttp/v2"
	"go.elastic.co/apm/v2"
)

type TraceIdentifiers struct {
	TraceID       string
	TransactionID string
	SpanID        string
}

func TraceIdentifiersFromContext(ctx context.Context) TraceIdentifiers {
	var ids TraceIdentifiers

	if tx := apm.TransactionFromContext(ctx); tx != nil {
		tc := tx.TraceContext()
		ids.TraceID = tc.Trace.String()
		ids.TransactionID = tc.Span.String()
	}

	if span := apm.SpanFromContext(ctx); span != nil {
		sc := span.TraceContext()
		if ids.TraceID == "" {
			ids.TraceID = sc.Trace.String()
		}
		ids.SpanID = sc.Span.String()
	}

	return ids
}

func ContextFields(ctx context.Context) []any {
	ids := TraceIdentifiersFromContext(ctx)
	fields := make([]any, 0, 8)
	if ids.TraceID != "" {
		fields = append(fields, "trace.id", ids.TraceID)
	}
	if ids.TransactionID != "" {
		fields = append(fields, "transaction.id", ids.TransactionID)
	}
	if ids.SpanID != "" {
		fields = append(fields, "span.id", ids.SpanID)
	}

	if requestID := RequestIDFromContext(ctx); requestID != "" {
		fields = append(fields, "request_id", requestID)
	}
	if correlationID := CorrelationIDFromContext(ctx); correlationID != "" {
		fields = append(fields, "correlation_id", correlationID)
	}
	return fields
}

func AppendContextFields(ctx context.Context, fields ...any) []any {
	out := make([]any, 0, len(fields)+8)
	out = append(out, fields...)
	out = append(out, ContextFields(ctx)...)
	return out
}

func PropagationHeadersFromContext(ctx context.Context) map[string]string {
	headers := make(map[string]string, 4)
	if requestID := RequestIDFromContext(ctx); requestID != "" {
		headers[HeaderRequestID] = requestID
	}
	if correlationID := CorrelationIDFromContext(ctx); correlationID != "" {
		headers[HeaderCorrelationID] = correlationID
	}
	for k, v := range TraceHeadersFromContext(ctx) {
		headers[k] = v
	}
	return headers
}

func ContextWithPropagationHeaders(ctx context.Context, headers map[string]string) context.Context {
	if headers == nil {
		return ctx
	}
	requestID := strings.TrimSpace(headers[HeaderRequestID])
	correlationID := strings.TrimSpace(headers[HeaderCorrelationID])
	ctx = ContextWithRequestID(ctx, requestID)
	ctx = ContextWithCorrelationID(ctx, correlationID)
	return ctx
}

func TraceHeadersFromContext(ctx context.Context) map[string]string {
	headers := make(map[string]string, 2)

	var tc apm.TraceContext
	if span := apm.SpanFromContext(ctx); span != nil {
		tc = span.TraceContext()
	} else if tx := apm.TransactionFromContext(ctx); tx != nil {
		tc = tx.TraceContext()
	}

	if tc.Trace.Validate() != nil || tc.Span.Validate() != nil {
		return headers
	}

	headers[HeaderTraceparent] = apmhttp.FormatTraceparentHeader(tc)
	if tracestate := tc.State.String(); tracestate != "" {
		headers[HeaderTracestate] = tracestate
	}
	return headers
}

func TraceContextFromHeaders(headers map[string]string) (apm.TraceContext, bool, error) {
	if headers == nil {
		return apm.TraceContext{}, false, nil
	}
	traceparent := strings.TrimSpace(headers[HeaderTraceparent])
	if traceparent == "" {
		return apm.TraceContext{}, false, nil
	}

	traceContext, err := apmhttp.ParseTraceparentHeader(traceparent)
	if err != nil {
		return apm.TraceContext{}, false, fmt.Errorf("parse traceparent: %w", err)
	}

	if tracestate := strings.TrimSpace(headers[HeaderTracestate]); tracestate != "" {
		state, err := apmhttp.ParseTracestateHeader(tracestate)
		if err != nil {
			return apm.TraceContext{}, false, fmt.Errorf("parse tracestate: %w", err)
		}
		traceContext.State = state
	}

	return traceContext, true, nil
}

func StartTransactionFromHeaders(
	ctx context.Context,
	name string,
	transactionType string,
	headers map[string]string,
) (*apm.Transaction, context.Context, error) {
	opts := apm.TransactionOptions{}
	traceContext, ok, err := TraceContextFromHeaders(headers)
	if err != nil {
		// Start a root transaction even if incoming trace headers are malformed.
		tx := apm.DefaultTracer().StartTransaction(name, transactionType)
		ctx = apm.ContextWithTransaction(ctx, tx)
		return tx, ctx, err
	}
	if ok {
		opts.TraceContext = traceContext
	}

	tx := apm.DefaultTracer().StartTransactionOptions(name, transactionType, opts)
	ctx = apm.ContextWithTransaction(ctx, tx)
	return tx, ctx, nil
}

func HTTPOperation(method string, route string) string {
	normalizedMethod := strings.ToLower(strings.TrimSpace(method))
	if normalizedMethod == "" {
		normalizedMethod = "unknown"
	}
	normalizedRoute := normalizeOperationPart(route)
	return fmt.Sprintf("http.%s.%s", normalizedMethod, normalizedRoute)
}

func JobOperation(jobName string) string {
	return fmt.Sprintf("job.%s.process", normalizeOperationPart(jobName))
}

func BusConsumeOperation(topic string) string {
	return fmt.Sprintf("bus.consume.%s", normalizeOperationPart(topic))
}

func BusPublishOperation(topic string) string {
	return fmt.Sprintf("bus.publish.%s", normalizeOperationPart(topic))
}

func normalizeOperationPart(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "unknown"
	}
	replacer := strings.NewReplacer(
		" ", "_",
		"/", ".",
		"\\", ".",
		"-", "_",
		":", "_",
		"{", "",
		"}", "",
	)
	value = replacer.Replace(value)
	value = strings.Trim(value, ".")
	if value == "" {
		return "unknown"
	}
	return value
}
