package jsonlstore

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// recoverActive repairs or archives an interrupted active segment before a
// writer is opened. The degraded result means corruption was quarantined.
func recoverActive(layout Layout, clock Clock, config Config, sequence uint64) (uint64, bool, error) {
	path := layout.ActivePath()
	if err := os.MkdirAll(layout.QuarantineDir(), 0o755); err != nil {
		return sequence, false, fmt.Errorf("jsonlstore: create recovery quarantine directory: %w", err)
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return sequence, false, nil
	}
	if err != nil {
		return sequence, false, fmt.Errorf("jsonlstore: read active segment for recovery: %w", err)
	}
	if len(data) == 0 {
		return sequence, false, nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return sequence, false, fmt.Errorf("jsonlstore: stat active segment for recovery: %w", err)
	}

	lines := bytes.Split(data, []byte{'\n'})
	complete := len(lines)
	incomplete := len(lines) > 0 && len(lines[len(lines)-1]) > 0
	if len(lines) > 0 && (len(lines[len(lines)-1]) == 0 || incomplete) {
		complete--
	}
	for index := 0; index < complete; index++ {
		if _, err := DecodeRecord(lines[index]); err != nil {
			if index != complete-1 {
				if quarantineErr := quarantineActive(layout, path); quarantineErr != nil {
					return sequence, false, quarantineErr
				}
				return sequence, true, nil
			}
			cut := bytes.LastIndex(data, []byte{'\n'}) + 1 - len(lines[index]) - 1
			data = data[:cut]
			if err := os.WriteFile(path, data, 0o644); err != nil {
				return sequence, false, fmt.Errorf("jsonlstore: repair final active record: %w", err)
			}
			complete--
			break
		}
	}
	if incomplete {
		if complete > 0 {
			data = data[:len(data)-len(lines[len(lines)-1])]
		} else {
			data = nil
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return sequence, false, fmt.Errorf("jsonlstore: truncate incomplete active record: %w", err)
		}
	}
	if len(data) == 0 {
		return sequence, false, nil
	}
	next, err := finalizeSegment(layout, path, info.ModTime(), clock.Now(), sequence, config.Compression)
	if err != nil {
		return sequence, false, fmt.Errorf("jsonlstore: finalize recovered active segment: %w", err)
	}
	return next, false, nil
}

func quarantineActive(layout Layout, path string) error {
	name := filepath.Base(path)
	for index := 0; ; index++ {
		candidate := layout.QuarantinePath(name)
		if index > 0 {
			candidate = layout.QuarantinePath(fmt.Sprintf("%s.%d", name, index))
		}
		if _, err := os.Stat(candidate); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("jsonlstore: inspect quarantine path: %w", err)
		}
		if err := os.Rename(path, candidate); err != nil {
			return fmt.Errorf("jsonlstore: quarantine corrupt active segment: %w", err)
		}
		return nil
	}
}
