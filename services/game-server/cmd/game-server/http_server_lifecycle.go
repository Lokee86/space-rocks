package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

func serveHTTPServer(ctx context.Context, server *http.Server, listener net.Listener, shutdownTimeout time.Duration, onStopping func()) error {
	serveResult := make(chan error, 1)
	go func() { serveResult <- server.Serve(listener) }()
	select {
	case err := <-serveResult:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		if onStopping != nil {
			onStopping()
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		shutdownErr := server.Shutdown(shutdownCtx)
		cancel()
		if shutdownErr != nil {
			_ = server.Close()
			<-serveResult
			return shutdownErr
		}
		err := <-serveResult
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
