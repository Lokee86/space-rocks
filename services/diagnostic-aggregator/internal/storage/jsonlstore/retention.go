package jsonlstore

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type archiveEntry struct {
	path string
	mod  time.Time
	size int64
}

// enforceRetention applies age first, then the total archive byte cap.
func enforceRetention(layout Layout, now time.Time, maxAge time.Duration, maxBytes int64) error {
	entries := make([]archiveEntry, 0)
	if err := filepath.Walk(layout.ArchiveDir(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !(strings.HasSuffix(info.Name(), ".jsonl") || strings.HasSuffix(info.Name(), ".jsonl.gz")) {
			return nil
		}
		entries = append(entries, archiveEntry{path: path, mod: info.ModTime(), size: info.Size()})
		return nil
	}); err != nil {
		return fmt.Errorf("jsonlstore: scan archives for retention: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].mod.Equal(entries[j].mod) {
			return entries[i].path < entries[j].path
		}
		return entries[i].mod.Before(entries[j].mod)
	})
	remaining := make([]archiveEntry, 0, len(entries))
	for _, entry := range entries {
		if now.Sub(entry.mod) >= maxAge {
			if err := os.Remove(entry.path); err != nil {
				return fmt.Errorf("jsonlstore: delete expired archive %q: %w", entry.path, err)
			}
			continue
		}
		remaining = append(remaining, entry)
	}
	var total int64
	for _, entry := range remaining {
		total += entry.size
	}
	for _, entry := range remaining {
		if total <= maxBytes {
			break
		}
		if err := os.Remove(entry.path); err != nil {
			return fmt.Errorf("jsonlstore: delete over-cap archive %q: %w", entry.path, err)
		}
		total -= entry.size
	}
	return removeEmptyArchiveDirs(layout.ArchiveDir())
}

func removeEmptyArchiveDirs(root string) error {
	var directories []string
	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path != root && info.IsDir() {
			directories = append(directories, path)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("jsonlstore: scan archive directories: %w", err)
	}
	sort.Slice(directories, func(i, j int) bool { return directoryDepth(directories[i]) > directoryDepth(directories[j]) })
	for _, path := range directories {
		entries, err := os.ReadDir(path)
		if err != nil {
			return fmt.Errorf("jsonlstore: read archive directory %q: %w", path, err)
		}
		if len(entries) > 0 {
			continue
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("jsonlstore: remove empty archive directory %q: %w", path, err)
		}
	}
	return nil
}

func directoryDepth(path string) int {
	return len(strings.Split(filepath.Clean(path), string(os.PathSeparator)))
}
