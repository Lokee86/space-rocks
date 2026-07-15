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
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage/jsonlstore"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"github.com/Lokee86/space-rocks/shared/go/servicelog"
	"github.com/google/uuid"
)

type reportStore interface {
	storage.ReportStore
	EnforceRetention(context.Context) (int, error)
}

type serviceDependencies struct {
	newStore         func(jsonlstore.Config) (reportStore, error)
	newAuthorizer    func(string) (diagnosticapi.RequestAuthorizer, error)
	newReportService func(reportStore, Config, diagnosticreports.UUIDGenerator, diagnosticreports.EventEmitter) (diagnosticapi.ReportService, error)
	newHandler       func(diagnosticapi.ReportService, diagnosticapi.HandlerConfig) (http.Handler, error)
	newUUID          func() string
}

var defaultServiceDependencies = serviceDependencies{
	newStore: func(config jsonlstore.Config) (reportStore, error) {
		return jsonlstore.NewReportStore(config)
	},
	newAuthorizer: diagnosticapi.NewBearerTokenAuthorizer,
	newReportService: func(store reportStore, config Config, newUUID diagnosticreports.UUIDGenerator, emitter diagnosticreports.EventEmitter) (diagnosticapi.ReportService, error) {
		return diagnosticreports.NewService(store, config.SubmissionLimits, redaction.DefaultPolicy(), time.Now, newUUID, emitter)
	},
	newHandler: diagnosticapi.NewHandler,
	newUUID:    uuid.NewString,
}

type Service struct {
	handler        http.Handler
	store          reportStore
	operationalLog *operational.Logger
	closeOnce      sync.Once
	closeErr       error
	newUUID        func() string
}

func New(config Config) (*Service, error) {
	return newWithDependencies(config, defaultServiceDependencies)
}

func newWithDependencies(config Config, deps serviceDependencies) (*Service, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if !config.Enabled {
		return &Service{}, nil
	}
	if deps.newStore == nil || deps.newAuthorizer == nil || deps.newReportService == nil || deps.newHandler == nil || deps.newUUID == nil {
		return nil, errors.New("hosted: service dependencies are required")
	}

	logger, err := operational.Open(operationalLogConfig(config))
	if err != nil {
		return nil, err
	}
	startupTrace := deps.newUUID()
	logger.Emit(observability.Request{Event: observability.EventNameServiceStarting, Context: observability.Context{TraceID: startupTrace}})

	store, err := deps.newStore(reportStoreConfig(config))
	if err != nil {
		logger.Emit(observability.Request{
			Event:   observability.EventNameDependencyInitializationFailed,
			Context: observability.Context{TraceID: startupTrace},
			Fields:  observability.Fields{"dependency": "diagnostic_report_store", "failure_stage": "store_initialization", "failure_mode": lifecycleFailureMode(err)},
		})
		_ = logger.Close()
		return nil, err
	}
	fail := func(stage string, err error) (*Service, error) {
		logger.Emit(observability.Request{
			Event:   observability.EventNameServiceStartupFailed,
			Context: observability.Context{TraceID: startupTrace},
			Fields:  observability.Fields{"failure_stage": stage, "failure_mode": lifecycleFailureMode(err)},
		})
		_ = store.Close()
		_ = logger.Close()
		return nil, err
	}
	if _, err := store.EnforceRetention(context.Background()); err != nil {
		return fail("retention_enforcement", err)
	}
	authorize, err := deps.newAuthorizer(config.BearerToken)
	if err != nil {
		return fail("authorizer_initialization", err)
	}
	reports, err := deps.newReportService(store, config, func() (string, error) { return deps.newUUID(), nil }, logger)
	if err != nil {
		return fail("report_service_initialization", err)
	}
	handler, err := deps.newHandler(reports, diagnosticapi.HandlerConfig{MaxRequestBytes: config.MaxRequestBytes, Authorize: authorize, Emitter: logger, NewUUID: deps.newUUID})
	if err != nil {
		return fail("handler_initialization", err)
	}
	logger.Emit(observability.Request{Event: observability.EventNameServiceStarted, Context: observability.Context{TraceID: startupTrace}})
	return &Service{handler: handler, store: store, operationalLog: logger, newUUID: deps.newUUID}, nil
}

func lifecycleFailureMode(err error) string {
	switch {
	case errors.Is(err, context.Canceled):
		return "canceled"
	case errors.Is(err, context.DeadlineExceeded):
		return "deadline_exceeded"
	default:
		return "unavailable"
	}
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
		if s.operationalLog == nil {
			if s.store != nil {
				s.closeErr = s.store.Close()
			}
			return
		}
		shutdownTrace := s.newUUID()
		s.operationalLog.Emit(observability.Request{Event: observability.EventNameServiceStopping, Context: observability.Context{TraceID: shutdownTrace}})
		var storeErr error
		if s.store != nil {
			storeErr = s.store.Close()
		}
		if storeErr != nil {
			s.operationalLog.Emit(observability.Request{
				Event: observability.EventNameAggregatorStorageFailed, Context: observability.Context{TraceID: shutdownTrace},
				Fields: observability.Fields{"operation": "close", "failure_mode": lifecycleFailureMode(storeErr)},
			})
		}
		s.operationalLog.Emit(observability.Request{Event: observability.EventNameServiceStopped, Context: observability.Context{TraceID: shutdownTrace}})
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
