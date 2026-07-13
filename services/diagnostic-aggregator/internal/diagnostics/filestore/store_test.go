package filestore

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/diagnostics"
)

const reportID = "123e4567-e89b-12d3-a456-426614174001"

func testBundle() diagnostics.Bundle {
	return diagnostics.Bundle{
		Version: diagnostics.BundleVersion, DiagnosticReportID: reportID,
		TraceID: "123e4567-e89b-12d3-a456-426614174000",
		Events: []diagnostics.Event{{Payload: json.RawMessage(`{"x":1}`)}},
		Services: []string{"game-server"},
	}
}

func TestStoreSaveAndGet(t *testing.T) {
	store, err := New(t.TempDir(), 4096)
	if err != nil { t.Fatal(err) }
	bundle := testBundle()
	if err := store.Save(context.Background(), bundle); err != nil { t.Fatal(err) }
	loaded, err := store.Get(context.Background(), reportID)
	if err != nil { t.Fatal(err) }
	if string(loaded.Events[0].Payload) != string(bundle.Events[0].Payload) || loaded.Services[0] != "game-server" { t.Fatalf("unexpected bundle: %#v", loaded) }
}

func TestStoreDoesNotOverwriteAndCleansTemporaryFiles(t *testing.T) {
	store, err := New(t.TempDir(), 4096)
	if err != nil { t.Fatal(err) }
	bundle := testBundle()
	if err := store.Save(context.Background(), bundle); err != nil { t.Fatal(err) }
	if err := store.Save(context.Background(), bundle); !errors.Is(err, ErrDuplicateBundle) { t.Fatalf("expected duplicate error, got %v", err) }
	if matches, _ := filepath.Glob(filepath.Join(store.root, ".diagnostic-*.tmp")); len(matches) != 0 { t.Fatalf("temporary files remain: %v", matches) }
}

func TestStoreRejectsInvalidIDsAndMapsMissing(t *testing.T) {
	store, err := New(t.TempDir(), 4096)
	if err != nil { t.Fatal(err) }
	if err := store.Save(context.Background(), diagnostics.Bundle{DiagnosticReportID: "bad"}); !errors.Is(err, diagnostics.ErrInvalidDiagnosticReportID) { t.Fatal(err) }
	if _, err := store.Get(context.Background(), "bad"); !errors.Is(err, diagnostics.ErrInvalidDiagnosticReportID) { t.Fatal(err) }
	if _, err := store.Get(context.Background(), reportID); !errors.Is(err, diagnostics.ErrBundleNotFound) { t.Fatal(err) }
}

func TestStoreRejectsMalformedMismatchAndOversizedFiles(t *testing.T) {
	root := t.TempDir()
	const maxBytes int64 = 32
	store, err := New(root, maxBytes)
	if err != nil { t.Fatal(err) }
	if err := store.Save(context.Background(), testBundle()); !errors.Is(err, ErrBundleTooLarge) { t.Fatal(err) }
	store, err = New(root, 4096)
	if err != nil { t.Fatal(err) }
	path := filepath.Join(root, reportID+".json")
	if err := os.WriteFile(path, []byte("{bad"), 0o644); err != nil { t.Fatal(err) }
	if _, err := store.Get(context.Background(), reportID); !errors.Is(err, ErrInvalidStoredData) { t.Fatal(err) }
	if err := os.WriteFile(path, []byte(`{"diagnostic_report_id":"123e4567-e89b-12d3-a456-426614174002"}`), 0o644); err != nil { t.Fatal(err) }
	if _, err := store.Get(context.Background(), reportID); !errors.Is(err, ErrInvalidStoredData) { t.Fatal(err) }
	store, err = New(root, maxBytes)
	if err != nil { t.Fatal(err) }
	if err := os.WriteFile(path, make([]byte, int(maxBytes)+1), 0o644); err != nil { t.Fatal(err) }
	if _, err := store.Get(context.Background(), reportID); !errors.Is(err, ErrBundleTooLarge) { t.Fatal(err) }
}

func TestStoreHonorsCanceledContexts(t *testing.T) {
	store, err := New(t.TempDir(), 4096)
	if err != nil { t.Fatal(err) }
	ctx, cancel := context.WithCancel(context.Background()); cancel()
	if err := store.Save(ctx, testBundle()); !errors.Is(err, context.Canceled) { t.Fatal(err) }
	if _, err := store.Get(ctx, reportID); !errors.Is(err, context.Canceled) { t.Fatal(err) }
}

func TestStoreRequiresRoot(t *testing.T) {
	if _, err := New("", 0); !errors.Is(err, ErrInvalidRoot) { t.Fatal(err) }
}

func TestStoreConcurrentSavesPublishOneBundle(t *testing.T) {
	store, err := New(t.TempDir(), 4096)
	if err != nil { t.Fatal(err) }
	bundle := testBundle()
	errorsFound := make(chan error, 2)
	var wait sync.WaitGroup
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			errorsFound <- store.Save(context.Background(), bundle)
		}()
	}
	wait.Wait()
	close(errorsFound)
	var successes, duplicates int
	for saveErr := range errorsFound {
		if saveErr == nil {
			successes++
		} else if errors.Is(saveErr, ErrDuplicateBundle) {
			duplicates++
		} else {
			t.Fatalf("unexpected concurrent save error: %v", saveErr)
		}
	}
	if successes != 1 || duplicates != 1 {
		t.Fatalf("successes=%d duplicates=%d", successes, duplicates)
	}
	if _, err := store.Get(context.Background(), reportID); err != nil {
		t.Fatal(err)
	}
	if matches, _ := filepath.Glob(filepath.Join(store.root, ".diagnostic-*.tmp")); len(matches) != 0 {
		t.Fatalf("temporary files remain: %v", matches)
	}
}
