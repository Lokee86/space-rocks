package jsonlstore

import (
	"fmt"
	"os"
	"time"
)

func (store *Store) rotateLocked(end time.Time, reopen bool) error {
	if store.activeBytes == 0 {
		if !reopen {
			err := store.writer.close()
			if err != nil {
				store.writerErr = err
			}
			return err
		}
		return nil
	}
	start := store.segmentStart
	activePath := store.layout.ActivePath()
	if err := store.writer.close(); err != nil {
		store.writerErr = err
		return err
	}
	nextSequence, err := finalizeSegment(store.layout, activePath, start, end, store.sequence, store.config.Compression)
	if err != nil {
		store.writerErr = err
		return fmt.Errorf("jsonlstore: finalize segment: %w", err)
	}
	store.sequence = nextSequence
	store.activeBytes = 0
	if !reopen {
		if err := enforceRetention(store.layout, end, store.config.RetentionMaxAge, store.config.RetentionMaxBytes); err != nil {
			return err
		}
		return nil
	}
	if err := store.openActiveLocked(end); err != nil {
		store.writerErr = err
		return fmt.Errorf("jsonlstore: open replacement active writer: %w", err)
	}
	if err := enforceRetention(store.layout, end, store.config.RetentionMaxAge, store.config.RetentionMaxBytes); err != nil {
		return err
	}
	return nil
}

func (store *Store) openActiveLocked(start time.Time) error {
	file, err := os.OpenFile(store.layout.ActivePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	store.writer = newWriter(file, store.config.FlushInterval)
	store.writerErr = nil
	store.segmentStart = start
	return nil
}

func (store *Store) removeEmptyActiveLocked() error {
	return os.Remove(store.layout.ActivePath())
}
