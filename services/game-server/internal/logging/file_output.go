package logging

import (
	"fmt"
	"os"
	"path/filepath"
)

func openSequentialLogFile(baseDir string, prefix string) (*os.File, string, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, "", err
	}

	for sequence := 1; ; sequence++ {
		path := filepath.Join(baseDir, fmt.Sprintf("%s-%06d.jsonl", prefix, sequence))
		file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			return file, path, nil
		}

		if os.IsExist(err) {
			continue
		}

		return nil, "", err
	}
}
