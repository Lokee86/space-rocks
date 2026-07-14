package servicelog

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type archiveRecord struct {
	path string
	info os.FileInfo
}

func enforceArchiveRetention(policy FilePolicy, dependencies runtimeDependencies, now time.Time) error {
	if policy.RetentionMaxAge <= 0 && policy.RetentionMaxBytes <= 0 {
		return nil
	}

	archiveRoot := filepath.Join(policy.Directory, "archive")
	records, err := collectArchiveRecords(archiveRoot, dependencies)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}

	if policy.RetentionMaxAge > 0 {
		for _, record := range records {
			if now.Sub(record.info.ModTime()) > policy.RetentionMaxAge {
				if err := dependencies.remove(record.path); err != nil {
					return err
				}
			}
		}
		records, err = collectArchiveRecords(archiveRoot, dependencies)
		if err != nil {
			return err
		}
	}

	if policy.RetentionMaxBytes <= 0 {
		return nil
	}

	sortArchiveRecords(records)
	var total int64
	for _, record := range records {
		total += record.info.Size()
	}
	for _, record := range records {
		if total <= policy.RetentionMaxBytes {
			break
		}
		if err := dependencies.remove(record.path); err != nil {
			return err
		}
		total -= record.info.Size()
	}
	return nil
}

func collectArchiveRecords(root string, dependencies runtimeDependencies) ([]archiveRecord, error) {
	entries, err := dependencies.readDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var records []archiveRecord
	for _, entry := range entries {
		path := filepath.Join(root, entry.Name())
		if entry.IsDir() {
			nested, err := collectArchiveRecords(path, dependencies)
			if err != nil {
				return nil, err
			}
			records = append(records, nested...)
			continue
		}
		if !isCompletedArchiveFile(entry.Name()) {
			continue
		}
		info, err := dependencies.stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		if info.IsDir() {
			continue
		}
		records = append(records, archiveRecord{path: path, info: info})
	}
	sortArchiveRecords(records)
	return records, nil
}

func sortArchiveRecords(records []archiveRecord) {
	sort.Slice(records, func(i, j int) bool {
		left := records[i].info.ModTime().UTC()
		right := records[j].info.ModTime().UTC()
		if !left.Equal(right) {
			return left.Before(right)
		}
		return records[i].path < records[j].path
	})
}

func isCompletedArchiveFile(name string) bool {
	return strings.HasSuffix(name, ".jsonl") || strings.HasSuffix(name, ".jsonl.gz")
}
