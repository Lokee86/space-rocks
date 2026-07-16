package networking

func requireAuthenticatedAccount(session *webSocketSession, traceID string) bool {
	if session == nil {
		return false
	}

	if session.authVerifier == nil {
		session.EnqueueRoomError(traceID, "auth_unavailable", "Authentication unavailable.")
		return false
	}

	if session.SessionIdentity().IsAuthenticatedAccount() {
		return true
	}

	session.EnqueueRoomError(traceID, "auth_required", "Authentication required.")
	return false
}
