package networking

type SessionIdentityState string

const (
	SessionIdentityStateGuest               SessionIdentityState = "guest"
	SessionIdentityStateAuthenticatedAccount SessionIdentityState = "authenticated_account"
)

type SessionIdentity struct {
	State         SessionIdentityState
	AccountUserID int64
	DisplayName   string
}

func NewGuestSessionIdentity() SessionIdentity {
	return SessionIdentity{
		State: SessionIdentityStateGuest,
	}
}

func NewAuthenticatedAccountIdentity(userID int64, displayName string) SessionIdentity {
	return SessionIdentity{
		State:         SessionIdentityStateAuthenticatedAccount,
		AccountUserID: userID,
		DisplayName:   displayName,
	}
}

func (identity SessionIdentity) IsAuthenticatedAccount() bool {
	return identity.State == SessionIdentityStateAuthenticatedAccount
}
