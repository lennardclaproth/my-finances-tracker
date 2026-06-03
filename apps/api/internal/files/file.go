package files

import "strings"

// FileType identifies the kind of stored file.
type FileType string

const (
	// FileTypeCSV identifies comma-separated or parser-compatible text files.
	FileTypeCSV FileType = "csv"
)

// StoredFile describes a file after it has been persisted.
type StoredFile struct {
	Path string
	Type FileType
}

// NewStoredFile validates persisted file metadata.
func NewStoredFile(path string, fileType FileType) (*StoredFile, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, ErrPathRequired
	}
	return &StoredFile{
		Path: path,
		Type: fileType,
	}, nil
}
