package files

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	randomFilenameLength = 15
	filenameCharset      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var (
	_ commandStore = (*Disk)(nil)
	_ queryStore   = (*Disk)(nil)
)

// Disk provides filesystem persistence for uploaded CSV-compatible files.
type Disk struct {
	basePath string
}

// NewDisk creates a disk-backed file store rooted at basePath.
func NewDisk(basePath string) *Disk {
	return &Disk{basePath: basePath}
}

// WriteCsv writes CSV-compatible contents to a generated filename and returns the stored path.
func (d *Disk) WriteCsv(r io.Reader) (path string, err error) {
	if r == nil {
		return "", ErrContentRequired
	}
	if err := os.MkdirAll(d.basePath, 0o755); err != nil {
		return "", fmt.Errorf("create base dir: %w", err)
	}

	filename, err := d.generateRandFilename()
	if err != nil {
		return "", err
	}
	fullPath := filepath.Join(d.basePath, filename+".csv")

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

	if err := f.Sync(); err != nil {
		if removeErr := os.Remove(fullPath); removeErr != nil {
			return "", fmt.Errorf("sync file: %w (cleanup remove failed: %v)", err, removeErr)
		}
		return "", fmt.Errorf("sync file: %w", err)
	}

	return fullPath, nil
}

// Remove removes a file at the given path.
func (d *Disk) Remove(path string) error {
	if strings.TrimSpace(path) == "" {
		return ErrPathRequired
	}
	return os.Remove(path)
}

// ReadCsv opens a CSV-compatible file for reading.
func (d *Disk) ReadCsv(path string) (io.ReadCloser, error) {
	if strings.TrimSpace(path) == "" {
		return nil, ErrPathRequired
	}
	return os.Open(path)
}

func (d *Disk) generateRandFilename() (string, error) {
	b := make([]byte, randomFilenameLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = filenameCharset[int(b[i])%len(filenameCharset)]
	}
	return string(b), nil
}
