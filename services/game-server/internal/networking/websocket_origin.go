package networking

import "net/http"

const trustedWebSocketOrigin = "https://space-rocks-client.local"

var localWebSocketOrigins = map[string]struct{}{
	"http://localhost:8080":  {},
	"http://127.0.0.1:8080": {},
	"http://[::1]:8080":     {},
}

func allowWebSocketOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	if origin == trustedWebSocketOrigin {
		return true
	}
	_, ok := localWebSocketOrigins[origin]
	return ok
}
