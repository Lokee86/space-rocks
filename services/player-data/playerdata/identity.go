package playerdata

import "github.com/Lokee86/space-rocks/player-data/protocol"

const (
	IdentityKindAuthenticatedAccount = "authenticated_account"
	IdentityKindLocalProfile         = "local_profile"
	IdentityKindGuest                = "guest"
)

func IdentityKey(identity protocol.PlayerDataIdentity) string {
	switch identity.IdentityKind {
	case IdentityKindAuthenticatedAccount:
		if identity.AccountID == "" {
			return ""
		}
		return "account:" + identity.AccountID
	case IdentityKindLocalProfile:
		if identity.LocalProfileID == "" {
			return ""
		}
		return "local:" + identity.LocalProfileID
	case IdentityKindGuest:
		return "guest"
	default:
		return ""
	}
}
