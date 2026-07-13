package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
)

func restoreAggregatorLogging(t *testing.T) {
	t.Helper()
	if err := logging.CloseAggregatorOutput(context.Background()); err != nil {
		t.Fatal(err)
	}
	logging.Configure("warn")
}

func TestConfigureAggregatorLoggingDisabled(t *testing.T) {
	defer restoreAggregatorLogging(t)
	t.Setenv("OBS_AGGREGATOR_ENABLED", "false")
	enabled, err := configureAggregatorLoggingFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if enabled {
		t.Fatal("disabled aggregator reported enabled")
	}
}

func TestConfigureAggregatorLoggingInvalidEnabled(t *testing.T) {
	defer restoreAggregatorLogging(t)
	t.Setenv("OBS_AGGREGATOR_ENABLED", "true")
	t.Setenv("OBS_AGGREGATOR_ENDPOINT_URL", "not-a-url")
	t.Setenv("OBS_AGGREGATOR_BEARER_TOKEN", "do-not-leak")
	if _, err := configureAggregatorLoggingFromEnv(); err == nil {
		t.Fatal("expected configuration error")
	} else if strings.Contains(err.Error(), "do-not-leak") {
		t.Fatal("bearer token leaked in error")
	}
}

func TestConfigureAggregatorLoggingValidDeliversOnClose(t *testing.T) {
	received := make(chan []byte, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		received <- append([]byte(nil), payload...)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	defer restoreAggregatorLogging(t)
	t.Setenv("OBS_AGGREGATOR_ENABLED", "true")
	t.Setenv("OBS_AGGREGATOR_ENDPOINT_URL", server.URL)
	t.Setenv("OBS_AGGREGATOR_BATCH_SIZE", "10")
	t.Setenv("OBS_AGGREGATOR_FLUSH_INTERVAL", "1h")
	t.Setenv("OBS_AGGREGATOR_SPOOL_ENABLED", "false")
	enabled, err := configureAggregatorLoggingFromEnv()
	if err != nil || !enabled {
		t.Fatalf("enabled = %v, err = %v", enabled, err)
	}
	logging.Configure("info")
	logging.Info("startup aggregator test", "component", "game-server")
	if err := closeAggregatorLogging(time.Second); err != nil {
		t.Fatal(err)
	}
	select {
	case payload := <-received:
		if !json.Valid(payload) {
			t.Fatal("received invalid JSON")
		}
	case <-time.After(time.Second):
		t.Fatal("queued log was not delivered")
	}
}

func TestCloseAggregatorLoggingTimeoutIsBounded(t *testing.T) {
	defer restoreAggregatorLogging(t)
	if err := closeAggregatorLogging(10 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
}
