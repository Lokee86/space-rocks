package runtime

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sync"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/config"
)

type ListenFunc func(network, address string) (net.Listener, error)

type App struct {
	config    config.Config
	deps      Dependencies
	server    *http.Server
	listen    ListenFunc
	logger    *slog.Logger
	runMu     sync.Mutex
	run       bool
	closeOnce sync.Once
}

func NewApp(cfg config.Config, deps Dependencies, listen ListenFunc, logger *slog.Logger) (*App, error) {
	if err := deps.Validate(); err != nil {
		return nil, err
	}
	if listen == nil {
		listen = net.Listen
	}
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	handler, err := NewHandler(deps)
	if err != nil {
		return nil, err
	}
	return &App{config: cfg, deps: deps, server: &http.Server{Addr: cfg.ListenAddress, Handler: handler, ReadHeaderTimeout: cfg.ReadHeaderTimeout, ReadTimeout: cfg.ReadTimeout, WriteTimeout: cfg.WriteTimeout, IdleTimeout: cfg.IdleTimeout}, listen: listen, logger: logger}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.runMu.Lock()
	if a.run {
		a.runMu.Unlock()
		return fmt.Errorf("runtime: app can only be run once")
	}
	a.run = true
	a.runMu.Unlock()
	a.log(slog.LevelInfo, "service_starting", "")
	listener, err := a.listen("tcp", a.config.ListenAddress)
	if err != nil {
		a.log(slog.LevelError, "service_startup_failed", "bind_failed")
		return a.closeStore(fmt.Errorf("bind: %w", err))
	}
	a.deps.Health.MarkReady()
	a.log(slog.LevelInfo, "service_ready", "")
	serve := make(chan error, 1)
	go func() { serve <- a.server.Serve(listener) }()
	select {
	case err := <-serve:
		a.deps.Health.MarkStopping()
		if errors.Is(err, http.ErrServerClosed) {
			err = a.closeStore(nil)
			a.log(slog.LevelInfo, "service_stopped", "")
			return err
		}
		err = a.closeStore(fmt.Errorf("serve: %w", err))
		a.log(slog.LevelError, "service_stopped", "serve_failed")
		return err
	case <-ctx.Done():
		a.deps.Health.MarkStopping()
		a.log(slog.LevelInfo, "service_stopping", "")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.config.ShutdownTimeout)
		shutdownErr := a.server.Shutdown(shutdownCtx)
		cancel()
		var forceCloseErr error
		if shutdownErr != nil {
			forceCloseErr = a.server.Close()
		}
		serveErr := <-serve
		if errors.Is(serveErr, http.ErrServerClosed) {
			serveErr = nil
		}
		err := errors.Join(wrap("shutdown", shutdownErr), wrap("force_close", forceCloseErr), wrap("serve", serveErr))
		err = a.closeStore(err)
		a.log(slog.LevelInfo, "service_stopped", "")
		return err
	}
}

func (a *App) closeStore(previous error) error {
	var closeErr error
	a.closeOnce.Do(func() { closeErr = a.deps.Store.Close() })
	return errors.Join(previous, wrap("store_close", closeErr))
}

func wrap(label string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", label, err)
}
func (a *App) log(level slog.Level, event, code string) {
	args := []any{"event", event}
	if code != "" {
		args = append(args, "error_code", code)
	}
	a.logger.Log(context.Background(), level, "log aggregator lifecycle", args...)
}
