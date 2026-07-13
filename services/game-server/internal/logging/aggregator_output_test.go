package logging

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/logging/aggregatorclient"
)

func aggregatorTestConfig(endpoint string) aggregatorclient.Config {
	return aggregatorclient.Config{Enabled: true, EndpointURL: endpoint, QueueCapacity: 16, BatchSize: 1, FlushInterval: time.Hour, RequestTimeout: time.Second, SpoolDirectory: "", SpoolByteCap: 1 << 20, SpoolEnabled: false}
}

func resetAggregator(t *testing.T) {
	t.Helper()
	if err := CloseAggregatorOutput(context.Background()); err != nil {
		t.Fatal(err)
	}
	Configure("warn")
}

func TestAggregatorOutputSendsStructuredEvents(t *testing.T) {
	var mu sync.Mutex
	var records []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var batch struct {
			Events []json.RawMessage `json:"events"`
		}
		if err := json.NewDecoder(bufio.NewReader(r.Body)).Decode(&batch); err != nil {
			t.Error(err)
			return
		}
		for _, event := range batch.Events {
			var record map[string]any
			_ = json.Unmarshal(event, &record)
			mu.Lock()
			records = append(records, record)
			mu.Unlock()
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	defer resetAggregator(t)
	Configure("info")
	if err := ConfigureAggregatorOutput(aggregatorTestConfig(server.URL)); err != nil {
		t.Fatal(err)
	}
	Game.Info("aggregated", "value", 7)
	if err := CloseAggregatorOutput(context.Background()); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(records) != 1 || records[0]["msg"] != "aggregated" || records[0][FieldCategory] != CategoryGame {
		t.Fatalf("records = %#v", records)
	}
}

func TestDisabledAggregatorSendsNothing(t *testing.T) {
	calls := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls <- struct{}{}; w.WriteHeader(http.StatusNoContent) }))
	defer server.Close()
	defer resetAggregator(t)
	Configure("info")
	if err := ConfigureAggregatorOutput(aggregatorclient.Config{}); err != nil {
		t.Fatal(err)
	}
	Info("not aggregated")
	select {
	case <-calls:
		t.Fatal("disabled aggregator received a request")
	case <-time.After(30 * time.Millisecond):
	}
}

func TestAggregatorCloseDrainsQueuedRecords(t *testing.T) {
	calls := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls <- struct{}{}; w.WriteHeader(http.StatusNoContent) }))
	defer server.Close()
	defer resetAggregator(t)
	Configure("info")
	config := aggregatorTestConfig(server.URL)
	config.BatchSize = 50
	if err := ConfigureAggregatorOutput(config); err != nil {
		t.Fatal(err)
	}
	Info("drain me")
	if err := CloseAggregatorOutput(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-calls:
	case <-time.After(time.Second):
		t.Fatal("queued record was not delivered")
	}
}

func TestAggregatorReplacementDetachesOldSink(t *testing.T) {
	var mu sync.Mutex
	var oldMessages, newMessages []string
	decode := func(messages *[]string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var batch struct {
				Events []json.RawMessage `json:"events"`
			}
			_ = json.NewDecoder(r.Body).Decode(&batch)
			for _, event := range batch.Events {
				var record map[string]any
				_ = json.Unmarshal(event, &record)
				mu.Lock()
				*messages = append(*messages, record["msg"].(string))
				mu.Unlock()
			}
			w.WriteHeader(http.StatusNoContent)
		}
	}
	oldServer := httptest.NewServer(decode(&oldMessages))
	defer oldServer.Close()
	newServer := httptest.NewServer(decode(&newMessages))
	defer newServer.Close()
	defer resetAggregator(t)
	Configure("info")
	if err := ConfigureAggregatorOutput(aggregatorTestConfig(oldServer.URL)); err != nil {
		t.Fatal(err)
	}
	Game.Info("before replacement")
	if err := ConfigureAggregatorOutput(aggregatorTestConfig(newServer.URL)); err != nil {
		t.Fatal(err)
	}
	Game.Info("after replacement")
	if err := CloseAggregatorOutput(context.Background()); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(newMessages) != 1 || newMessages[0] != "after replacement" {
		t.Fatalf("new messages = %#v", newMessages)
	}
	for _, message := range oldMessages {
		if message == "after replacement" {
			t.Fatalf("old sink received replacement record: %#v", oldMessages)
		}
	}
}
