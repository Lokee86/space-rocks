package servicelog

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

// ErrFileOutputUnsupported is retained for compatibility with earlier stages.
// Open no longer returns it now that file output is active.
var ErrFileOutputUnsupported = errors.New("servicelog: file output is not supported yet")

// Runtime owns the configured servicelog logger and its lifecycle.
type Runtime struct {
	logger   *slog.Logger
	file     io.Closer
	closeErr error

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

	handlers := make([]slog.Handler, 0, 2)
	if config.ConsoleEnabled {
		handlers = append(handlers, slog.NewTextHandler(dependencies.consoleWriter, nil))
	}

	var activeFile io.Closer
	if config.FileEnabled {
		if err := dependencies.mkdir(config.File.Directory, 0o755); err != nil {
			return nil, err
		}

		filePath := filepath.Join(config.File.Directory, config.File.Prefix+".jsonl.open")
		fileHandle, err := dependencies.openFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, err
		}
		activeFile = fileHandle
		handlers = append(handlers, slog.NewJSONHandler(fileHandle, nil))
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

	return &Runtime{logger: logger, file: activeFile}, nil
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
	didClose := false
	r.closeOnce.Do(func() {
		didClose = true
		r.mu.Lock()
		r.closed = true
		file := r.file
		r.file = nil
		r.mu.Unlock()

		if file != nil {
			r.closeErr = file.Close()
		}
	})
	if didClose {
		return r.closeErr
	}
	return nil
}
