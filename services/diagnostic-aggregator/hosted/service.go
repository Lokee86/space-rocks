package hosted

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/diagnosticapi"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/diagnosticreports"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/redaction"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage/jsonlstore"
	"github.com/google/uuid"
)

type Service struct {
	handler   http.Handler
	store     *jsonlstore.ReportStore
	closeOnce sync.Once
	closeErr  error
}

func New(config Config) (*Service, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if !config.Enabled {
		return &Service{}, nil
	}
	store, err := jsonlstore.NewReportStore(reportStoreConfig(config))
	if err != nil {
		return nil, err
	}
	fail := func(err error) (*Service, error) { _ = store.Close(); return nil, err }
	if _, err := store.EnforceRetention(context.Background()); err != nil {
		return fail(err)
	}
	authorize, err := diagnosticapi.NewBearerTokenAuthorizer(config.BearerToken)
	if err != nil {
		return fail(err)
	}
	reports, err := diagnosticreports.NewService(store, config.SubmissionLimits, redaction.DefaultPolicy(), time.Now, func() (string, error) { return uuid.NewString(), nil })
	if err != nil {
		return fail(err)
	}
	handler, err := diagnosticapi.NewHandler(reports, diagnosticapi.HandlerConfig{MaxRequestBytes: config.MaxRequestBytes, Authorize: authorize})
	if err != nil {
		return fail(err)
	}
	return &Service{handler: handler, store: store}, nil
}

func (s *Service) Register(mux *http.ServeMux) error {
	if mux == nil {
		return errors.New("hosted: mux is required")
	}
	if s.handler == nil {
		return nil
	}
	mux.Handle(diagnosticapi.DiagnosticReportsPath, s.handler)
	mux.Handle(diagnosticapi.DiagnosticReportItemPathPrefix, s.handler)
	return nil
}

func (s *Service) Close() error {
	s.closeOnce.Do(func() {
		if s.store != nil {
			s.closeErr = s.store.Close()
		}
	})
	return s.closeErr
}
