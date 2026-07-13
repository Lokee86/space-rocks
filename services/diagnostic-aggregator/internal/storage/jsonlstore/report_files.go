package jsonlstore

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func reportSegmentPaths(layout reportLayout) ([]string, error) {
	var paths []string
	err := filepath.Walk(layout.archiveDir(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".jsonl") || strings.HasSuffix(path, ".jsonl.gz") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	sort.Strings(paths)
	paths = append(paths, layout.activePath())
	return paths, nil
}

func openReportSegment(path string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(path, ".gz") {
		return file, nil
	}
	reader, err := gzip.NewReader(file)
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("jsonlstore: open compressed report segment: %w", err)
	}
	return &reportReadCloser{Reader: reader, close: func() error { _ = reader.Close(); return file.Close() }}, nil
}

type reportReadCloser struct {
	io.Reader
	close func() error
}

func (r *reportReadCloser) Close() error { return r.close() }
