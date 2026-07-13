package diagnosticapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewBearerTokenAuthorizer(t *testing.T) {
	authorize, err := NewBearerTokenAuthorizer("configured-secret")
	if err != nil {
		t.Fatalf("NewBearerTokenAuthorizer() error = %v", err)
	}

	tests := map[string]struct {
		header string
		want   bool
	}{
		"valid":                   {header: "Bearer configured-secret", want: true},
		"case-insensitive scheme": {header: "bEaReR configured-secret", want: true},
		"missing":                 {},
		"malformed scheme":        {header: "Basic configured-secret"},
		"malformed spacing":       {header: "Bearer"},
		"extra credentials":       {header: "Bearer configured-secret extra"},
		"wrong token":             {header: "Bearer received-secret"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", test.header)
			if got := authorize(req); got != test.want {
				t.Fatalf("authorized=%v, want %v", got, test.want)
			}
		})
	}
}

func TestNewBearerTokenAuthorizerRejectsBlankConfiguredToken(t *testing.T) {
	for _, token := range []string{"", " ", "\t\n"} {
		t.Run(strings.ReplaceAll(token, "\n", "newline"), func(t *testing.T) {
			if authorize, err := NewBearerTokenAuthorizer(token); err == nil || authorize != nil {
				t.Fatalf("got authorizer=%v, error=%v; want rejection", authorize, err)
			}
		})
	}
}

func TestNewBearerTokenAuthorizerRejectsWhitespaceConfiguredToken(t *testing.T) {
	for _, token := range []string{" leading", "trailing ", "embedded secret", "embedded	tab"} {
		t.Run(strings.ReplaceAll(token, " ", "space"), func(t *testing.T) {
			if authorize, err := NewBearerTokenAuthorizer(token); err == nil || authorize != nil {
				t.Fatalf("got authorizer=%v, error=%v; want rejection", authorize, err)
			}
		})
	}
}
