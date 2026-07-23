package servicelog

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var errRollingWriterClosed = errors.New("servicelog: rolling writer closed")

// rollingWriter serializes JSONL record writes, rotates the active file when
// the configured size or age limit is reached, and preserves the active path
// expected by the runtime.
type rollingWriter struct {
	mu sync.Mutex

	deps   runtimeDependencies
	policy FilePolicy

	file         io.WriteCloser
	activePath   string
	segmentStart time.Time
	segmentBytes int64
	pending      []byte
	closed       bool
}

func newRollingWriter(policy FilePolicy, dependencies runtimeDependencies) (*rollingWriter, error) {
	if err := dependencies.mkdir(policy.Directory, 0o755); err != nil {
		return nil, err
	}

	activePath := filepath.Join(policy.Directory, policy.Prefix+".jsonl.open")
	now := nowTime(dependencies)
	if err := recoverInterruptedSegment(policy, dependencies, activePath, now.UTC()); err != nil {
		return nil, err
	}

	file, err := openWriteCloser(dependencies, activePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	if err := enforceArchiveRetention(policy, dependencies, now.UTC()); err != nil {
		_ = file.Close()
		return nil, err
	}

	return &rollingWriter{
		deps:         dependencies,
		policy:       policy,
		file:         file,
		activePath:   activePath,
		segmentStart: now.UTC(),
	}, nil
}

func (w *rollingWriter) reportFailure(err error) {
	if err != nil && w.deps.reportFailure != nil {
		w.deps.reportFailure(err)
	}
}

func (w *rollingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return 0, errRollingWriterClosed
	}

	w.pending = append(w.pending, p...)
	for {
		idx := bytes.IndexByte(w.pending, '\n')
		if idx < 0 {
			return len(p), nil
		}

		record := append([]byte(nil), w.pending[:idx+1]...)
		w.pending = w.pending[idx+1:]
		if err := w.writeRecordLocked(record); err != nil {
			w.reportFailure(err)
			return 0, err
		}
	}
}

func (w *rollingWriter) writeRecordLocked(record []byte) error {
	if w.closed {
		return errRollingWriterClosed
	}

	now := nowTime(w.deps).UTC()
	if w.shouldRotateLocked(int64(len(record)), now) {
		if err := w.rotateLocked(now); err != nil {
			return err
		}
	}

	n, err := w.file.Write(record)
	w.segmentBytes += int64(n)
	if err != nil {
		return err
	}
	if n != len(record) {
		return io.ErrShortWrite
	}
	return nil
}

func (w *rollingWriter) shouldRotateLocked(recordBytes int64, now time.Time) bool {
	if w.segmentBytes == 0 {
		return false
	}
	if w.segmentBytes+recordBytes > w.policy.SegmentMaxBytes {
		return true
	}
	if now.Sub(w.segmentStart) >= w.policy.SegmentMaxAge {
		return true
	}
	return false
}

func (w *rollingWriter) rotateLocked(now time.Time) error {
	if w.file == nil {
		return errRollingWriterClosed
	}

	segmentStart := w.segmentStart
	segmentEnd := now.UTC()
	archivePath := archivePathForSegment(w.policy, segmentStart, segmentEnd)
	if err := w.deps.mkdir(filepath.Dir(archivePath), 0o755); err != nil {
		return err
	}
	selectedArchivePath, err := selectArchivePath(w.deps, archivePath)
	if err != nil {
		return err
	}
	archivePath = selectedArchivePath

	if err := w.file.Close(); err != nil {
		return err
	}
	if err := w.deps.rename(w.activePath, archivePath); err != nil {
		return err
	}

	if w.policy.CompressionEnabled {
		if err := compressRotatedSegment(archivePath, archivePath+".gz"); err != nil {
			return err
		}
		if err := w.deps.remove(archivePath); err != nil {
			return err
		}
	}

	file, err := openWriteCloser(w.deps, w.activePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	w.file = file
	w.segmentStart = now.UTC()
	w.segmentBytes = 0
	if err := enforceArchiveRetention(w.policy, w.deps, now.UTC()); err != nil {
		return err
	}
	return nil
}

func (w *rollingWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed || w.file == nil {
		return nil
	}
	syncer, ok := w.file.(interface{ Sync() error })
	if !ok {
		return nil
	}
	if err := syncer.Sync(); err != nil {
		w.reportFailure(err)
		return err
	}
	return nil
}

func (w *rollingWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}
	w.closed = true

	if w.file == nil {
		return nil
	}
	if err := w.file.Close(); err != nil {
		w.reportFailure(err)
		return err
	}
	return nil
}
