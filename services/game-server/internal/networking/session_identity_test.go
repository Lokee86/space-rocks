package networking

import "testing"

func TestNewGuestSessionIdentity(t *testing.T) {
	identity := NewGuestSessionIdentity()

	if identity.State != SessionIdentityStateGuest {
		t.Fatalf("expected guest state, got %q", identity.State)
	}
}

func TestGuestSessionIdentityIsNotAuthenticated(t *testing.T) {
	identity := NewGuestSessionIdentity()

	if identity.IsAuthenticatedAccount() {
		t.Fatalf("expected guest identity to not be authenticated")
	}
}

func TestNewAuthenticatedAccountIdentity(t *testing.T) {
	identity := NewAuthenticatedAccountIdentity(123, "11111111-2222-3333-4444-555555555555", "Ada")

	if identity.State != SessionIdentityStateAuthenticatedAccount {
		t.Fatalf("expected authenticated account state, got %q", identity.State)
	}
	if identity.AccountUserID != 123 {
		t.Fatalf("expected user id 123, got %d", identity.AccountUserID)
	}
	if identity.AccountID != "11111111-2222-3333-4444-555555555555" {
		t.Fatalf("expected account id uuid, got %q", identity.AccountID)
	}
	if identity.DisplayName != "Ada" {
		t.Fatalf("expected display name Ada, got %q", identity.DisplayName)
	}
}

func TestAuthenticatedAccountIdentityIsAuthenticated(t *testing.T) {
	identity := NewAuthenticatedAccountIdentity(123, "11111111-2222-3333-4444-555555555555", "Ada")

	if !identity.IsAuthenticatedAccount() {
		t.Fatalf("expected authenticated account identity to be authenticated")
	}
}

func TestAccountIDForAuthenticatedSessionUsesRailsAccountID(t *testing.T) {
	session := &webSocketSession{
		identity: SessionIdentity{
			State:         SessionIdentityStateAuthenticatedAccount,
			AccountUserID: 123,
			AccountID:     "acct-uuid-1",
			DisplayName:   "Ada",
		},
	}

	if got := accountIDForSession(session); got != "acct-uuid-1" {
		t.Fatalf("expected account id acct-uuid-1, got %q", got)
	}
}
