package jsonlstore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func recoverReportActive(layout reportLayout, config Config, now time.Time) (uint64, error) {
	path := layout.activePath()
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("jsonlstore: read active reports: %w", err)
	}
	last := 0
	for i := len(data) - 1; i >= 0; i-- {
		if data[i] == '\n' {
			last = i + 1
			break
		}
	}
	if last < len(data) {
		trailing := data[last:]
		if _, decodeErr := decodeReport(trailing); decodeErr != nil {
			if !json.Valid(trailing) {
				if err := os.Truncate(path, int64(last)); err != nil {
					return 0, err
				}
				data = data[:last]
			} else {
				_, qerr := quarantineReportActive(layout, path, now)
				if qerr != nil {
					return 0, qerr
				}
				return 0, nil
			}
		}
	}
	for _, line := range splitReportLines(data) {
		if _, err := decodeReport(line); err != nil {
			_, qerr := quarantineReportActive(layout, path, now)
			if qerr != nil {
				return 0, qerr
			}
			return 0, nil
		}
	}
	if len(data) == 0 {
		return 0, nil
	}
	start := info.ModTime()
	sequence := uint64(0)
	for {
		archive := layout.archivePath(start, now, sequence, config.Compression)
		if _, err := os.Stat(archive); os.IsNotExist(err) {
			if err := finalizeArchive(path, archive, config.Compression); err != nil {
				return 0, err
			}
			return sequence + 1, nil
		} else if err != nil {
			return 0, err
		}
		sequence++
	}
}

func splitReportLines(data []byte) [][]byte {
	var lines [][]byte
	for len(data) > 0 {
		i := 0
		for i < len(data) && data[i] != '\n' {
			i++
		}
		line := data[:i]
		if len(line) > 0 {
			lines = append(lines, line)
		}
		if i == len(data) {
			break
		}
		data = data[i+1:]
	}
	return lines
}

func quarantineReportActive(layout reportLayout, path string, now time.Time) (string, error) {
	for sequence := uint64(0); ; sequence++ {
		name := fmt.Sprintf("diagnostic-reports-%s-%06d.jsonl.quarantine", formatTimestamp(now), sequence)
		destination := filepath.Join(layout.quarantineDir(), name)
		if _, err := os.Stat(destination); os.IsNotExist(err) {
			if err := os.Rename(path, destination); err != nil {
				return "", err
			}
			return destination, nil
		} else if err != nil {
			return "", err
		}
	}
}
