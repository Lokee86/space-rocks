package servicelog

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const rollingWriterBufferSize = 64 * 1024

// rollingWriter owns the active JSONL segment for a single service log stream.
// It keeps one <prefix>.jsonl.open file under active/ and rotates completed
// segments into archive/YYYY/MM/DD/ by size or age.
type rollingWriter struct {
	mu               sync.Mutex
	deps             runtimeDependencies
	policy           FilePolicy
	file             *os.File
	buffered         *bufio.Writer
	activePath       string
	segmentStartedAt time.Time
	bytesWritten     int64
	status           statusTracker
	closed           bool
}

func newRollingWriter(policy FilePolicy, dependencies runtimeDependencies) (*rollingWriter, error) {
	if dependencies.mkdir == nil {
		dependencies = defaultRuntimeDependencies()
	}

	writer := &rollingWriter{
		deps:             dependencies,
		policy:           policy,
		segmentStartedAt: dependencies.now().UTC(),
	}
	if err := writer.recoverInterruptedActiveSegmentLocked(writer.segmentStartedAt); err != nil {
		return nil, err
	}
	if err := writer.openActiveSegmentLocked(writer.segmentStartedAt); err != nil {
		return nil, err
	}
	writer.runRetentionLocked(writer.segmentStartedAt)
	return writer, nil
}

func (w *rollingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return 0, os.ErrClosed
	}
	if w.buffered == nil || w.file == nil {
		return 0, errors.New("servicelog: rolling writer is unavailable")
	}

	now := w.deps.now().UTC()
	if w.shouldRotateLocked(len(p), now) {
		if err := w.rotateActiveSegmentLocked(now, false); err != nil {
			w.recordFailureLocked()
			return 0, err
		}
	}

	n, err := w.buffered.Write(p)
	w.bytesWritten += int64(n)
	w.status.setActive(w.activePath, w.bytesWritten)
	if err == nil && n < len(p) {
		err = io.ErrShortWrite
	}
	if err != nil {
		w.recordFailureLocked()
	}
	return n, err
}

func (w *rollingWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return os.ErrClosed
	}
	if w.buffered == nil || w.file == nil {
		return errors.New("servicelog: rolling writer is unavailable")
	}
	if err := w.flushCurrentSegmentLocked(); err != nil {
		w.recordFailureLocked()
		return err
	}
	return nil
}

func (w *rollingWriter) Maintain(now time.Time) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return os.ErrClosed
	}
	if w.buffered == nil || w.file == nil {
		return errors.New("servicelog: rolling writer is unavailable")
	}
	if err := w.flushCurrentSegmentLocked(); err != nil {
		w.recordFailureLocked()
		return err
	}
	if w.shouldRotateForMaintenanceLocked(now) {
		if err := w.rotateActiveSegmentLocked(now, true); err != nil {
			w.recordFailureLocked()
			return err
		}
	}
	w.runRetentionLocked(now)
	return nil
}

func (w *rollingWriter) Close() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	w.closed = true
	buffered := w.buffered
	file := w.file
	w.buffered = nil
	w.file = nil
	w.mu.Unlock()

	if buffered == nil || file == nil {
		return nil
	}

	if err := buffered.Flush(); err != nil {
		w.recordFailureLocked()
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		w.recordFailureLocked()
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		w.recordFailureLocked()
		return err
	}
	return nil
}

func (w *rollingWriter) shouldRotateLocked(incomingBytes int, now time.Time) bool {
	if w.bytesWritten == 0 {
		return false
	}
	if w.policy.SegmentMaxBytes > 0 && w.bytesWritten+int64(incomingBytes) > w.policy.SegmentMaxBytes {
		return true
	}
	if w.policy.SegmentMaxAge > 0 && now.Sub(w.segmentStartedAt) >= w.policy.SegmentMaxAge {
		return true
	}
	return false
}

func (w *rollingWriter) shouldRotateForMaintenanceLocked(now time.Time) bool {
	if w.bytesWritten == 0 {
		return false
	}
	if w.policy.SegmentMaxAge > 0 && now.Sub(w.segmentStartedAt) >= w.policy.SegmentMaxAge {
		return true
	}
	return false
}

func (w *rollingWriter) flushCurrentSegmentLocked() error {
	if err := w.buffered.Flush(); err != nil {
		return err
	}
	if err := w.file.Sync(); err != nil {
		return err
	}
	return nil
}

func (w *rollingWriter) activeSegmentPath() string {
	return filepath.Join(w.policy.Directory, "active", w.policy.Prefix+".jsonl.open")
}

func (w *rollingWriter) statusSnapshot() Status {
	return w.status.snapshot()
}

func (w *rollingWriter) recordFailureLocked() {
	w.status.addFailure()
}

func (w *rollingWriter) runRetentionLocked(now time.Time) {
	if err := enforceArchiveRetention(w.policy, w.deps, now); err != nil {
		w.recordFailureLocked()
	}
}

func (w *rollingWriter) openActiveSegmentLocked(startedAt time.Time) error {
	activeDir := filepath.Join(w.policy.Directory, "active")
	if err := w.deps.mkdir(activeDir, 0o755); err != nil {
		return err
	}

	path := filepath.Join(activeDir, w.policy.Prefix+".jsonl.open")
	file, err := w.deps.openFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}

	w.file = file
	w.buffered = bufio.NewWriterSize(file, rollingWriterBufferSize)
	w.activePath = path
	w.segmentStartedAt = startedAt
	w.bytesWritten = 0
	w.status.setFileEnabled(true)
	w.status.setActive(path, 0)
	return nil
}

func (w *rollingWriter) rotateActiveSegmentLocked(rotatedAt time.Time, alreadyFlushed bool) error {
	if w.buffered == nil || w.file == nil {
		return errors.New("servicelog: rolling writer is unavailable")
	}

	buffered := w.buffered
	file := w.file
	activePath := w.activePath
	startedAt := w.segmentStartedAt
	w.buffered = nil
	w.file = nil

	if !alreadyFlushed {
		if err := buffered.Flush(); err != nil {
			return err
		}
		if err := file.Sync(); err != nil {
			return err
		}
	}
	if err := file.Close(); err != nil {
		return err
	}

	archivePath, err := nextArchiveSegmentPath(w.policy.Directory, w.policy.Prefix, startedAt, rotatedAt, w.deps.stat)
	if err != nil {
		return err
	}
	if err := w.deps.mkdir(filepath.Dir(archivePath), 0o755); err != nil {
		return err
	}
	if err := w.deps.rename(activePath, archivePath); err != nil {
		return err
	}

	if w.policy.CompressionEnabled {
		compressedPath, err := nextCompressedArchiveSegmentPath(w.policy.Directory, w.policy.Prefix, startedAt, rotatedAt, w.deps.stat)
		if err != nil {
			return err
		}
		if err := compressArchivedSegment(archivePath, compressedPath, w.deps); err != nil {
			return err
		}
		if err := w.deps.remove(archivePath); err != nil {
			return err
		}
	}

	if err := w.openActiveSegmentLocked(rotatedAt); err != nil {
		return err
	}
	w.runRetentionLocked(rotatedAt)
	return nil
}
