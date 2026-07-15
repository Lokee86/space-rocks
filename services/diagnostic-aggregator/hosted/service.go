package hosted

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/diagnosticapi"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/diagnosticreports"
	operational "github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/logging"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/redaction"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage/jsonlstore"
	"github.com/Lokee86/space-rocks/shared/go/servicelog"
	"github.com/google/uuid"
)

type Service struct {
	handler        http.Handler
	store          *jsonlstore.ReportStore
	operationalLog *operational.Logger
	closeOnce      sync.Once
	closeErr       error
}

func New(config Config) (*Service, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if !config.Enabled {
		return &Service{}, nil
	}

	logger, err := operational.Open(operationalLogConfig(config))
	if err != nil {
		return nil, err
	}
	logger.Info("diagnostic aggregator hosted service starting")

	store, err := jsonlstore.NewReportStore(reportStoreConfig(config))
	if err != nil {
		logger.Error("diagnostic report store initialization failed", err)
		_ = logger.Close()
		return nil, err
	}
	fail := func(err error) (*Service, error) {
		logger.Error("diagnostic aggregator hosted service initialization failed", err)
		_ = store.Close()
		_ = logger.Close()
		return nil, err
	}
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
	logger.Info("diagnostic aggregator hosted service started")
	return &Service{handler: handler, store: store, operationalLog: logger}, nil
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
		var storeErr error
		if s.store != nil {
			storeErr = s.store.Close()
		}
		if s.operationalLog == nil {
			s.closeErr = storeErr
			return
		}
		if storeErr != nil {
			s.operationalLog.Error("diagnostic report store close failed", storeErr)
		}
		s.operationalLog.Info("diagnostic aggregator hosted service stopped")
		s.closeErr = errors.Join(storeErr, s.operationalLog.Close())
	})
	return s.closeErr
}

func (s *Service) LoggingStatus() servicelog.Status {
	if s == nil || s.operationalLog == nil {
		return servicelog.Status{}
	}
	return s.operationalLog.Status()
}
