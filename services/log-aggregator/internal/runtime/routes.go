package runtime

import (
	"fmt"
	"net/http"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/health"
	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/storage"
)

type Dependencies struct {
	Store  storage.EventStore
	Health *health.State
}

func (d Dependencies) Validate() error {
	if d.Store == nil {
		return fmt.Errorf("runtime: storage event store is required")
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
	return mux, nil
}
