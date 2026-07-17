package main

import (
	"net/http"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/hosted"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

func registerDiagnosticAggregator(mux *http.ServeMux) (*hosted.Service, error) {
	config, err := hosted.LoadConfig()
	if err != nil {
		return nil, err
	}
	service, err := hosted.New(config)
	if err != nil {
		return nil, err
	}
	if err := service.Register(mux); err != nil {
		_ = service.Close()
		return nil, err
	}
	return service, nil
}

func closeDiagnosticAggregator(service *hosted.Service, lifecycleTraceID string) {
	if service != nil {
		if err := service.Close(); err != nil {
			logging.Emit(observability.Request{
				Event:   observability.EventNameObservabilityUnavailable,
				Context: observability.Context{TraceID: lifecycleTraceID},
				Fields:  observability.Fields{"failure_mode": "diagnostic_aggregator_close_failed"},
			})
		}
	}
}
