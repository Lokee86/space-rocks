package servicelog

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const archiveTimestampLayout = "20060102T150405.000000000Z"

func archiveDirectoryForTime(root string, moment time.Time) string {
	utc := moment.UTC()
	return filepath.Join(
		root,
		"archive",
		fmt.Sprintf("%04d", utc.Year()),
		fmt.Sprintf("%02d", int(utc.Month())),
		fmt.Sprintf("%02d", utc.Day()),
	)
}

func archiveSegmentName(prefix string, startedAt, endedAt time.Time, sequence int) string {
	return fmt.Sprintf(
		"%s-%s-%s-%04d.jsonl",
		prefix,
		startedAt.UTC().Format(archiveTimestampLayout),
		endedAt.UTC().Format(archiveTimestampLayout),
		sequence,
	)
}

func compressedArchiveSegmentName(prefix string, startedAt, endedAt time.Time, sequence int) string {
	return archiveSegmentName(prefix, startedAt, endedAt, sequence) + ".gz"
}

func archiveSegmentPath(root, prefix string, startedAt, endedAt time.Time, sequence int) string {
	return filepath.Join(archiveDirectoryForTime(root, endedAt), archiveSegmentName(prefix, startedAt, endedAt, sequence))
}

func compressedArchiveSegmentPath(root, prefix string, startedAt, endedAt time.Time, sequence int) string {
	return filepath.Join(archiveDirectoryForTime(root, endedAt), compressedArchiveSegmentName(prefix, startedAt, endedAt, sequence))
}

func nextArchiveSegmentPath(root, prefix string, startedAt, endedAt time.Time, stat func(string) (os.FileInfo, error)) (string, error) {
	for sequence := 1; ; sequence++ {
		path := archiveSegmentPath(root, prefix, startedAt, endedAt, sequence)
		if _, err := stat(path); err != nil {
			if os.IsNotExist(err) {
				return path, nil
			}
			return "", err
		}
	}
}

func nextCompressedArchiveSegmentPath(root, prefix string, startedAt, endedAt time.Time, stat func(string) (os.FileInfo, error)) (string, error) {
	for sequence := 1; ; sequence++ {
		path := compressedArchiveSegmentPath(root, prefix, startedAt, endedAt, sequence)
		if _, err := stat(path); err != nil {
			if os.IsNotExist(err) {
				return path, nil
			}
			return "", err
		}
	}
}
