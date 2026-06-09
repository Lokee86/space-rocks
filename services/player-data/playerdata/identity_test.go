package playerdata

import (
	"testing"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func TestIdentityKey(t *testing.T) {
	t.Run("account", func(t *testing.T) {
		got := IdentityKey(protocol.PlayerDataIdentity{
			IdentityKind: IdentityKindAuthenticatedAccount,
			AccountID:    "acct-123",
		})
		want := "account:acct-123"
		if got != want {
			t.Fatalf("IdentityKey() = %q, want %q", got, want)
		}
	})

	t.Run("local", func(t *testing.T) {
		got := IdentityKey(protocol.PlayerDataIdentity{
			IdentityKind:   IdentityKindLocalProfile,
			LocalProfileID: "local-456",
		})
		want := "local:local-456"
		if got != want {
			t.Fatalf("IdentityKey() = %q, want %q", got, want)
		}
	})

	t.Run("guest", func(t *testing.T) {
		got := IdentityKey(protocol.PlayerDataIdentity{
			IdentityKind: IdentityKindGuest,
		})
		want := "guest"
		if got != want {
			t.Fatalf("IdentityKey() = %q, want %q", got, want)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		cases := []protocol.PlayerDataIdentity{
			{IdentityKind: IdentityKindAuthenticatedAccount},
			{IdentityKind: IdentityKindLocalProfile},
			{IdentityKind: "unknown"},
		}

		for _, identity := range cases {
			if got := IdentityKey(identity); got != "" {
				t.Fatalf("IdentityKey(%+v) = %q, want empty string", identity, got)
			}
		}
	})
}
