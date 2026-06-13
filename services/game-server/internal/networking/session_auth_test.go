package networking

import (
	"context"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/authclient"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

type fakeTokenVerifier struct {
	result authclient.VerifyResult
	err    error
}

func (f fakeTokenVerifier) VerifyToken(ctx context.Context, rawToken string) (authclient.VerifyResult, error) {
	return f.result, f.err
}

func TestHandleAuthenticateRequestStoresRailsAccountID(t *testing.T) {
	session := &webSocketSession{
		outbound: make(chan []byte, 1),
		authVerifier: fakeTokenVerifier{
			result: authclient.VerifyResult{
				Valid: true,
				Identity: authclient.Identity{
					UserID:      1,
					AccountID:   "439e2746-9a06-45f1-b36b-b741b5bcfb12",
					DisplayName: "Ada",
				},
			},
		},
		matchResultReporter: rooms.NoopMatchResultReporter{},
	}

	session.handleAuthenticateRequest("submitted-user-token")

	if session.identity.AccountID != "439e2746-9a06-45f1-b36b-b741b5bcfb12" {
		t.Fatalf("expected account id uuid, got %q", session.identity.AccountID)
	}
	if session.identity.AccountUserID != 1 {
		t.Fatalf("expected user id 1, got %d", session.identity.AccountUserID)
	}
	if session.identity.DisplayName != "Ada" {
		t.Fatalf("expected display name Ada, got %q", session.identity.DisplayName)
	}
}
