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

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/diagnostics"
)

const DefaultMaxBundleBytes int64 = 4 << 20

var (
	ErrInvalidRoot       = errors.New("diagnostic filestore: root directory is required")
	ErrDuplicateBundle   = errors.New("diagnostic filestore: bundle already exists")
	ErrBundleTooLarge    = errors.New("diagnostic filestore: bundle exceeds byte limit")
	ErrInvalidStoredData  = errors.New("diagnostic filestore: invalid stored bundle")
)

type Store struct {
	root       string
	maxBytes   int64
}

func New(root string, maxBytes int64) (*Store, error) {
	if root == "" {
		return nil, ErrInvalidRoot
	}
	if maxBytes == 0 {
		maxBytes = DefaultMaxBundleBytes
	}
	if maxBytes < 1 {
		return nil, ErrBundleTooLarge
	}
	return &Store{root: root, maxBytes: maxBytes}, nil
}

var _ diagnostics.BundleStore = (*Store)(nil)

func (s *Store) Save(ctx context.Context, bundle diagnostics.Bundle) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if s == nil || s.root == "" {
		return ErrInvalidRoot
	}
	if !validUUID(bundle.DiagnosticReportID) {
		return diagnostics.ErrInvalidDiagnosticReportID
	}
	encoded, err := json.Marshal(bundle)
	if err != nil {
		return err
	}
	if len(encoded) == 0 || int64(len(encoded)) > s.maxBytes {
		return ErrBundleTooLarge
	}
	if err := os.MkdirAll(s.root, 0o755); err != nil {
		return err
	}
	path := filepath.Join(s.root, bundle.DiagnosticReportID+".json")
	temporary, err := os.CreateTemp(s.root, ".diagnostic-*.tmp")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	removeTemporary := true
	defer func() {
		if removeTemporary {
			_ = os.Remove(temporaryName)
		}
	}()
	if _, err = temporary.Write(encoded); err == nil {
		err = temporary.Sync()
	}
	if closeErr := temporary.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err := contextError(ctx); err != nil {
		return err
	}
	if err := os.Link(temporaryName, path); err != nil {
		if errors.Is(err, os.ErrExist) {
			return ErrDuplicateBundle
		}
		return err
	}
	if err := os.Remove(temporaryName); err != nil {
		return err
	}
	removeTemporary = false
	return nil
}

func (s *Store) Get(ctx context.Context, diagnosticReportID string) (diagnostics.Bundle, error) {
	if err := contextError(ctx); err != nil {
		return diagnostics.Bundle{}, err
	}
	if s == nil || s.root == "" {
		return diagnostics.Bundle{}, ErrInvalidRoot
	}
	if !validUUID(diagnosticReportID) {
		return diagnostics.Bundle{}, diagnostics.ErrInvalidDiagnosticReportID
	}
	path := filepath.Join(s.root, diagnosticReportID+".json")
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return diagnostics.Bundle{}, &diagnostics.BundleNotFoundError{DiagnosticReportID: diagnosticReportID}
	}
	if err != nil {
		return diagnostics.Bundle{}, err
	}
	defer file.Close()
	limited := io.LimitReader(file, s.maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return diagnostics.Bundle{}, err
	}
	if int64(len(data)) > s.maxBytes {
		return diagnostics.Bundle{}, ErrBundleTooLarge
	}
	var bundle diagnostics.Bundle
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&bundle); err != nil {
		return diagnostics.Bundle{}, fmt.Errorf("%w: %v", ErrInvalidStoredData, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return diagnostics.Bundle{}, ErrInvalidStoredData
	}
	if bundle.DiagnosticReportID != diagnosticReportID {
		return diagnostics.Bundle{}, ErrInvalidStoredData
	}
	return bundle, nil
}

func contextError(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func validUUID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return false
	}
	for index, character := range value {
		if index == 8 || index == 13 || index == 18 || index == 23 {
			continue
		}
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f') || (character >= 'A' && character <= 'F')) {
			return false
		}
	}
	return true
}


