package storage

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Disk provides filesystem persistence for uploaded CSV files.
type Disk struct {
	basePath string
}

// NewDisk creates a disk-backed storage helper rooted at basePath.
func NewDisk(basePath string) *Disk {
	return &Disk{basePath: basePath}
}

// WriteCsv writes CSV contents to a generated filename and returns the absolute path.
func (dw *Disk) WriteCsv(r io.Reader) (path string, err error) {
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
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			if err == nil {
				err = fmt.Errorf("close file: %w", closeErr)
			} else {
				err = errors.Join(err, fmt.Errorf("close file: %w", closeErr))
			}
		}
	}()

	if _, err := io.Copy(f, r); err != nil {
		if removeErr := os.Remove(fullPath); removeErr != nil {
			return "", fmt.Errorf("write file: %w (cleanup remove failed: %v)", err, removeErr)
		}
		return "", fmt.Errorf("write file: %w", err)
	}

	// Optional: ensure it's flushed to disk
	if err := f.Sync(); err != nil {
		if removeErr := os.Remove(fullPath); removeErr != nil {
			return "", fmt.Errorf("sync file: %w (cleanup remove failed: %v)", err, removeErr)
		}
		return "", fmt.Errorf("sync file: %w", err)
	}

	return fullPath, nil
}

// Remove removes a file at the given path.
func (dw *Disk) Remove(path string) error {
	return os.Remove(path)
}

// ReadCsv opens a CSV file for reading.
func (dw *Disk) ReadCsv(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (dw *Disk) generateRandFilename() (string, error) {
	b := make([]byte, 15)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}
