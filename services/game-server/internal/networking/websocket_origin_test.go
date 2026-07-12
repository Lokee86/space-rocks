package networking

import (
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
	"github.com/gorilla/websocket"
)

func TestWebSocketOriginPolicy(t *testing.T) {
	tests := []struct {
		name     string
		envSet   bool
		envValue string
		origin   string
		want     bool
	}{
		{name: "default", origin: "https://space-rocks-client.local", want: true},
		{name: "empty origin rejected", origin: "", want: false},
		{name: "unapproved origin rejected", origin: "https://evil.example", want: false},
		{name: "replacement", envSet: true, envValue: " https://allowed.example, ,http://localhost:9000 ", origin: "https://allowed.example", want: true},
		{name: "replacement excludes default", envSet: true, envValue: "https://allowed.example", origin: "http://localhost:8080", want: false},
		{name: "explicit empty replacement", envSet: true, envValue: "", origin: "https://space-rocks-client.local", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			previous, wasSet := os.LookupEnv(websocketAllowedOriginsEnv)
			t.Cleanup(func() {
				if wasSet {
					_ = os.Setenv(websocketAllowedOriginsEnv, previous)
				} else {
					_ = os.Unsetenv(websocketAllowedOriginsEnv)
				}
			})
			if test.envSet {
				_ = os.Setenv(websocketAllowedOriginsEnv, test.envValue)
			} else {
				_ = os.Unsetenv(websocketAllowedOriginsEnv)
			}

			r := httptest.NewRequest("GET", "/ws", nil)
			if test.origin != "" {
				r.Header.Set("Origin", test.origin)
			}
			if got := newWebSocketOriginPolicy().allows(r); got != test.want {
				t.Fatalf("allows() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestWebSocketReadLimitClosesOversizedMessage(t *testing.T) {
	roomManager := rooms.NewRoomManager()
	t.Cleanup(roomManager.StopAll)
	server := httptest.NewServer(WebSocketHandler(roomManager))
	t.Cleanup(server.Close)

	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	serverURL.Scheme = "ws"
	client, _, err := websocket.DefaultDialer.Dial(serverURL.String(), map[string][]string{
		"Origin": {"https://space-rocks-client.local"},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})

	if err := client.WriteMessage(websocket.TextMessage, []byte(strings.Repeat("x", webSocketReadLimit+1))); err != nil {
		t.Fatal(err)
	}
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, _, err := client.ReadMessage(); err == nil {
		t.Fatal("expected oversized message connection close")
	}
}
