package ingest

import (
	"net/http/httptest"
	"testing"
)

func TestBearerTokenAuthorizer(t *testing.T) {
	authorize := BearerTokenAuthorizer("configured-secret")
	for name, header := range map[string]string{
		"missing":      "",
		"wrong scheme": "Basic configured-secret",
		"wrong token":  "Bearer received-secret",
		"empty token":  "Bearer ",
		"valid":        "Bearer configured-secret",
	} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", nil)
			req.Header.Set("Authorization", header)
			want := name == "valid"
			if got := authorize(req); got != want {
				t.Fatalf("authorized=%v, want %v", got, want)
			}
		})
	}
}

func TestBearerTokenAuthorizerDoesNotExposeTokens(t *testing.T) {
	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Authorization", "Bearer received-secret")
	if BearerTokenAuthorizer("configured-secret")(req) {
		t.Fatal("unexpected authorization")
	}
}
