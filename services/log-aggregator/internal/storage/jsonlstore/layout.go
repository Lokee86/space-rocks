package jsonlstore

import (
	"fmt"
	"path/filepath"
	"time"
)

const (
	activeDirectory     = "active"
	archiveDirectory    = "archive"
	quarantineDirectory = "quarantine"
	activeFileName      = "events.jsonl.open"
)

// Layout centralizes the on-disk directories and names used by the backend.
type Layout struct {
	Root string
}

// NewLayout creates a path layout rooted at root.
func NewLayout(root string) Layout {
	return Layout{Root: root}
}

func (layout Layout) ActiveDir() string {
	return filepath.Join(layout.Root, activeDirectory)
}

func (layout Layout) ArchiveDir() string {
	return filepath.Join(layout.Root, archiveDirectory)
}

func (layout Layout) QuarantineDir() string {
	return filepath.Join(layout.Root, quarantineDirectory)
}

// ActivePath returns the single writable segment path.
func (layout Layout) ActivePath() string {
	return filepath.Join(layout.ActiveDir(), activeFileName)
}

// ArchivePath returns a deterministic archive path. The segment start selects
// the UTC archive date, while both timestamps and sequence identify the file.
func (layout Layout) ArchivePath(start, end time.Time, sequence uint64, compressed bool) string {
	extension := ".jsonl"
	if compressed {
		extension += ".gz"
	}
	start = start.UTC()
	end = end.UTC()
	name := fmt.Sprintf("events-%s-%s-%06d%s", formatTimestamp(start), formatTimestamp(end), sequence, extension)
	return filepath.Join(layout.ArchiveDir(), start.Format("2006"), start.Format("01"), start.Format("02"), name)
}

func (layout Layout) QuarantinePath(name string) string {
	return filepath.Join(layout.QuarantineDir(), filepath.Base(name))
}

func formatTimestamp(value time.Time) string {
	return value.UTC().Format("20060102T150405.000000000Z")
}
