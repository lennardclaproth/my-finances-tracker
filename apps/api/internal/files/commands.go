package files

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type commandStore interface {
	WriteCsv(r io.Reader) (string, error)
	Remove(path string) error
}

// Commands exposes file-storage write-side use cases.
type Commands struct {
	store commandStore
}

// NewCommands creates file-storage write-side use cases.
func NewCommands(store commandStore) *Commands {
	return &Commands{store: store}
}

// WriteCsv stores CSV-compatible content and returns the persisted file metadata.
func (c *Commands) WriteCsv(ctx context.Context, r io.Reader) (*StoredFile, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("write csv: %w", err)
	}
	if r == nil {
		return nil, ErrContentRequired
	}
	if c.store == nil {
		return nil, ErrStoreNotConfigured
	}

	path, err := c.store.WriteCsv(r)
	if err != nil {
		return nil, fmt.Errorf("write csv: %w", err)
	}
	file, err := NewStoredFile(path, FileTypeCSV)
	if err != nil {
		return nil, fmt.Errorf("write csv: %w", err)
	}
	return file, nil
}

// Remove deletes a stored file.
func (c *Commands) Remove(ctx context.Context, path string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("remove file: %w", err)
	}
	if strings.TrimSpace(path) == "" {
		return ErrPathRequired
	}
	if c.store == nil {
		return ErrStoreNotConfigured
	}
	if err := c.store.Remove(path); err != nil {
		return fmt.Errorf("remove file: %w", err)
	}
	return nil
}
