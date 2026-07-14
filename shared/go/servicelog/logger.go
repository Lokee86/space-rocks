package servicelog

import (
	"io"
	"log/slog"
	"sync"
	"time"
)

// Runtime owns the configured servicelog logger and its lifecycle.
type Runtime struct {
	logger           *slog.Logger
	fileWriter       *rollingWriter
	maintenanceStop  chan struct{}
	maintenanceDone  chan struct{}
	closeOnce        sync.Once
	mu               sync.RWMutex
	closed           bool
	closeErr         error
}

// Open constructs a servicelog runtime from caller-supplied policy.
func Open(config Config) (*Runtime, error) {
	return openWithDependencies(config, defaultRuntimeDependencies())
}

func openWithDependencies(config Config, dependencies runtimeDependencies) (*Runtime, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	var handlers []slog.Handler
	if config.ConsoleEnabled {
		handlers = append(handlers, slog.NewTextHandler(dependencies.consoleWriter, nil))
	} else {
		handlers = append(handlers, slog.NewTextHandler(io.Discard, nil))
	}

	var fileWriter *rollingWriter
	if config.FileEnabled {
		var err error
		fileWriter, err = newRollingWriter(config.File, dependencies)
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, slog.NewJSONHandler(fileWriter, nil))
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

	runtime := &Runtime{logger: logger, fileWriter: fileWriter}
	if fileWriter != nil && config.Flush.Interval > 0 && dependencies.newTicker != nil {
		runtime.maintenanceStop = make(chan struct{})
		runtime.maintenanceDone = make(chan struct{})
		go runtime.runMaintenanceLoop(dependencies, config.Flush.Interval)
	}
	return runtime, nil
}

// Logger returns the runtime's structured logger.
func (r *Runtime) Logger() *slog.Logger {
	return r.logger
}

// Status returns the runtime's current lifecycle state.
func (r *Runtime) Status() Status {
	r.mu.RLock()
	closed := r.closed
	fileWriter := r.fileWriter
	r.mu.RUnlock()

	if fileWriter == nil {
		return Status{Closed: closed}
	}
	status := fileWriter.statusSnapshot()
	status.Closed = closed
	return status
}

// Close marks the runtime closed. Repeated calls are safe and return nil.
func (r *Runtime) Close() error {
	r.closeOnce.Do(func() {
		r.mu.Lock()
		r.closed = true
		stop := r.maintenanceStop
		done := r.maintenanceDone
		r.mu.Unlock()

		if stop != nil {
			close(stop)
			<-done
		}
		if r.fileWriter != nil {
			r.closeErr = r.fileWriter.Close()
		}
	})
	return r.closeErr
}

func (r *Runtime) runMaintenanceLoop(dependencies runtimeDependencies, interval time.Duration) {
	ticker := dependencies.newTicker(interval)
	defer ticker.Stop()
	defer close(r.maintenanceDone)

	for {
		select {
		case <-r.maintenanceStop:
			return
		case tick, ok := <-ticker.C():
			if !ok {
				return
			}
			if r.fileWriter != nil {
				_ = r.fileWriter.Maintain(tick.UTC())
			}
		}
	}
}
