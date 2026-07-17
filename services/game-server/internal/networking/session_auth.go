package networking

import (
	"context"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/authclient"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/packetcodec"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
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
	TraceID       string `json:"trace_id,omitempty"`
	Message       string `json:"message,omitempty"`
}

const authenticateRequestTimeout = 2 * time.Second

func (session *webSocketSession) EnqueueAuthenticateResult(result authenticateResultPacket) {
	payload, err := packetcodec.Encode(result)
	if err != nil {
		traceID := result.TraceID
		if traceID == "" {
			traceID = session.connectionTraceID
		}
		logging.Emit(observability.Request{
			Event: observability.EventNameOutboundPacketEncodeFailed,
			Context: observability.Context{
				TraceID:    traceID,
				SessionID:  session.sessionID,
				PacketType: "authenticate_result",
			},
			Fields: observability.Fields{
				"error_code":   "authenticate_result_encode_failed",
				"failure_mode": "authenticate_result_encode_failed",
			},
		})
		return
	}

	session.enqueue(payload)
}

func (session *webSocketSession) handleAuthenticateRequest(rawToken string, traceID string) {
	if rawToken == "" {
		session.EnqueueAuthenticateResult(authenticateResultPacket{
			Type:          "authenticate_result",
			TraceID:       traceID,
			Authenticated: false,
			ErrorCode:     "invalid_token",
		})
		return
	}

	if session.authVerifier == nil {
		session.EnqueueAuthenticateResult(authenticateResultPacket{
			Type:          "authenticate_result",
			TraceID:       traceID,
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
			TraceID:       traceID,
			Authenticated: false,
			ErrorCode:     "token_verification_unavailable",
		})
		return
	}

	if !result.Valid {
		session.EnqueueAuthenticateResult(authenticateResultPacket{
			Type:          "authenticate_result",
			TraceID:       traceID,
			Authenticated: false,
			ErrorCode:     "invalid_token",
		})
		return
	}

	session.SetAuthenticatedAccountIdentity(result.Identity.UserID, result.Identity.AccountID, result.Identity.DisplayName)
	session.EnqueueAuthenticateResult(authenticateResultPacket{
		Type:          "authenticate_result",
		TraceID:       traceID,
		Authenticated: true,
		UserID:        result.Identity.UserID,
		DisplayName:   result.Identity.DisplayName,
	})
}
