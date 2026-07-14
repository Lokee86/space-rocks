package servicelog

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"time"
)

func recoverInterruptedSegment(policy FilePolicy, deps runtimeDependencies, activePath string, now time.Time) error {
	data, err := readFileBytes(deps, activePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	complete := completeJSONLPrefix(data)
	if len(complete) == 0 {
		if err := removePath(deps, activePath); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	info, err := statPath(deps, activePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	archivePath := archivePathForSegment(policy, info.ModTime().UTC(), now)
	archivePath, err = selectArchivePath(deps, archivePath)
	if err != nil {
		return err
	}
	if err := writeRecoveredArchive(policy, deps, archivePath, complete); err != nil {
		return err
	}
	if err := removePath(deps, activePath); err != nil {
		return err
	}
	return nil
}

func writeRecoveredArchive(policy FilePolicy, deps runtimeDependencies, archivePath string, payload []byte) error {
	if err := mkdirPath(deps, filepath.Dir(archivePath)); err != nil {
		return err
	}

	archiveFile, err := openWriteCloser(deps, archivePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(archiveFile, bytes.NewReader(payload)); err != nil {
		_ = archiveFile.Close()
		_ = removePath(deps, archivePath)
		return err
	}
	if err := archiveFile.Close(); err != nil {
		_ = removePath(deps, archivePath)
		return err
	}

	if policy.CompressionEnabled {
		gzPath := archivePath + ".gz"
		if err := compressRotatedSegment(archivePath, gzPath); err != nil {
			_ = removePath(deps, archivePath)
			_ = removePath(deps, gzPath)
			return err
		}
		if err := removePath(deps, archivePath); err != nil {
			_ = removePath(deps, gzPath)
			return err
		}
	}

	return nil
}

func completeJSONLPrefix(data []byte) []byte {
	lastNewline := bytes.LastIndexByte(data, '\n')
	if lastNewline < 0 {
		return nil
	}
	return data[:lastNewline+1]
}
