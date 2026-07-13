package jsonlstore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/storage"
)

var ErrReportNotFound = errors.New("jsonlstore: diagnostic report not found")

type ReportStore struct {
	mu           sync.Mutex
	path         string
	writer       *writer
	config       Config
	layout       reportLayout
	activeBytes  int64
	segmentStart time.Time
	sequence     uint64
	retention    time.Duration
	now          func() time.Time
	closed       bool
}

var _ storage.ReportStore = (*ReportStore)(nil)

func NewReportStore(config Config) (*ReportStore, error) { return newReportStore(config, time.Now) }

func NewReportStoreWithClock(config Config, now func() time.Time) (*ReportStore, error) {
	if now == nil {
		return nil, errors.New("jsonlstore: report clock is required")
	}
	return newReportStore(config, now)
}

func newReportStore(config Config, now func() time.Time) (*ReportStore, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if config.Root == "" {
		return nil, errors.New("jsonlstore: report storage root is required")
	}
	layout := newReportLayout(config.Root)
	for _, path := range []string{layout.root, layout.activeDir(), layout.archiveDir(), layout.quarantineDir()} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return nil, fmt.Errorf("jsonlstore: create report directory: %w", err)
		}
	}
	sequence, err := recoverReportActive(layout, config, now())
	if err != nil {
		return nil, err
	}
	path := layout.activePath()
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("jsonlstore: open report file: %w", err)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	return &ReportStore{path: path, writer: newWriter(file, config.FlushInterval), config: config, layout: layout, activeBytes: info.Size(), sequence: sequence, segmentStart: now(), retention: config.DiagnosticReportRetention, now: now}, nil
}

func (store *ReportStore) EnforceRetention(ctx context.Context) (int, error) {
	return store.DeleteExpired(ctx, store.now().Add(-store.retention))
}

func (store *ReportStore) Save(ctx context.Context, report storage.Report) error {
	if err := reportContextError(ctx); err != nil {
		return err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.closed {
		return storage.ErrClosed
	}
	line, err := encodeReport(report)
	if err != nil {
		return fmt.Errorf("jsonlstore: encode diagnostic report: %w", err)
	}
	line = append(line, '\n')
	now := store.now()
	if store.activeBytes > 0 && (store.activeBytes+int64(len(line)) > store.config.SegmentMaxBytes || now.Sub(store.segmentStart) >= store.config.SegmentMaxAge) {
		if err := store.rotateReportLocked(now); err != nil {
			return err
		}
	}
	if err := store.writer.write(line); err != nil {
		return fmt.Errorf("jsonlstore: append diagnostic report: %w", err)
	}
	store.activeBytes += int64(len(line))
	if err := store.writer.durableFlush(); err != nil {
		return fmt.Errorf("jsonlstore: sync diagnostic report: %w", err)
	}
	return nil
}

func (store *ReportStore) Get(ctx context.Context, id string) (storage.Report, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.closed {
		return storage.Report{}, storage.ErrClosed
	}
	if err := store.writer.flush(); err != nil {
		return storage.Report{}, err
	}
	paths, err := reportSegmentPaths(store.layout)
	if err != nil {
		return storage.Report{}, err
	}
	var found storage.Report
	matched := false
	for _, path := range paths {
		input, openErr := openReportSegment(path)
		if os.IsNotExist(openErr) {
			continue
		}
		if openErr != nil {
			return storage.Report{}, openErr
		}
		report, scanErr := scanReportFile(ctx, input, id)
		_ = input.Close()
		if scanErr == nil {
			found, matched = report, true
			continue
		}
		if !errors.Is(scanErr, ErrReportNotFound) {
			return storage.Report{}, scanErr
		}
	}
	if !matched {
		return storage.Report{}, ErrReportNotFound
	}
	return found, nil
}

func (store *ReportStore) DeleteExpired(ctx context.Context, cutoff time.Time) (int, error) {
	if err := reportContextError(ctx); err != nil {
		return 0, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.closed {
		return 0, storage.ErrClosed
	}
	return store.deleteExpiredReports(ctx, cutoff)
}

func (store *ReportStore) Close() error {
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.closed {
		return nil
	}
	store.closed = true
	return store.finalizeReportLocked()
}
