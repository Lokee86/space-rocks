package jsonlstore

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (store *ReportStore) deleteExpiredReports(ctx context.Context, cutoff time.Time) (int, error) {
	if store.activeBytes > 0 {
		if err := store.rotateReportLocked(store.now()); err != nil {
			return 0, err
		}
	}
	paths, err := reportSegmentPaths(store.layout)
	if err != nil {
		return 0, err
	}
	removed := 0
	for _, path := range paths {
		if path == store.layout.activePath() {
			continue
		}
		count, err := rewriteReportArchive(ctx, path, cutoff)
		if err != nil {
			return removed, err
		}
		removed += count
	}
	return removed, removeEmptyArchiveDirs(store.layout.archiveDir())
}

func rewriteReportArchive(ctx context.Context, path string, cutoff time.Time) (int, error) {
	input, err := openReportSegment(path)
	if err != nil {
		return 0, err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".diagnostic-reports-retention-")
	if err != nil {
		_ = input.Close()
		return 0, err
	}
	tempName := temp.Name()
	compressed := strings.HasSuffix(path, ".gz")
	var output io.Writer = temp
	var gzipOutput *gzip.Writer
	if compressed {
		gzipOutput = gzip.NewWriter(temp)
		output = gzipOutput
	}
	kept, removed, rewriteErr := rewriteReports(ctx, input, output, cutoff)
	if gzipOutput != nil {
		if closeErr := gzipOutput.Close(); rewriteErr == nil {
			rewriteErr = closeErr
		}
	}
	if closeErr := input.Close(); rewriteErr == nil {
		rewriteErr = closeErr
	}
	if syncErr := temp.Sync(); rewriteErr == nil {
		rewriteErr = syncErr
	}
	if closeErr := temp.Close(); rewriteErr == nil {
		rewriteErr = closeErr
	}
	if rewriteErr != nil {
		_ = os.Remove(tempName)
		return 0, rewriteErr
	}
	if kept == 0 {
		_ = os.Remove(tempName)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return 0, err
		}
		return removed, nil
	}
	backup := tempName + ".backup"
	if err := os.Rename(path, backup); err != nil {
		_ = os.Remove(tempName)
		return 0, err
	}
	if err := os.Rename(tempName, path); err != nil {
		if restoreErr := os.Rename(backup, path); restoreErr != nil {
			_ = os.Remove(tempName)
			return 0, fmt.Errorf("jsonlstore: replace report archive: %w; restore original: %v", err, restoreErr)
		}
		_ = os.Remove(tempName)
		return 0, fmt.Errorf("jsonlstore: replace report archive: %w", err)
	}
	if err := os.Remove(backup); err != nil {
		return 0, fmt.Errorf("jsonlstore: remove report archive backup: %w", err)
	}
	return removed, nil
}
