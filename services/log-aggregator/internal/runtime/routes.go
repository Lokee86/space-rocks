package runtime

import (
	"fmt"
	"net/http"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/health"
	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/ingest"
	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/query"
	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/storage"
)

type Dependencies struct {
	Store    storage.EventStore
	Ingestor ingest.BatchIngestor
	Querier  query.EventQuerier
	Health   *health.State
}

func (d Dependencies) Validate() error {
	if d.Store == nil {
		return fmt.Errorf("runtime: storage event store is required")
	}
	if d.Ingestor == nil {
		return fmt.Errorf("runtime: ingest batch ingestor is required")
	}
	if d.Querier == nil {
		return fmt.Errorf("runtime: query event querier is required")
	}
	if d.Health == nil {
		return fmt.Errorf("runtime: health state is required")
	}
	return nil
}

func NewHandler(d Dependencies) (http.Handler, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	healthHandler := health.NewHandler(d.Health, d.Store)
	mux.Handle("/live", healthHandler)
	mux.Handle("/ready", healthHandler)
	mux.Handle("/status", healthHandler)
	querier := NewInstrumentedQuerier(d.Querier, d.Health)
	ingestor := NewInstrumentedIngestor(d.Ingestor, d.Health)
	mux.Handle("/v1/traces/", query.NewTraceHandler(querier))
	getEvents := query.NewEventsHandler(querier)
	postEvents := ingest.NewHandler(ingestor)
	mux.Handle("/v1/events", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getEvents.ServeHTTP(w, r)
		case http.MethodPost:
			postEvents.ServeHTTP(w, r)
		default:
			w.Header().Set("Allow", "GET, POST")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte(`{"code":"method_not_allowed"}` + "\n"))
		}
	}))
	return mux, nil
}
