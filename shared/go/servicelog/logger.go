package servicelog

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"
)

// ErrFileOutputUnsupported is retained for compatibility with earlier stages.
// Open no longer returns it now that file output is active.
var ErrFileOutputUnsupported = errors.New("servicelog: file output is not supported yet")

// Runtime owns the configured servicelog logger and its lifecycle.
type Runtime struct {
	logger *slog.Logger
	file   *rollingWriter

	directMu       sync.Mutex
	consoleWriter  io.Writer
	consoleEnabled bool

	failureMu     sync.Mutex
	degraded      bool
	failureCount  int
	lastError     string
	warned        bool
	warningWriter io.Writer

	flushStop chan struct{}
	flushDone chan struct{}

	closeErr  error
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

	runtime := &Runtime{
		warningWriter:  io.Discard,
		consoleWriter:  io.Discard,
		consoleEnabled: config.ConsoleEnabled,
	}
	if dependencies.consoleWriter != nil {
		runtime.warningWriter = dependencies.consoleWriter
		runtime.consoleWriter = dependencies.consoleWriter
	}

	deps := dependencies
	deps.reportFailure = runtime.recordFailure

	handlers := make([]slog.Handler, 0, 2)
	if config.ConsoleEnabled {
		handlers = append(handlers, slog.NewTextHandler(dependencies.consoleWriter, nil))
	}

	var activeWriter *rollingWriter
	if config.FileEnabled {
		rollingFile, err := newRollingWriter(config.File, deps)
		if err != nil {
			runtime.recordFailure(err)
		} else {
			activeWriter = rollingFile
			handlers = append(handlers, slog.NewJSONHandler(rollingFile, nil))
		}
	}

	if len(handlers) == 0 {
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

	runtime.logger = logger
	runtime.file = activeWriter
	if activeWriter != nil && config.Flush.Interval > 0 {
		runtime.startFlushLoop(deps, config.Flush.Interval)
	}
	return runtime, nil
}

func (r *Runtime) startFlushLoop(dependencies runtimeDependencies, interval time.Duration) {
	if r.file == nil || interval <= 0 || dependencies.newTicker == nil {
		return
	}

	ticker := dependencies.newTicker(interval)
	if ticker == nil {
		return
	}

	writer := r.file
	stop := make(chan struct{})
	done := make(chan struct{})
	r.flushStop = stop
	r.flushDone = done
	go func() {
		defer close(done)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C():
				_ = writer.Flush()
			case <-stop:
				return
			}
		}
	}()
}

func (r *Runtime) stopFlushLoop() {
	stop := r.flushStop
	done := r.flushDone
	if stop == nil || done == nil {
		return
	}

	close(stop)
	<-done
	r.flushStop = nil
	r.flushDone = nil
}

func (r *Runtime) recordFailure(err error) {
	if err == nil {
		return
	}

	r.failureMu.Lock()
	r.degraded = true
	r.failureCount++
	r.lastError = err.Error()
	shouldWarn := !r.warned
	if shouldWarn {
		r.warned = true
	}
	warningWriter := r.warningWriter
	r.failureMu.Unlock()

	if shouldWarn && warningWriter != nil {
		_, _ = fmt.Fprintf(warningWriter, "servicelog warning: %v\n", err)
	}
}

// Logger returns the runtime's structured logger.
func (r *Runtime) Logger() *slog.Logger {
	return r.logger
}

// WriteRecord writes already-serialized canonical JSON bytes through the
// existing rolling runtime without reparsing or changing the record.
func (r *Runtime) WriteRecord(jsonLine []byte, consoleLine string) error {
	if r == nil {
		return errors.New("servicelog: nil runtime")
	}
	r.directMu.Lock()
	defer r.directMu.Unlock()

	r.mu.RLock()
	closed := r.closed
	file := r.file
	r.mu.RUnlock()
	if closed {
		return errRollingWriterClosed
	}

	if r.consoleEnabled && consoleLine != "" {
		if _, err := fmt.Fprintln(r.consoleWriter, consoleLine); err != nil {
			r.recordFailure(err)
			return err
		}
	}
	if file == nil {
		return nil
	}

	record := append([]byte(nil), jsonLine...)
	if len(record) == 0 || record[len(record)-1] != '\n' {
		record = append(record, '\n')
	}
	_, err := file.Write(record)
	return err
}

// Status returns the runtime's current lifecycle state.
func (r *Runtime) Status() Status {
	r.mu.RLock()
	closed := r.closed
	r.mu.RUnlock()

	r.failureMu.Lock()
	degraded := r.degraded
	failureCount := r.failureCount
	lastError := r.lastError
	r.failureMu.Unlock()

	return Status{
		Closed:       closed,
		Degraded:     degraded,
		FailureCount: failureCount,
		LastError:    lastError,
	}
}

// Close marks the runtime closed. Repeated calls are safe and return nil.
func (r *Runtime) Close() error {
	didClose := false
	r.closeOnce.Do(func() {
		didClose = true
		r.mu.Lock()
		r.closed = true
		file := r.file
		r.file = nil
		r.mu.Unlock()

		r.stopFlushLoop()
		if file != nil {
			if err := file.Flush(); err != nil {
				r.closeErr = err
			}
			if err := file.Close(); err != nil {
				if r.closeErr == nil {
					r.closeErr = err
				}
			}
		}
	})
	if didClose {
		return r.closeErr
	}
	return nil
}
