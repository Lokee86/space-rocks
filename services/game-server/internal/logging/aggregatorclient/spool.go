package aggregatorclient

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	spoolFilePrefix = "aggregator-batch-"
	spoolFileSuffix = ".batch"
)

type spoolStore struct {
	directory string
	byteCap   int64
}

type pendingBatch struct {
	Path string
	Size int64
}

type spoolSaveResult struct {
	Stored         bool
	EvictedBatches int
	EvictedEvents  int
}

func newSpoolStore(config Config) *spoolStore {
	return &spoolStore{directory: config.SpoolDirectory, byteCap: config.SpoolByteCap}
}

func (s *spoolStore) save(batch []byte) (spoolSaveResult, error) {
	if int64(len(batch)) > s.byteCap {
		return spoolSaveResult{}, fmt.Errorf("batch size %d exceeds spool byte cap %d", len(batch), s.byteCap)
	}
	if err := os.MkdirAll(s.directory, 0o755); err != nil {
		return spoolSaveResult{}, err
	}
	name, err := uniqueSpoolName()
	if err != nil {
		return spoolSaveResult{}, err
	}
	temporary, err := os.CreateTemp(s.directory, ".aggregator-batch-*.tmp")
	if err != nil {
		return spoolSaveResult{}, err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if _, err := temporary.Write(batch); err != nil {
		temporary.Close()
		return spoolSaveResult{}, err
	}
	if err := temporary.Close(); err != nil {
		return spoolSaveResult{}, err
	}
	if err := os.Rename(temporaryName, filepath.Join(s.directory, name)); err != nil {
		return spoolSaveResult{}, err
	}
	result, err := s.evict()
	result.Stored = true
	return result, err
}

func (s *spoolStore) pending() ([]pendingBatch, error) {
	entries, err := os.ReadDir(s.directory)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	batches := make([]pendingBatch, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !isSpoolFile(entry.Name()) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		batches = append(batches, pendingBatch{Path: filepath.Join(s.directory, entry.Name()), Size: info.Size()})
	}
	sort.Slice(batches, func(i, j int) bool {
		return filepath.Base(batches[i].Path) < filepath.Base(batches[j].Path)
	})
	return batches, nil
}

func (s *spoolStore) remove(batch pendingBatch) error {
	return os.Remove(batch.Path)
}

func (s *spoolStore) load(batch pendingBatch) ([]byte, int, error) {
	if batch.Size > s.byteCap {
		return nil, 0, fmt.Errorf("pending batch size %d exceeds spool byte cap %d", batch.Size, s.byteCap)
	}
	file, err := os.Open(batch.Path)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()
	payload, err := io.ReadAll(io.LimitReader(file, s.byteCap+1))
	if err != nil {
		return nil, 0, err
	}
	if int64(len(payload)) > s.byteCap {
		return nil, 0, fmt.Errorf("pending batch exceeds spool byte cap %d", s.byteCap)
	}
	var decoded encodedBatch
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return nil, 0, fmt.Errorf("decode pending batch: %w", err)
	}
	return payload, len(decoded.Events), nil
}

func (s *spoolStore) evict() (spoolSaveResult, error) {
	batches, err := s.pending()
	if err != nil {
		return spoolSaveResult{}, err
	}
	var total int64
	for _, batch := range batches {
		total += batch.Size
	}
	var result spoolSaveResult
	for total > s.byteCap && len(batches) > 0 {
		oldest := batches[0]
		evictedEvents := encodedBatchEventCount(oldest.Path)
		if err := s.remove(oldest); err != nil {
			return result, err
		}
		result.EvictedBatches++
		result.EvictedEvents += evictedEvents
		total -= oldest.Size
		batches = batches[1:]
	}
	return result, nil
}

func encodedBatchEventCount(path string) int {
	payload, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	var batch encodedBatch
	if json.Unmarshal(payload, &batch) != nil {
		return 0
	}
	return len(batch.Events)
}

func isSpoolFile(name string) bool {
	return strings.HasPrefix(name, spoolFilePrefix) && strings.HasSuffix(name, spoolFileSuffix)
}

func uniqueSpoolName() (string, error) {
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", err
	}
	return spoolFilePrefix + strconv.FormatInt(time.Now().UnixNano(), 10) + "-" + fmt.Sprintf("%x", random[:]) + spoolFileSuffix, nil
}
