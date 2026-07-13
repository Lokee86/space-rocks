package filestore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/audit"
)

const DefaultMaxRecordBytes int64 = 4 << 20

var (
	ErrInvalidRoot = errors.New("audit filestore: root directory is required")
	ErrDuplicateRecord = errors.New("audit filestore: record already exists")
	ErrRecordTooLarge = errors.New("audit filestore: record exceeds byte limit")
	ErrInvalidStoredData = errors.New("audit filestore: invalid stored record")
)

type Store struct { root string; maxBytes int64 }

func New(root string, maxBytes int64) (*Store, error) {
	if root == "" { return nil, ErrInvalidRoot }
	if maxBytes == 0 { maxBytes = DefaultMaxRecordBytes }
	if maxBytes < 1 { return nil, ErrRecordTooLarge }
	return &Store{root: root, maxBytes: maxBytes}, nil
}

var _ audit.Store = (*Store)(nil)

func (s *Store) Save(ctx context.Context, record audit.Record) error {
	if err := contextError(ctx); err != nil { return err }
	if s == nil || s.root == "" { return ErrInvalidRoot }
	if !validUUID(record.AuditEventID) { return audit.ErrInvalidAuditEventID }
	encoded, err := json.Marshal(record)
	if err != nil { return err }
	if len(encoded) == 0 || int64(len(encoded)) > s.maxBytes { return ErrRecordTooLarge }
	if err := os.MkdirAll(s.root, 0o755); err != nil { return err }
	path := filepath.Join(s.root, record.AuditEventID+".json")
	file, err := os.CreateTemp(s.root, ".audit-*.tmp")
	if err != nil { return err }
	temporary := file.Name()
	keepTemporary := true
	defer func() { if keepTemporary { _ = os.Remove(temporary) } }()
	if _, err = file.Write(encoded); err == nil { err = file.Sync() }
	if closeErr := file.Close(); err == nil { err = closeErr }
	if err != nil { return err }
	if err := contextError(ctx); err != nil { return err }
	if err := os.Link(temporary, path); err != nil {
		if errors.Is(err, os.ErrExist) { return ErrDuplicateRecord }
		return err
	}
	if err := os.Remove(temporary); err != nil { return err }
	keepTemporary = false
	return nil
}

func (s *Store) Get(ctx context.Context, id string) (audit.Record, error) {
	if err := contextError(ctx); err != nil { return audit.Record{}, err }
	if s == nil || s.root == "" { return audit.Record{}, ErrInvalidRoot }
	if !validUUID(id) { return audit.Record{}, audit.ErrInvalidAuditEventID }
	file, err := os.Open(filepath.Join(s.root, id+".json"))
	if errors.Is(err, os.ErrNotExist) { return audit.Record{}, &audit.RecordNotFoundError{AuditEventID: id} }
	if err != nil { return audit.Record{}, err }
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, s.maxBytes+1))
	if err != nil { return audit.Record{}, err }
	if int64(len(data)) > s.maxBytes { return audit.Record{}, ErrRecordTooLarge }
	var record audit.Record
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&record); err != nil { return audit.Record{}, fmt.Errorf("%w: %v", ErrInvalidStoredData, err) }
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF { return audit.Record{}, ErrInvalidStoredData }
	if record.AuditEventID != id { return audit.Record{}, ErrInvalidStoredData }
	return record, nil
}

func contextError(ctx context.Context) error {
	if ctx == nil { return nil }
	select { case <-ctx.Done(): return ctx.Err(); default: return nil }
}

func validUUID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' { return false }
	for index, character := range value {
		if index == 8 || index == 13 || index == 18 || index == 23 { continue }
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f') || (character >= 'A' && character <= 'F')) { return false }
	}
	return true
}
