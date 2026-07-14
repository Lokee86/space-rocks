package servicelog

import (
	"errors"
	"io"
	"log/slog"
	"sync"
)

// ErrFileOutputUnsupported indicates that file output was requested before
// the rolling file sink is available.
var ErrFileOutputUnsupported = errors.New("servicelog: file output is not supported yet")

// Runtime owns the configured servicelog logger and its lifecycle.
type Runtime struct {
	logger    *slog.Logger
	closeOnce sync.Once
	mu        sync.RWMutex
	closed    bool
}

// Open constructs a servicelog runtime from caller-supplied policy.
func Open(config Config) (*Runtime, error) {
	return openWithDependencies(config, defaultRuntimeDependencies())
}

func openWithDependencies(config Config, dependencies runtimeDependencies) (*Runtime, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if config.FileEnabled {
		return nil, ErrFileOutputUnsupported
	}

	var handlers []slog.Handler
	if config.ConsoleEnabled {
		handlers = append(handlers, slog.NewTextHandler(dependencies.consoleWriter, nil))
	} else {
		handlers = append(handlers, slog.NewTextHandler(io.Discard, nil))
	}

	logger := slog.New(newFanoutHandler(handlers...)).With(
		slog.String("service", config.Identity.Name),
	)
	if config.Identity.InstanceID != "" {
		logger = logger.With(slog.String("service_instance_id", config.Identity.InstanceID))
	}
	if config.Identity.Environment != "" {
		logger = logger.With(slog.String("environment", config.Identity.Environment))
	}
	if config.Identity.Version != "" {
		logger = logger.With(slog.String("build_version", config.Identity.Version))
	}

	return &Runtime{logger: logger}, nil
}

// Logger returns the runtime's structured logger.
func (r *Runtime) Logger() *slog.Logger {
	return r.logger
}

// Status returns the runtime's current lifecycle state.
func (r *Runtime) Status() Status {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return Status{Closed: r.closed}
}

// Close marks the runtime closed. Repeated calls are safe and return nil.
func (r *Runtime) Close() error {
	r.closeOnce.Do(func() {
		r.mu.Lock()
		r.closed = true
		r.mu.Unlock()
	})
	return nil
}
