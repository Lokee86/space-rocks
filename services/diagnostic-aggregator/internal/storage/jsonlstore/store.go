package jsonlstore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage"
)

// Store is the rolling JSONL event store.
type Store struct {
	mu           sync.RWMutex
	config       Config
	layout       Layout
	clock        Clock
	writer       *writer
	activeBytes  int64
	segmentStart time.Time
	sequence     uint64
	degraded     bool
	writerErr    error
	closed       bool
	closeErr     error
}

// New opens a Store using the real process clock.
func New(config Config) (*Store, error) { return NewWithClock(config, RealClock{}) }

// NewWithClock opens a Store using the supplied clock.
func NewWithClock(config Config, clock Clock) (*Store, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if clock == nil {
		return nil, errors.New("jsonlstore: clock is required")
	}
	layout := NewLayout(config.Root)
	for _, path := range []string{layout.Root, layout.ArchiveDir(), layout.QuarantineDir(), layout.ActiveDir()} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return nil, err
		}
	}
	sequence, degraded, err := recoverActive(layout, clock, config, 0)
	if err != nil {
		return nil, err
	}
	if err := enforceRetention(layout, clock.Now(), config.RetentionMaxAge, config.RetentionMaxBytes); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(layout.ActivePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	return &Store{config: config, layout: layout, clock: clock, writer: newWriter(file, config.FlushInterval), activeBytes: info.Size(), segmentStart: clock.Now(), sequence: sequence, degraded: degraded}, nil
}

func (store *Store) unavailable() error {
	if store.writer == nil {
		return errors.New("jsonlstore: writer is unavailable")
	}
	if store.writerErr != nil {
		return fmt.Errorf("jsonlstore: writer unavailable: %w", store.writerErr)
	}
	if store.writer.closed {
		return errors.New("jsonlstore: writer is closed")
	}
	return nil
}

func (store *Store) AppendBatch(ctx context.Context, records []storage.Record) error {
	encoded := make([][]byte, len(records))
	var incomingBytes int64
	for index, record := range records {
		if err := contextError(ctx); err != nil {
			return err
		}
		line, err := EncodeRecord(record)
		if err != nil {
			return err
		}
		encoded[index] = line
		incomingBytes += int64(len(line))
	}
	if err := contextError(ctx); err != nil {
		return err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.closed {
		return storage.ErrClosed
	}
	if err := store.unavailable(); err != nil {
		return err
	}
	if len(encoded) == 0 {
		return nil
	}
	now := store.clock.Now()
	if store.activeBytes > 0 && (store.activeBytes+incomingBytes > store.config.SegmentMaxBytes || now.Sub(store.segmentStart) >= store.config.SegmentMaxAge) {
		if err := store.rotateLocked(now, true); err != nil {
			return err
		}
	}
	for _, line := range encoded {
		if err := store.writer.write(line); err != nil {
			return fmt.Errorf("jsonlstore: append active segment: %w", err)
		}
		store.activeBytes += int64(len(line))
	}
	return nil
}

func (store *Store) Flush() error {
	store.mu.RLock()
	defer store.mu.RUnlock()
	if store.closed {
		return storage.ErrClosed
	}
	if err := store.unavailable(); err != nil {
		return err
	}
	return store.writer.flush()
}

func (store *Store) Close() error {
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.closed {
		return store.closeErr
	}
	store.closed = true
	if store.writer == nil {
		store.closeErr = store.writerErr
		return store.closeErr
	}
	if store.activeBytes == 0 {
		store.closeErr = store.writer.close()
		if store.closeErr == nil {
			store.closeErr = store.removeEmptyActiveLocked()
		}
		return store.closeErr
	}
	store.closeErr = store.rotateLocked(store.clock.Now(), false)
	return store.closeErr
}

func contextError(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
