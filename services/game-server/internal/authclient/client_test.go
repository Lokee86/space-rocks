package authclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Run("blank base url returns error", func(t *testing.T) {
		t.Parallel()

		_, err := New(Config{
			BaseURL:       "",
			InternalToken: "token",
			Timeout:       time.Second,
		})
		if err == nil {
			t.Fatalf("expected error for blank base url")
		}
	})

	t.Run("blank internal token returns error", func(t *testing.T) {
		t.Parallel()

		_, err := New(Config{
			BaseURL:       "https://example.com",
			InternalToken: "",
			Timeout:       time.Second,
		})
		if err == nil {
			t.Fatalf("expected error for blank internal token")
		}
	})

	t.Run("valid config returns a client", func(t *testing.T) {
		t.Parallel()

		client, err := New(Config{
			BaseURL:       "https://example.com",
			InternalToken: "token",
			Timeout:       time.Second,
		})
		if err != nil {
			t.Fatalf("expected client, got error: %v", err)
		}
		if client == nil {
			t.Fatalf("expected client, got nil")
		}
	})

	t.Run("trailing slash is accepted", func(t *testing.T) {
		t.Parallel()

		client, err := New(Config{
			BaseURL:       "https://example.com/",
			InternalToken: "token",
			Timeout:       time.Second,
		})
		if err != nil {
			t.Fatalf("expected client, got error: %v", err)
		}
		if client.baseURL != "https://example.com" {
			t.Fatalf("expected trimmed base url, got %q", client.baseURL)
		}
	})

	t.Run("zero timeout gets default timeout", func(t *testing.T) {
		t.Parallel()

		client, err := New(Config{
			BaseURL:       "https://example.com",
			InternalToken: "token",
			Timeout:       0,
		})
		if err != nil {
			t.Fatalf("expected client, got error: %v", err)
		}
		if client.httpClient == nil {
			t.Fatalf("expected http client, got nil")
		}
		if got := client.httpClient.Timeout; got != 2*time.Second {
			t.Fatalf("expected default timeout of 2s, got %v", got)
		}
	})
}

func TestVerifyTokenRequestShape(t *testing.T) {
	t.Parallel()

	var gotMethod string
	var gotPath string
	var gotAuthorization string
	var gotToken string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuthorization = r.Header.Get("Authorization")

		defer r.Body.Close()

		var payload struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		gotToken = payload.Token

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"valid":false}`)
	}))
	defer server.Close()

	client, err := New(Config{
		BaseURL:       server.URL,
		InternalToken: "test-internal-token",
		Timeout:       time.Second,
	})
	if err != nil {
		t.Fatalf("expected client, got error: %v", err)
	}

	result, err := client.VerifyToken(t.Context(), "submitted-user-token")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result.Valid {
		t.Fatalf("expected invalid result")
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("expected POST method, got %q", gotMethod)
	}
	if gotPath != "/internal/auth/verify-token" {
		t.Fatalf("expected verify-token path, got %q", gotPath)
	}
	if gotAuthorization != "Bearer test-internal-token" {
		t.Fatalf("expected bearer auth header, got %q", gotAuthorization)
	}
	if gotToken != "submitted-user-token" {
		t.Fatalf("expected submitted token in request body, got %q", gotToken)
	}
}

func TestVerifyTokenValidResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"valid":true,"user":{"id":123,"display_name":"Ada"}}`)
	}))
	defer server.Close()

	client, err := New(Config{
		BaseURL:       server.URL,
		InternalToken: "test-internal-token",
		Timeout:       time.Second,
	})
	if err != nil {
		t.Fatalf("expected client, got error: %v", err)
	}

	result, err := client.VerifyToken(t.Context(), "submitted-user-token")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid result")
	}
	if result.Identity.UserID != 123 {
		t.Fatalf("expected user id 123, got %d", result.Identity.UserID)
	}
	if result.Identity.DisplayName != "Ada" {
		t.Fatalf("expected display name Ada, got %q", result.Identity.DisplayName)
	}
}

func TestVerifyTokenFailureModes(t *testing.T) {
	t.Parallel()

	t.Run("rails 401 returns an error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, `{"error":"unauthorized"}`)
		}))
		defer server.Close()

		client, err := New(Config{
			BaseURL:       server.URL,
			InternalToken: "test-internal-token",
			Timeout:       time.Second,
		})
		if err != nil {
			t.Fatalf("expected client, got error: %v", err)
		}

		result, err := client.VerifyToken(t.Context(), "submitted-user-token")
		if err == nil {
			t.Fatalf("expected error for 401 response")
		}
		if result.Valid {
			t.Fatalf("expected invalid zero result on error")
		}
	})

	t.Run("rails 500 returns an error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = io.WriteString(w, `{"error":"server error"}`)
		}))
		defer server.Close()

		client, err := New(Config{
			BaseURL:       server.URL,
			InternalToken: "test-internal-token",
			Timeout:       time.Second,
		})
		if err != nil {
			t.Fatalf("expected client, got error: %v", err)
		}

		result, err := client.VerifyToken(t.Context(), "submitted-user-token")
		if err == nil {
			t.Fatalf("expected error for 500 response")
		}
		if result.Valid {
			t.Fatalf("expected invalid zero result on error")
		}
	})

	t.Run("malformed json returns an error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"valid":true,`)
		}))
		defer server.Close()

		client, err := New(Config{
			BaseURL:       server.URL,
			InternalToken: "test-internal-token",
			Timeout:       time.Second,
		})
		if err != nil {
			t.Fatalf("expected client, got error: %v", err)
		}

		result, err := client.VerifyToken(t.Context(), "submitted-user-token")
		if err == nil {
			t.Fatalf("expected error for malformed json")
		}
		if result.Valid {
			t.Fatalf("expected invalid zero result on error")
		}
	})

	t.Run("request context cancellation returns an error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, `{"valid":false}`)
		}))
		defer server.Close()

		client, err := New(Config{
			BaseURL:       server.URL,
			InternalToken: "test-internal-token",
			Timeout:       time.Second,
		})
		if err != nil {
			t.Fatalf("expected client, got error: %v", err)
		}

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		result, err := client.VerifyToken(ctx, "submitted-user-token")
		if err == nil {
			t.Fatalf("expected error for canceled context")
		}
		if !strings.Contains(err.Error(), context.Canceled.Error()) {
			t.Fatalf("expected context canceled error, got %v", err)
		}
		if result.Valid {
			t.Fatalf("expected invalid zero result on error")
		}
	})
}
