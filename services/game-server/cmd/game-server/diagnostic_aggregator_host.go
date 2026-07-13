package main

import (
	"net/http"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/hosted"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
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
	if config.Enabled {
		logging.Server.Info("diagnostic aggregator hosted service registered")
	} else {
		logging.Server.Info("diagnostic aggregator hosted service disabled")
	}
	return service, nil
}

func closeDiagnosticAggregator(service *hosted.Service) {
	if service != nil {
		if err := service.Close(); err != nil {
			logging.Server.Error("diagnostic aggregator close failed", err)
		}
	}
}
