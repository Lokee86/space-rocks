package networking

import (
	"context"
	"time"

	"github.com/Lokee86/space-rocks/server/internal/authclient"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
)

type TokenVerifier interface {
	VerifyToken(ctx context.Context, rawToken string) (authclient.VerifyResult, error)
}

type authenticateResultPacket struct {
	Type          string `json:"type"`
	Authenticated bool   `json:"authenticated"`
	UserID        int64  `json:"user_id,omitempty"`
	DisplayName   string `json:"display_name,omitempty"`
	ErrorCode     string `json:"error_code,omitempty"`
	Message       string `json:"message,omitempty"`
}

const authenticateRequestTimeout = 2 * time.Second

func (session *webSocketSession) EnqueueAuthenticateResult(result authenticateResultPacket) {
	payload, err := packetcodec.Encode(result)
	if err != nil {
		logging.Network.Error("authenticate result marshal failed", err,
			"session_id", session.sessionID,
			"authenticated", result.Authenticated,
			"user_id", result.UserID,
		)
		return
	}

	session.enqueue(payload)
}

func (session *webSocketSession) handleAuthenticateRequest(rawToken string) {
	if rawToken == "" {
		session.EnqueueAuthenticateResult(authenticateResultPacket{
			Type:          "authenticate_result",
			Authenticated: false,
			ErrorCode:     "invalid_token",
		})
		return
	}

	if session.authVerifier == nil {
		session.EnqueueAuthenticateResult(authenticateResultPacket{
			Type:          "authenticate_result",
			Authenticated: false,
			ErrorCode:     "token_verification_unavailable",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), authenticateRequestTimeout)
	defer cancel()

	result, err := session.authVerifier.VerifyToken(ctx, rawToken)
	if err != nil {
		session.EnqueueAuthenticateResult(authenticateResultPacket{
			Type:          "authenticate_result",
			Authenticated: false,
			ErrorCode:     "token_verification_unavailable",
		})
		return
	}

	if !result.Valid {
		session.EnqueueAuthenticateResult(authenticateResultPacket{
			Type:          "authenticate_result",
			Authenticated: false,
			ErrorCode:     "invalid_token",
		})
		return
	}

	session.SetAuthenticatedAccountIdentity(result.Identity.UserID, result.Identity.DisplayName)
	session.EnqueueAuthenticateResult(authenticateResultPacket{
		Type:          "authenticate_result",
		Authenticated: true,
		UserID:        result.Identity.UserID,
		DisplayName:   result.Identity.DisplayName,
	})
}
