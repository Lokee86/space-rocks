package diagnostics

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage"
)

type fakeBundleStore struct {
	saved      Bundle
	stored     Bundle
	saveErr    error
	getErr     error
	saveCalls  int
	getCalls   int
	requested  string
}

func (f *fakeBundleStore) Save(_ context.Context, bundle Bundle) error {
	f.saveCalls++
	f.saved = bundle
	return f.saveErr
}

func (f *fakeBundleStore) Get(_ context.Context, id string) (Bundle, error) {
	f.getCalls++
	f.requested = id
	return f.stored, f.getErr
}

func serviceBuilder() Builder {
	return Builder{
		Store: &fakeQuerier{result: storage.QueryResult{Records: []storage.Record{{
			EventID: "event", Timestamp: time.Unix(1, 0), Payload: json.RawMessage(`{"x":1}`),
		}}, Total: 1}},
		NewUUID: func() (string, error) { return testReportID, nil },
		Now:     func() time.Time { return time.Unix(2, 0).UTC() },
		Sanitize: func(payload json.RawMessage) (json.RawMessage, SanitizationSummary, error) {
			return append(json.RawMessage(nil), payload...), SanitizationSummary{}, nil
		},
	}
}

func storedBundle() Bundle {
	return Bundle{
		DiagnosticReportID: testReportID,
		Events: []Event{{Payload: json.RawMessage(`{"stored":true}`)}},
		Services: []string{"service"},
		RequestIDs: []string{"request"},
	}
}

func TestServiceCreateBuildsAndSaves(t *testing.T) {
	store := &fakeBundleStore{}
	service := Service{Builder: serviceBuilder(), Store: store}
	bundle, err := service.Create(context.Background(), testTraceID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if store.saveCalls != 1 || store.saved.DiagnosticReportID != testReportID || bundle.DiagnosticReportID != testReportID {
		t.Fatalf("unexpected create result: %#v %#v", bundle, store.saved)
	}
}

func TestServiceCreateBuildAndSaveFailures(t *testing.T) {
	store := &fakeBundleStore{}
	service := Service{Builder: Builder{}, Store: store}
	if _, err := service.Create(context.Background(), testTraceID, 1); err == nil {
		t.Fatal("expected build failure")
	}
	store.saveErr = errors.New("save failed")
	service.Builder = serviceBuilder()
	if _, err := service.Create(context.Background(), testTraceID, 1); err == nil {
		t.Fatal("expected save failure")
	}
}

func TestServiceGetValidatesAndCopies(t *testing.T) {
	store := &fakeBundleStore{stored: storedBundle()}
	service := Service{Store: store}
	bundle, err := service.Get(context.Background(), testReportID)
	if err != nil {
		t.Fatal(err)
	}
	bundle.Events[0].Payload[0] = 'x'
	bundle.Services[0] = "changed"
	if string(store.stored.Events[0].Payload) != `{"stored":true}` || store.stored.Services[0] != "service" {
		t.Fatal("get result aliases store data")
	}
	if store.requested != testReportID || store.getCalls != 1 {
		t.Fatalf("unexpected get call: %#v", store)
	}
}

func TestServiceGetInvalidIDBeforeStoreAndErrors(t *testing.T) {
	store := &fakeBundleStore{getErr: ErrBundleNotFound}
	service := Service{Store: store}
	if _, err := service.Get(context.Background(), "invalid"); !errors.Is(err, ErrInvalidDiagnosticReportID) {
		t.Fatalf("expected invalid id, got %v", err)
	}
	if store.getCalls != 0 {
		t.Fatal("store called for invalid id")
	}
	if _, err := service.Get(context.Background(), testReportID); !errors.Is(err, ErrBundleNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	store.getErr = errors.New("store failed")
	if _, err := service.Get(context.Background(), testReportID); err == nil {
		t.Fatal("expected store failure")
	}
}
