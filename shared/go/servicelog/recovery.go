package servicelog

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"time"
)

func (w *rollingWriter) recoverInterruptedActiveSegmentLocked(now time.Time) error {
	activePath := w.activeSegmentPath()
	info, err := w.deps.stat(activePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	contents, err := w.deps.readFile(activePath)
	if err != nil {
		return err
	}
	complete := completeJSONLContent(contents)
	if len(complete) == 0 {
		return w.deps.remove(activePath)
	}

	startedAt := info.ModTime().UTC()
	endedAt := now.UTC()
	archivePath, err := nextArchiveSegmentPath(w.policy.Directory, w.policy.Prefix, startedAt, endedAt, w.deps.stat)
	if err != nil {
		return err
	}
	if err := w.deps.mkdir(filepath.Dir(archivePath), 0o755); err != nil {
		return err
	}
	if err := writeRecoveredArchiveSegment(archivePath, complete, w.deps); err != nil {
		return err
	}

	if w.policy.CompressionEnabled {
		compressedPath, err := nextCompressedArchiveSegmentPath(w.policy.Directory, w.policy.Prefix, startedAt, endedAt, w.deps.stat)
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

	return w.deps.remove(activePath)
}

func completeJSONLContent(contents []byte) []byte {
	lastNewline := bytes.LastIndexByte(contents, '\n')
	if lastNewline < 0 {
		return nil
	}
	return contents[:lastNewline+1]
}

func writeRecoveredArchiveSegment(path string, contents []byte, deps runtimeDependencies) (err error) {
	file, err := deps.openFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	cleanup := true
	defer func() {
		if file != nil {
			_ = file.Close()
		}
		if cleanup {
			_ = deps.remove(path)
		}
	}()

	if _, err = io.Copy(file, bytes.NewReader(contents)); err != nil {
		return err
	}
	if err = file.Sync(); err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	file = nil
	cleanup = false
	return nil
}
