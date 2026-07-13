package runtime

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/health"
	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/ingest"
	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/query"
	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/storage"
	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/storage/storagefake"
)

type fakeIngestor struct{}

func (fakeIngestor) IngestBatch(context.Context, []json.RawMessage) (ingest.BatchResult, error) {
	return ingest.BatchResult{BatchID: "batch-1", Accepted: 1}, nil
}

type fakeQuerier struct{}

func (fakeQuerier) Query(context.Context, query.Filter) (query.Result, error) {
	return query.Result{}, nil
}

func dependencies() Dependencies {
	return Dependencies{Store: storagefake.New(), Ingestor: fakeIngestor{}, Querier: fakeQuerier{}, Health: health.NewState("i", "v", "test", time.Time{})}
}

func TestDependenciesValidateRequiredSeams(t *testing.T) {
	base := dependencies()
	cases := []struct {
		name   string
		mutate func(*Dependencies)
		want   string
	}{
		{"store", func(d *Dependencies) { d.Store = nil }, "event store"},
		{"ingestor", func(d *Dependencies) { d.Ingestor = nil }, "ingestor"},
		{"querier", func(d *Dependencies) { d.Querier = nil }, "querier"},
		{"health", func(d *Dependencies) { d.Health = nil }, "health state"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := base
			tc.mutate(&d)
			if err := d.Validate(); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestNewHandlerComposesRoutes(t *testing.T) {
	handler, err := NewHandler(dependencies())
	if err != nil {
		t.Fatal(err)
	}
	for _, check := range []struct {
		path string
		want int
	}{{"/live", http.StatusOK}, {"/ready", http.StatusServiceUnavailable}, {"/status", http.StatusOK}} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, check.path, nil))
		if recorder.Code != check.want {
			t.Fatalf("%s status = %d, want %d", check.path, recorder.Code, check.want)
		}
	}
	checks := []struct {
		method, path string
		want         int
	}{
		{http.MethodGet, "/v1/events", http.StatusOK},
		{http.MethodPost, "/v1/events", http.StatusOK},
		{http.MethodGet, "/v1/traces/123e4567-e89b-42d3-a456-426614174000", http.StatusOK},
	}
	for _, check := range checks {
		recorder := httptest.NewRecorder()
		requestBody := strings.NewReader(`{"events":[{}]}`)
		handler.ServeHTTP(recorder, httptest.NewRequest(check.method, check.path, requestBody))
		if recorder.Code != check.want {
			t.Fatalf("%s %s status = %d, want %d", check.method, check.path, recorder.Code, check.want)
		}
	}
}

func TestEventsRouteMethodGuard(t *testing.T) {
	handler, err := NewHandler(dependencies())
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPut, "/v1/events", nil))
	if recorder.Code != http.StatusMethodNotAllowed || recorder.Header().Get("Allow") != "GET, POST" {
		t.Fatalf("unexpected response: %d allow=%q", recorder.Code, recorder.Header().Get("Allow"))
	}
}

var _ storage.EventStore = (*storagefake.Store)(nil)
