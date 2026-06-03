package files

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiskWriteReadRemoveCsv(t *testing.T) {
	dir := t.TempDir()
	store := NewDisk(dir)

	path, err := store.WriteCsv(strings.NewReader("a,b\n1,2\n"))
	if err != nil {
		t.Fatalf("WriteCsv failed: %v", err)
	}
	if filepath.Dir(path) != dir {
		t.Fatalf("expected file in %s, got %s", dir, path)
	}
	if filepath.Ext(path) != ".csv" {
		t.Fatalf("expected .csv extension, got %s", path)
	}

	rc, err := store.ReadCsv(path)
	if err != nil {
		t.Fatalf("ReadCsv failed: %v", err)
	}
	raw, err := io.ReadAll(rc)
	if closeErr := rc.Close(); closeErr != nil {
		t.Fatalf("failed closing file: %v", closeErr)
	}
	if err != nil {
		t.Fatalf("failed reading file: %v", err)
	}
	if string(raw) != "a,b\n1,2\n" {
		t.Fatalf("unexpected file contents: %q", string(raw))
	}

	if err := store.Remove(path); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected file to be removed, got %v", err)
	}
}

func TestCommandsAndQueries(t *testing.T) {
	ctx := context.Background()
	store := NewDisk(t.TempDir())

	commands := NewCommands(store)
	file, err := commands.WriteCsv(ctx, strings.NewReader("date,value\n2026-05-31,1\n"))
	if err != nil {
		t.Fatalf("WriteCsv command failed: %v", err)
	}
	if file.Type != FileTypeCSV {
		t.Fatalf("expected csv file type, got %s", file.Type)
	}

	queries := NewQueries(store)
	rc, err := queries.ReadCsv(ctx, file.Path)
	if err != nil {
		t.Fatalf("ReadCsv query failed: %v", err)
	}
	if err := rc.Close(); err != nil {
		t.Fatalf("failed closing file: %v", err)
	}

	if err := commands.Remove(ctx, file.Path); err != nil {
		t.Fatalf("Remove command failed: %v", err)
	}
}

func TestFileValidationErrors(t *testing.T) {
	ctx := context.Background()
	store := NewDisk(t.TempDir())

	if _, err := NewCommands(store).WriteCsv(ctx, nil); !errors.Is(err, ErrContentRequired) {
		t.Fatalf("expected ErrContentRequired, got %v", err)
	}
	if err := NewCommands(store).Remove(ctx, " "); !errors.Is(err, ErrPathRequired) {
		t.Fatalf("expected ErrPathRequired, got %v", err)
	}
	if _, err := NewQueries(store).ReadCsv(ctx, " "); !errors.Is(err, ErrPathRequired) {
		t.Fatalf("expected ErrPathRequired, got %v", err)
	}
}
