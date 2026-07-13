package aggregatorclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEncodeBatchPayloadShapeAndUUID(t *testing.T) {
	payload, err := encodeBatch([][]byte{[]byte(`{"kind":"one"}`), []byte(`{"kind":"two","value":2}`)})
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		BatchID string            `json:"batch_id"`
		Events  []json.RawMessage `json:"events"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	if _, err := uuid.Parse(decoded.BatchID); err != nil {
		t.Fatalf("batch_id is not UUID: %v", err)
	}
	if len(decoded.Events) != 2 || string(decoded.Events[0]) != `{"kind":"one"}` {
		t.Fatalf("unexpected events: %s", payload)
	}
}

func TestEncodeBatchRejectsInvalidJSON(t *testing.T) {
	if _, err := encodeBatch([][]byte{[]byte(`{"valid":true}`), []byte(`{invalid}`)}); err == nil {
		t.Fatal("expected invalid JSON error")
	}
}

func TestHTTPBatchSenderHeadersAndBearer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Errorf("method = %s", request.Method)
		}
		if request.Header.Get("Content-Type") != "application/json" {
			t.Errorf("content type = %q", request.Header.Get("Content-Type"))
		}
		if request.Header.Get("Authorization") != "Bearer secret" {
			t.Errorf("authorization = %q", request.Header.Get("Authorization"))
		}
		response.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()
	sender := newHTTPBatchSender(Config{EndpointURL: server.URL, BearerToken: "secret", RequestTimeout: time.Second})
	if err := sender.send(context.Background(), []byte(`{"events":[]}`)); err != nil {
		t.Fatal(err)
	}
}

func TestHTTPBatchSenderNon2xxIncludesBoundedBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.WriteHeader(http.StatusBadGateway)
		_, _ = response.Write([]byte(strings.Repeat("x", 10000)))
	}))
	defer server.Close()
	sender := newHTTPBatchSender(Config{EndpointURL: server.URL, RequestTimeout: time.Second})
	err := sender.send(context.Background(), []byte(`{}`))
	if err == nil || !strings.Contains(err.Error(), "502") {
		t.Fatalf("error = %v", err)
	}
	if len(err.Error()) > 4200 {
		t.Fatalf("error body was not bounded: %d bytes", len(err.Error()))
	}
}

func TestHTTPBatchSenderRequestTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) { time.Sleep(200 * time.Millisecond) }))
	defer server.Close()
	sender := newHTTPBatchSender(Config{EndpointURL: server.URL, RequestTimeout: 20 * time.Millisecond})
	if err := sender.send(context.Background(), []byte(`{}`)); err == nil {
		t.Fatal("expected timeout error")
	}
}
