package runtime

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/health"
	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/storage"
	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/storage/storagefake"
)

func dependencies() Dependencies {
	return Dependencies{Store: storagefake.New(), Health: health.NewState("i", "v", "test", time.Time{})}
}

func TestDependenciesValidateRequiredSeams(t *testing.T) {
	base := dependencies()
	cases := []struct {
		name   string
		mutate func(*Dependencies)
		want   string
	}{
		{"store", func(d *Dependencies) { d.Store = nil }, "event store"},
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
	}{
		{http.MethodGet, "/v1/events"},
		{http.MethodPost, "/v1/events"},
		{http.MethodGet, "/v1/traces/"},
		{http.MethodGet, "/v1/traces/123e4567-e89b-42d3-a456-426614174000"},
	}
	for _, check := range checks {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(check.method, check.path, nil))
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("%s %s status = %d, want %d", check.method, check.path, recorder.Code, http.StatusNotFound)
		}
	}
}

var _ storage.EventStore = (*storagefake.Store)(nil)
