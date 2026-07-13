package jsonlstore

import (
	"fmt"
	"os"
	"time"
)

func (store *ReportStore) rotateReportLocked(end time.Time) error {
	if store.activeBytes == 0 {
		return nil
	}
	if err := store.writer.close(); err != nil {
		return err
	}
	for {
		archive := store.layout.archivePath(store.segmentStart, end, store.sequence, store.config.Compression)
		if _, err := os.Stat(archive); os.IsNotExist(err) {
			if err := finalizeArchive(store.path, archive, store.config.Compression); err != nil {
				return err
			}
			break
		} else if err != nil {
			return fmt.Errorf("jsonlstore: inspect report archive: %w", err)
		}
		store.sequence++
	}
	store.sequence++
	file, err := os.OpenFile(store.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("jsonlstore: open active report segment: %w", err)
	}
	store.writer = newWriter(file, store.config.FlushInterval)
	store.activeBytes = 0
	store.segmentStart = end
	return nil
}

func (store *ReportStore) finalizeReportLocked() error {
	if store.activeBytes == 0 {
		if err := store.writer.close(); err != nil {
			return err
		}
		if err := os.Remove(store.path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	if err := store.rotateReportLocked(store.now()); err != nil {
		return err
	}
	if err := store.writer.close(); err != nil {
		return err
	}
	return os.Remove(store.path)
}
