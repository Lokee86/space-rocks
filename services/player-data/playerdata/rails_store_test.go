package playerdata

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func TestNewRailsStore(t *testing.T) {
	t.Run("empty base url", func(t *testing.T) {
		store, err := NewRailsStore(RailsStoreConfig{})
		if err == nil {
			t.Fatal("NewRailsStore returned nil error for empty BaseURL")
		}
		if store != nil {
			t.Fatalf("NewRailsStore returned store %+v for empty BaseURL", store)
		}
	})

	t.Run("trailing slash normalization", func(t *testing.T) {
		store, err := NewRailsStore(RailsStoreConfig{
			BaseURL: "https://example.com/",
		})
		if err != nil {
			t.Fatalf("NewRailsStore returned error: %v", err)
		}
		if store.BaseURL != "https://example.com" {
			t.Fatalf("BaseURL = %q, want %q", store.BaseURL, "https://example.com")
		}
	})
}

func TestRailsStoreNewJSONRequest(t *testing.T) {
	t.Run("internal auth header", func(t *testing.T) {
		store := &RailsStore{
			BaseURL:       "https://example.com",
			internalToken: "internal-token",
			httpClient:    http.DefaultClient,
		}

		request, err := store.newJSONRequest(http.MethodPost, "/internal/accounts", nil)
		if err != nil {
			t.Fatalf("newJSONRequest returned error: %v", err)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer internal-token" {
			t.Fatalf("Authorization = %q, want %q", got, "Bearer internal-token")
		}
		if got := request.Header.Get("Content-Type"); got != "" {
			t.Fatalf("Content-Type = %q, want empty", got)
		}
	})

	t.Run("non-internal bearer auth header", func(t *testing.T) {
		store := &RailsStore{
			BaseURL:     "https://example.com",
			bearerToken: "bearer-token",
			httpClient:  http.DefaultClient,
		}

		request, err := store.newJSONRequest(http.MethodGet, "/accounts", nil)
		if err != nil {
			t.Fatalf("newJSONRequest returned error: %v", err)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer bearer-token" {
			t.Fatalf("Authorization = %q, want %q", got, "Bearer bearer-token")
		}
	})

	t.Run("json content type", func(t *testing.T) {
		store := &RailsStore{
			BaseURL:    "https://example.com",
			httpClient: http.DefaultClient,
		}

		request, err := store.newJSONRequest(http.MethodPost, "/accounts", map[string]string{"name": "Ada"})
		if err != nil {
			t.Fatalf("newJSONRequest returned error: %v", err)
		}
		if got := request.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want %q", got, "application/json")
		}
	})
}

func TestRailsStoreLoadStats(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var gotAuth string
		server := newInMemoryHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			if r.Method != http.MethodGet {
				t.Fatalf("Method = %q, want %q", r.Method, http.MethodGet)
			}
			if r.URL.Path != "/api/player/stats" {
				t.Fatalf("Path = %q, want %q", r.URL.Path, "/api/player/stats")
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"stats":{"total_score":12,"high_score":9,"ship_deaths":3,"games_played":4,"wins":2}}`))
		}))

		store := &RailsStore{
			BaseURL:     server.URL,
			bearerToken: "bearer-token",
			httpClient:  server.Client(),
		}

		stats, found, err := store.LoadStats(protocol.PlayerDataIdentity{
			IdentityKind: IdentityKindAuthenticatedAccount,
			AccountID:    "acct-123",
		})
		if err != nil {
			t.Fatalf("LoadStats returned error: %v", err)
		}
		if !found {
			t.Fatal("LoadStats returned found=false for successful response")
		}
		if stats != (protocol.PlayerDataStats{
			TotalScore:  12,
			HighScore:   9,
			ShipDeaths:  3,
			GamesPlayed: 4,
			Wins:        2,
		}) {
			t.Fatalf("Stats = %+v, want mapped stats", stats)
		}
		if gotAuth != "Bearer bearer-token" {
			t.Fatalf("Authorization = %q, want %q", gotAuth, "Bearer bearer-token")
		}
	})

	t.Run("invalid identity", func(t *testing.T) {
		store := &RailsStore{}
		if _, _, err := store.LoadStats(protocol.PlayerDataIdentity{IdentityKind: IdentityKindGuest}); err == nil {
			t.Fatal("LoadStats returned nil error for invalid identity kind")
		}
		if _, _, err := store.LoadStats(protocol.PlayerDataIdentity{IdentityKind: IdentityKindAuthenticatedAccount}); err == nil {
			t.Fatal("LoadStats returned nil error for missing account id")
		}
	})

	t.Run("missing bearer token", func(t *testing.T) {
		store := &RailsStore{}
		if _, _, err := store.LoadStats(protocol.PlayerDataIdentity{
			IdentityKind: IdentityKindAuthenticatedAccount,
			AccountID:    "acct-123",
		}); err == nil {
			t.Fatal("LoadStats returned nil error for missing bearer token")
		}
	})

	t.Run("non-200", func(t *testing.T) {
		server := newInMemoryHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"nope"}`))
		}))

		store := &RailsStore{
			BaseURL:     server.URL,
			bearerToken: "bearer-token",
			httpClient:  server.Client(),
		}

		if _, _, err := store.LoadStats(protocol.PlayerDataIdentity{
			IdentityKind: IdentityKindAuthenticatedAccount,
			AccountID:    "acct-123",
		}); err == nil {
			t.Fatal("LoadStats returned nil error for non-200 response")
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		server := newInMemoryHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"stats":`))
		}))

		store := &RailsStore{
			BaseURL:     server.URL,
			bearerToken: "bearer-token",
			httpClient:  server.Client(),
		}

		if _, _, err := store.LoadStats(protocol.PlayerDataIdentity{
			IdentityKind: IdentityKindAuthenticatedAccount,
			AccountID:    "acct-123",
		}); err == nil {
			t.Fatal("LoadStats returned nil error for malformed JSON")
		}
	})
}

func TestRailsStoreRecordMatchResult(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var gotAuth string
		var gotPayload struct {
			ResultID   string `json:"result_id"`
			MatchID    string `json:"match_id"`
			AccountID  string `json:"account_id"`
			Score      int    `json:"score"`
			ShipDeaths int    `json:"ship_deaths"`
			Won        bool   `json:"won"`
		}

		server := newInMemoryHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			if r.Method != http.MethodPost {
				t.Fatalf("Method = %q, want %q", r.Method, http.MethodPost)
			}
			if r.URL.Path != "/internal/player-data/match-results" {
				t.Fatalf("Path = %q, want %q", r.URL.Path, "/internal/player-data/match-results")
			}
			if got := r.Header.Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type = %q, want %q", got, "application/json")
			}
			if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
				t.Fatalf("Decode request body: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"accepted":true,"duplicate":false,"stats":{"total_score":12,"high_score":12,"ship_deaths":2,"games_played":1,"wins":1}}`))
		}))

		store := &RailsStore{
			BaseURL:       server.URL,
			internalToken: "internal-token",
			httpClient:    server.Client(),
		}

		stats, duplicate, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			ResultID: "result-1",
			MatchID:  "match-1",
			Identity: protocol.PlayerDataIdentity{
				IdentityKind: IdentityKindAuthenticatedAccount,
				AccountID:    "acct-123",
			},
			Score:      12,
			ShipDeaths: 2,
			Won:        true,
		})
		if err != nil {
			t.Fatalf("RecordMatchResult returned error: %v", err)
		}
		if duplicate {
			t.Fatal("RecordMatchResult returned duplicate=true for non-duplicate response")
		}
		if stats != (protocol.PlayerDataStats{
			TotalScore:  12,
			HighScore:   12,
			ShipDeaths:  2,
			GamesPlayed: 1,
			Wins:        1,
		}) {
			t.Fatalf("Stats = %+v, want mapped stats", stats)
		}
		if gotAuth != "Bearer internal-token" {
			t.Fatalf("Authorization = %q, want %q", gotAuth, "Bearer internal-token")
		}
		if gotPayload != (struct {
			ResultID   string `json:"result_id"`
			MatchID    string `json:"match_id"`
			AccountID  string `json:"account_id"`
			Score      int    `json:"score"`
			ShipDeaths int    `json:"ship_deaths"`
			Won        bool   `json:"won"`
		}{
			ResultID:   "result-1",
			MatchID:    "match-1",
			AccountID:  "acct-123",
			Score:      12,
			ShipDeaths: 2,
			Won:        true,
		}) {
			t.Fatalf("request payload = %+v, want expected payload", gotPayload)
		}
	})

	t.Run("duplicate true", func(t *testing.T) {
		server := newInMemoryHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"accepted":true,"duplicate":true,"stats":{"total_score":7,"high_score":7,"ship_deaths":1,"games_played":1,"wins":0}}`))
		}))

		store := &RailsStore{
			BaseURL:       server.URL,
			internalToken: "internal-token",
			httpClient:    server.Client(),
		}

		_, duplicate, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			ResultID: "result-1",
			MatchID:  "match-1",
			Identity: protocol.PlayerDataIdentity{
				IdentityKind: IdentityKindAuthenticatedAccount,
				AccountID:    "acct-123",
			},
			Score:      7,
			ShipDeaths: 1,
			Won:        false,
		})
		if err != nil {
			t.Fatalf("RecordMatchResult returned error: %v", err)
		}
		if !duplicate {
			t.Fatal("RecordMatchResult returned duplicate=false for duplicate response")
		}
	})

	t.Run("invalid identity", func(t *testing.T) {
		store := &RailsStore{internalToken: "internal-token"}
		if _, _, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			Identity: protocol.PlayerDataIdentity{IdentityKind: IdentityKindGuest},
		}); err == nil {
			t.Fatal("RecordMatchResult returned nil error for invalid identity kind")
		}
		if _, _, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			Identity: protocol.PlayerDataIdentity{IdentityKind: IdentityKindAuthenticatedAccount},
		}); err == nil {
			t.Fatal("RecordMatchResult returned nil error for missing account id")
		}
		if _, _, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			Identity: protocol.PlayerDataIdentity{IdentityKind: IdentityKindAuthenticatedAccount, AccountID: "acct-123"},
			ResultID: "result-1",
		}); err == nil {
			t.Fatal("RecordMatchResult returned nil error for missing match id")
		}
		if _, _, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			Identity: protocol.PlayerDataIdentity{IdentityKind: IdentityKindAuthenticatedAccount, AccountID: "acct-123"},
			MatchID:  "match-1",
		}); err == nil {
			t.Fatal("RecordMatchResult returned nil error for missing result id")
		}
	})

	t.Run("missing internal token", func(t *testing.T) {
		store := &RailsStore{}
		if _, _, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			ResultID: "result-1",
			MatchID:  "match-1",
			Identity: protocol.PlayerDataIdentity{
				IdentityKind: IdentityKindAuthenticatedAccount,
				AccountID:    "acct-123",
			},
		}); err == nil {
			t.Fatal("RecordMatchResult returned nil error for missing internal token")
		}
	})

	t.Run("accepted false", func(t *testing.T) {
		server := newInMemoryHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"accepted":false,"duplicate":false,"error":"rejected","stats":{"total_score":0,"high_score":0,"ship_deaths":0,"games_played":0,"wins":0}}`))
		}))

		store := &RailsStore{
			BaseURL:       server.URL,
			internalToken: "internal-token",
			httpClient:    server.Client(),
		}

		if _, _, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			ResultID: "result-1",
			MatchID:  "match-1",
			Identity: protocol.PlayerDataIdentity{
				IdentityKind: IdentityKindAuthenticatedAccount,
				AccountID:    "acct-123",
			},
		}); err == nil {
			t.Fatal("RecordMatchResult returned nil error for accepted=false response")
		}
	})

	t.Run("non-2xx", func(t *testing.T) {
		server := newInMemoryHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"bad request"}`))
		}))

		store := &RailsStore{
			BaseURL:       server.URL,
			internalToken: "internal-token",
			httpClient:    server.Client(),
		}

		if _, _, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			ResultID: "result-1",
			MatchID:  "match-1",
			Identity: protocol.PlayerDataIdentity{
				IdentityKind: IdentityKindAuthenticatedAccount,
				AccountID:    "acct-123",
			},
		}); err == nil {
			t.Fatal("RecordMatchResult returned nil error for non-2xx response")
		}
	})
}

type inMemoryHTTPListener struct {
	conns  chan net.Conn
	closed chan struct{}
	once   sync.Once
}

func newInMemoryHTTPListener() *inMemoryHTTPListener {
	return &inMemoryHTTPListener{
		conns:  make(chan net.Conn),
		closed: make(chan struct{}),
	}
}

func (l *inMemoryHTTPListener) Accept() (net.Conn, error) {
	select {
	case conn := <-l.conns:
		return conn, nil
	case <-l.closed:
		return nil, net.ErrClosed
	}
}

func (l *inMemoryHTTPListener) Close() error {
	l.once.Do(func() {
		close(l.closed)
	})
	return nil
}

func (l *inMemoryHTTPListener) Addr() net.Addr {
	return inMemoryHTTPAddr("in-memory-listener")
}

type inMemoryHTTPAddr string

func (a inMemoryHTTPAddr) Network() string { return "in-memory" }

func (a inMemoryHTTPAddr) String() string { return string(a) }

func newInMemoryHTTPServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()

	listener := newInMemoryHTTPListener()
	server := &httptest.Server{
		Listener: listener,
		Config:   &http.Server{Handler: handler},
	}
	server.Start()

	server.Client().Transport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			clientConn, serverConn := net.Pipe()
			select {
			case listener.conns <- serverConn:
			case <-ctx.Done():
				_ = clientConn.Close()
				_ = serverConn.Close()
				return nil, ctx.Err()
			}
			return clientConn, nil
		},
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	t.Cleanup(func() {
		server.Close()
		_ = listener.Close()
	})

	return server
}
