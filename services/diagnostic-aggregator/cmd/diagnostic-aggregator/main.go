package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/hosted"
)

const (
	defaultPort     = "8080"
	shutdownTimeout = 5 * time.Second
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config, err := hosted.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "diagnostic-aggregator configuration failed: %v\n", err)
		return 1
	}
	service, err := hosted.New(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "diagnostic-aggregator initialization failed: %v\n", err)
		return 1
	}

	mux := http.NewServeMux()
	if err := service.Register(mux); err != nil {
		_ = service.Close()
		fmt.Fprintf(os.Stderr, "diagnostic-aggregator route registration failed: %v\n", err)
		return 1
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.ListenAndServe()
	}()

	exitCode := 0
	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "diagnostic-aggregator server failed: %v\n", err)
			exitCode = 1
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		if err := server.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "diagnostic-aggregator shutdown failed: %v\n", err)
			_ = server.Close()
			exitCode = 1
		}
		cancel()
		if err := <-serveErr; !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "diagnostic-aggregator server failed: %v\n", err)
			exitCode = 1
		}
	}

	if err := service.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "diagnostic-aggregator close failed: %v\n", err)
		exitCode = 1
	}
	return exitCode
}
