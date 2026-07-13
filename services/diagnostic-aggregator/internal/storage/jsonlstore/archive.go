package jsonlstore

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func finalizeSegment(layout Layout, activePath string, start, end time.Time, sequence uint64, compressed bool) (uint64, error) {
	for {
		archivePath := layout.ArchivePath(start, end, sequence, compressed)
		if _, err := os.Stat(archivePath); err == nil {
			sequence++
			continue
		} else if !os.IsNotExist(err) {
			return sequence, fmt.Errorf("jsonlstore: inspect archive path: %w", err)
		}
		if err := finalizeArchive(activePath, archivePath, compressed); err != nil {
			return sequence, err
		}
		return sequence + 1, nil
	}
}

// finalizeArchive durably copies or atomically renames a closed active source.
func finalizeArchive(source, destination string, compressed bool) (result error) {
	if source == "" || destination == "" {
		return errors.New("jsonlstore: archive source and destination are required")
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return fmt.Errorf("jsonlstore: create archive directory: %w", err)
	}
	if !compressed {
		if err := os.Rename(source, destination); err != nil {
			return fmt.Errorf("jsonlstore: rename archive source: %w", err)
		}
		return nil
	}

	temporary, err := os.CreateTemp(filepath.Dir(destination), "."+filepath.Base(destination)+".tmp-")
	if err != nil {
		return fmt.Errorf("jsonlstore: create temporary gzip archive: %w", err)
	}
	temporaryName := temporary.Name()
	defer func() {
		if err := os.Remove(temporaryName); err != nil && !os.IsNotExist(err) && result == nil {
			result = fmt.Errorf("jsonlstore: remove temporary gzip archive: %w", err)
		}
	}()

	input, err := os.Open(source)
	if err != nil {
		_ = temporary.Close()
		return fmt.Errorf("jsonlstore: open archive source: %w", err)
	}
	gzipWriter := gzip.NewWriter(temporary)
	if _, err = io.Copy(gzipWriter, input); err != nil {
		_ = input.Close()
		_ = gzipWriter.Close()
		_ = temporary.Close()
		return fmt.Errorf("jsonlstore: gzip archive source: %w", err)
	}
	if err = input.Close(); err != nil {
		_ = gzipWriter.Close()
		_ = temporary.Close()
		return fmt.Errorf("jsonlstore: close archive source: %w", err)
	}
	if err = gzipWriter.Close(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("jsonlstore: close gzip archive: %w", err)
	}
	if err = temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("jsonlstore: sync gzip archive: %w", err)
	}
	if err = temporary.Close(); err != nil {
		return fmt.Errorf("jsonlstore: close gzip archive file: %w", err)
	}
	if err = os.Rename(temporaryName, destination); err != nil {
		return fmt.Errorf("jsonlstore: rename gzip archive: %w", err)
	}
	if err = os.Remove(source); err != nil {
		return fmt.Errorf("jsonlstore: remove archive source: %w", err)
	}
	return nil
}
