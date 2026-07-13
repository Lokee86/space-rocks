package jsonlstore

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/storage"
)

// Status reports persisted records and the current backend lifecycle state.
func (store *Store) Status(ctx context.Context) (storage.Status, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	status := storage.Status{Ready: !store.closed && store.unavailable() == nil, Degraded: store.degraded || store.writerErr != nil}
	if !store.closed && store.writer != nil && store.writerErr == nil {
		if err := store.writer.flush(); err != nil {
			store.writerErr = err
			status.Ready = false
			status.Degraded = true
		}
	}
	paths, err := statusPaths(store.layout)
	if err != nil {
		status.Degraded = true
		return status, err
	}
	for _, path := range paths {
		if err := scanStatusFile(ctx, path, &status); err != nil {
			status.Degraded = true
			return status, err
		}
	}
	status.Ready = !store.closed && store.unavailable() == nil
	status.Degraded = status.Degraded || store.degraded || store.writerErr != nil
	return status, nil
}

func statusPaths(layout Layout) ([]string, error) {
	paths := []string{layout.ActivePath()}
	err := filepath.Walk(layout.ArchiveDir(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !(strings.HasSuffix(info.Name(), ".jsonl") || strings.HasSuffix(info.Name(), ".jsonl.gz")) {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	return paths, err
}

func scanStatusFile(ctx context.Context, path string, status *storage.Status) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("jsonlstore: scan status file %q: %w", path, ctx.Err())
	default:
	}
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("jsonlstore: open status file %q: %w", path, err)
	}
	defer file.Close()
	var input io.Reader = file
	var compressed *gzip.Reader
	if strings.HasSuffix(path, ".gz") {
		compressed, err = gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("jsonlstore: open gzip status file %q: %w", path, err)
		}
		defer compressed.Close()
		input = compressed
	}
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return fmt.Errorf("jsonlstore: scan status file %q: %w", path, ctx.Err())
		default:
		}
		record, err := DecodeRecord(scanner.Bytes())
		if err != nil {
			return fmt.Errorf("jsonlstore: decode status record in %q: %w", path, err)
		}
		status.RecordCount++
		status.ByteCount += uint64(len(record.Payload))
		if status.Oldest.IsZero() || record.Timestamp.Before(status.Oldest) {
			status.Oldest = record.Timestamp
		}
		if status.Newest.IsZero() || record.Timestamp.After(status.Newest) {
			status.Newest = record.Timestamp
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("jsonlstore: scan status file %q: %w", path, err)
	}
	return nil
}
