package servicelog

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func archivePathForSegment(policy FilePolicy, segmentStart, segmentEnd time.Time) string {
	archiveDir := filepath.Join(policy.Directory, "archive", segmentEnd.Format(time.DateOnly))
	return filepath.Join(archiveDir, archiveFilename(policy.Prefix, segmentStart, segmentEnd))
}

func selectArchivePath(deps runtimeDependencies, basePath string) (string, error) {
	candidate := basePath
	stem := strings.TrimSuffix(basePath, filepath.Ext(basePath))
	ext := filepath.Ext(basePath)
	for suffix := 2; ; suffix++ {
		_, err := statPath(deps, candidate)
		if os.IsNotExist(err) {
			return candidate, nil
		}
		if err != nil {
			return "", err
		}
		candidate = fmt.Sprintf("%s-%d%s", stem, suffix, ext)
	}
}

func mkdirPath(deps runtimeDependencies, path string) error {
	if deps.mkdir != nil {
		return deps.mkdir(path, 0o755)
	}
	return os.MkdirAll(path, 0o755)
}

func readFileBytes(deps runtimeDependencies, path string) ([]byte, error) {
	if deps.readFile != nil {
		return deps.readFile(path)
	}
	return os.ReadFile(path)
}

func statPath(deps runtimeDependencies, path string) (os.FileInfo, error) {
	if deps.stat != nil {
		return deps.stat(path)
	}
	return os.Stat(path)
}

func removePath(deps runtimeDependencies, path string) error {
	if deps.remove != nil {
		return deps.remove(path)
	}
	return os.Remove(path)
}

func openWriteCloser(deps runtimeDependencies, path string, flag int, perm os.FileMode) (io.WriteCloser, error) {
	if deps.openFile != nil {
		return deps.openFile(path, flag, perm)
	}
	return os.OpenFile(path, flag, perm)
}

func nowTime(deps runtimeDependencies) time.Time {
	if deps.now != nil {
		return deps.now()
	}
	return time.Now()
}

func compressRotatedSegment(sourcePath, compressedPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.OpenFile(compressedPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	gzipWriter := gzip.NewWriter(dest)
	_, copyErr := io.Copy(gzipWriter, source)
	closeErr := gzipWriter.Close()
	destCloseErr := dest.Close()
	if copyErr != nil {
		_ = os.Remove(compressedPath)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(compressedPath)
		return closeErr
	}
	if destCloseErr != nil {
		_ = os.Remove(compressedPath)
		return destCloseErr
	}
	return nil
}

func archiveFilename(prefix string, segmentStart, segmentEnd time.Time) string {
	return prefix + "-" + segmentStart.UTC().Format("20060102T150405.000000000Z") + "-" + segmentEnd.UTC().Format("20060102T150405.000000000Z") + ".jsonl"
}
