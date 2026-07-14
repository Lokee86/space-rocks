package servicelog

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type retentionCandidate struct {
	path    string
	size    int64
	modTime time.Time
}

func enforceArchiveRetention(policy FilePolicy, deps runtimeDependencies, now time.Time) error {
	archiveRoot := filepath.Join(policy.Directory, "archive")
	candidates, err := collectRetentionCandidates(deps, archiveRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(candidates) == 0 {
		return nil
	}

	sortRetentionCandidates(candidates)
	kept := candidates[:0]
	for _, candidate := range candidates {
		if now.Sub(candidate.modTime) > policy.RetentionMaxAge {
			if err := removePath(deps, candidate.path); err != nil && !os.IsNotExist(err) {
				return err
			}
			continue
		}
		kept = append(kept, candidate)
	}

	var remainingBytes int64
	for _, candidate := range kept {
		remainingBytes += candidate.size
	}
	if remainingBytes <= policy.RetentionMaxBytes {
		return nil
	}

	for _, candidate := range kept {
		if remainingBytes <= policy.RetentionMaxBytes {
			break
		}
		if err := removePath(deps, candidate.path); err != nil && !os.IsNotExist(err) {
			return err
		}
		remainingBytes -= candidate.size
	}
	return nil
}

func collectRetentionCandidates(deps runtimeDependencies, root string) ([]retentionCandidate, error) {
	entries, err := readDirEntries(deps, root)
	if err != nil {
		return nil, err
	}

	candidates := make([]retentionCandidate, 0, len(entries))
	for _, entry := range entries {
		path := filepath.Join(root, entry.Name())
		if entry.IsDir() {
			subCandidates, err := collectRetentionCandidates(deps, path)
			if err != nil {
				return nil, err
			}
			candidates = append(candidates, subCandidates...)
			continue
		}
		if !isRetentionCandidatePath(path) {
			continue
		}
		info, err := statPath(deps, path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		candidates = append(candidates, retentionCandidate{
			path:    path,
			size:    info.Size(),
			modTime: info.ModTime().UTC(),
		})
	}
	return candidates, nil
}

func readDirEntries(deps runtimeDependencies, path string) ([]os.DirEntry, error) {
	if deps.readDir != nil {
		return deps.readDir(path)
	}
	return os.ReadDir(path)
}

func sortRetentionCandidates(candidates []retentionCandidate) {
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].modTime.Equal(candidates[j].modTime) {
			return candidates[i].path < candidates[j].path
		}
		return candidates[i].modTime.Before(candidates[j].modTime)
	})
}

func isRetentionCandidatePath(path string) bool {
	base := filepath.Base(path)
	return strings.HasSuffix(base, ".jsonl") || strings.HasSuffix(base, ".jsonl.gz")
}
