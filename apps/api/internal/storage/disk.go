package storage

import "github.com/lennardclaproth/my-finances-tracker/internal/files"

// Disk provides filesystem persistence for uploaded CSV files.
//
// Deprecated: use files.Disk from internal/files.
type Disk = files.Disk

// NewDisk creates a disk-backed storage helper rooted at basePath.
//
// Deprecated: use files.NewDisk from internal/files.
func NewDisk(basePath string) *Disk {
	return files.NewDisk(basePath)
}
