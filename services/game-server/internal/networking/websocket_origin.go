package networking

import (
	"net/http"
	"os"
	"strings"
)

const websocketAllowedOriginsEnv = "SPACE_ROCKS_WEBSOCKET_ALLOWED_ORIGINS"

var defaultWebSocketOrigins = []string{
	"https://space-rocks-client.local",
	"https://space-rocks.laughingskull.ca",
	"http://localhost",
	"http://127.0.0.1",
	"http://[::1]",
	"http://localhost:8080",
	"http://127.0.0.1:8080",
	"http://[::1]:8080",
}

type webSocketOriginPolicy map[string]struct{}

func newWebSocketOriginPolicy() webSocketOriginPolicy {
	value, ok := os.LookupEnv(websocketAllowedOriginsEnv)
	if !ok {
		value = strings.Join(defaultWebSocketOrigins, ",")
	}
	policy := webSocketOriginPolicy{}
	for _, origin := range strings.Split(value, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			policy[origin] = struct{}{}
		}
	}
	return policy
}

func (p webSocketOriginPolicy) allows(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}
	_, ok := p[origin]
	return ok
}
