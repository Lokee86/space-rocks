package servicelog

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var errRollingWriterClosed = errors.New("servicelog: rolling writer is closed")

// rollingJSONLWriter owns the active JSONL sink and rotates the active segment
// before each completed record when the size or age policy is hit.
type rollingJSONLWriter struct {
	deps         runtimeDependencies
	directory    string
	prefix       string
	activePath   string
	file         io.WriteCloser
	segmentStart time.Time
	activeSize   int64
	maxBytes     int64
	maxAge       time.Duration

	mu        sync.Mutex
	closeOnce sync.Once
	closeErr  error
	closed    bool
}

func newRollingJSONLWriter(config Config, dependencies runtimeDependencies) (*rollingJSONLWriter, error) {
	activePath := filepath.Join(config.File.Directory, config.File.Prefix+".jsonl.open")

	file, err := dependencies.openFile(activePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	activeSize := int64(0)
	if info, err := dependencies.stat(activePath); err == nil {
		activeSize = info.Size()
	} else if !errors.Is(err, os.ErrNotExist) {
		_ = file.Close()
		return nil, err
	}

	return &rollingJSONLWriter{
		deps:         dependencies,
		directory:    config.File.Directory,
		prefix:       config.File.Prefix,
		activePath:   activePath,
		file:         file,
		segmentStart: dependencies.now().UTC(),
		activeSize:   activeSize,
		maxBytes:     config.File.SegmentMaxBytes,
		maxAge:       config.File.SegmentMaxAge,
	}, nil
}

func (w *rollingJSONLWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return 0, errRollingWriterClosed
	}

	remaining := p
	for len(remaining) > 0 {
		nextLine := len(remaining)
		if newline := indexByte(remaining, '\n'); newline >= 0 {
			nextLine = newline + 1
		}

		record := remaining[:nextLine]
		remaining = remaining[nextLine:]

		if err := w.rotateIfNeeded(int64(len(record))); err != nil {
			return 0, err
		}

		if len(record) == 0 {
			continue
		}

		n, err := w.file.Write(record)
		if err != nil {
			return 0, err
		}
		if n != len(record) {
			return 0, io.ErrShortWrite
		}
		w.activeSize += int64(n)
	}

	return len(p), nil
}

func (w *rollingJSONLWriter) Close() error {
	didClose := false
	w.closeOnce.Do(func() {
		didClose = true
		w.mu.Lock()
		w.closed = true
		file := w.file
		w.file = nil
		w.mu.Unlock()

		if file != nil {
			w.closeErr = file.Close()
		}
	})
	if didClose {
		return w.closeErr
	}
	return nil
}

func (w *rollingJSONLWriter) rotateIfNeeded(recordSize int64) error {
	now := w.deps.now().UTC()
	if w.activeSize > 0 && w.maxBytes > 0 && w.activeSize+recordSize > w.maxBytes {
		return w.rotate(now)
	}
	if w.maxAge > 0 && now.Sub(w.segmentStart) >= w.maxAge {
		return w.rotate(now)
	}
	return nil
}

func (w *rollingJSONLWriter) rotate(now time.Time) error {
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return err
		}
	}

	archivePath, err := w.nextArchivePath(now)
	if err != nil {
		return err
	}
	if err := w.deps.rename(w.activePath, archivePath); err != nil {
		return err
	}

	file, err := w.deps.openFile(w.activePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	w.file = file
	w.segmentStart = now
	w.activeSize = 0
	return nil
}

func (w *rollingJSONLWriter) nextArchivePath(now time.Time) (string, error) {
	baseName := fmt.Sprintf("%s.%s-%s.jsonl", w.prefix, formatArchiveTimestamp(w.segmentStart), formatArchiveTimestamp(now))
	basePath := filepath.Join(w.directory, baseName)
	if _, err := w.deps.stat(basePath); errors.Is(err, os.ErrNotExist) {
		return basePath, nil
	} else if err != nil {
		return "", err
	}

	for suffix := 1; ; suffix++ {
		candidate := filepath.Join(w.directory, fmt.Sprintf("%s.%s-%s.jsonl.%d", w.prefix, formatArchiveTimestamp(w.segmentStart), formatArchiveTimestamp(now), suffix))
		if _, err := w.deps.stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		} else if err != nil {
			return "", err
		}
	}
}

func formatArchiveTimestamp(t time.Time) string {
	return t.UTC().Format("20060102T150405.000000000Z")
}

func indexByte(b []byte, c byte) int {
	for i := range b {
		if b[i] == c {
			return i
		}
	}
	return -1
}
