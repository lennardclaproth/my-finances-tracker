package eventbus

import "context"

type contextKey int

const metadataContextKey contextKey = iota

// ContextWithMetadata returns a context carrying the given message metadata so
// that events published while a message is being handled inherit its
// correlation chain (correlation ID and causation from the parent message).
func ContextWithMetadata(ctx context.Context, meta Metadata) context.Context {
	return context.WithValue(ctx, metadataContextKey, meta)
}

// MetadataFromContext returns the message metadata stored in ctx by
// ContextWithMetadata, reporting whether any was present.
func MetadataFromContext(ctx context.Context) (Metadata, bool) {
	if ctx == nil {
		return Metadata{}, false
	}
	meta, ok := ctx.Value(metadataContextKey).(Metadata)
	return meta, ok
}
