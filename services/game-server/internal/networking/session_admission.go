package networking

func requireAuthenticatedAccount(session *webSocketSession) bool {
	if session == nil {
		return false
	}

	if session.authVerifier == nil {
		session.EnqueueRoomError("auth_unavailable", "Authentication unavailable.")
		return false
	}

	if session.SessionIdentity().IsAuthenticatedAccount() {
		return true
	}

	session.EnqueueRoomError("auth_required", "Authentication required.")
	return false
}
