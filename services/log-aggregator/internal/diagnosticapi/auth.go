package diagnosticapi

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"
	"unicode"
)

// NewBearerTokenAuthorizer returns an authorizer for exactly one configured
// bearer token. Token values are never included in returned errors or output.
func NewBearerTokenAuthorizer(token string) (RequestAuthorizer, error) {
	if token == "" {
		return nil, errors.New("diagnosticapi: bearer token is required")
	}
	if strings.IndexFunc(token, unicode.IsSpace) >= 0 {
		return nil, errors.New("diagnosticapi: bearer token cannot contain whitespace")
	}
	configuredDigest := sha256.Sum256([]byte(token))

	return func(r *http.Request) bool {
		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			return false
		}
		receivedDigest := sha256.Sum256([]byte(parts[1]))
		return subtle.ConstantTimeCompare(receivedDigest[:], configuredDigest[:]) == 1
	}, nil
}
