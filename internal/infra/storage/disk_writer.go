package storage

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type DiskWriter struct {
	basePath string
}

func NewDiskWriter(basePath string) *DiskWriter {
	return &DiskWriter{basePath: basePath}
}

func (dw *DiskWriter) WriteCsv(r io.Reader) (string, error) {
	// Ensure base path exists
	if err := os.MkdirAll(dw.basePath, 0o755); err != nil {
		return "", fmt.Errorf("create base dir: %w", err)
	}

	filename, err := dw.generateRandFilename()
	if err != nil {
		return "", err
	}

	fullPath := filepath.Join(dw.basePath, filename+".csv")

	f, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		_ = os.Remove(fullPath) // cleanup partial file
		return "", fmt.Errorf("write file: %w", err)
	}

	// Optional: ensure it's flushed to disk
	if err := f.Sync(); err != nil {
		_ = os.Remove(fullPath)
		return "", fmt.Errorf("sync file: %w", err)
	}

	return fullPath, nil
}

func (dw *DiskWriter) Remove(path string) error {
	return os.Remove(path)
}

func (dw *DiskWriter) ReadCsv(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (dw *DiskWriter) generateRandFilename() (string, error) {
	b := make([]byte, 15)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}
