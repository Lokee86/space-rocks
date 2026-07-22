package jsonlstore

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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
