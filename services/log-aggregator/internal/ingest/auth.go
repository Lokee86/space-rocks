package ingest

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// BearerTokenAuthorizer returns an authorizer that accepts exactly one bearer
// token. Neither the configured nor received token is exposed by the boundary.
func BearerTokenAuthorizer(token string) RequestAuthorizer {
	return func(r *http.Request) bool {
		const prefix = "Bearer "
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, prefix) {
			return false
		}
		received := strings.TrimSpace(strings.TrimPrefix(header, prefix))
		if received == "" || token == "" {
			return false
		}
		return subtle.ConstantTimeCompare([]byte(received), []byte(token)) == 1
	}
}
