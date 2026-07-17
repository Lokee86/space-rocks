package networking

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/networking/inbound"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

func readClientInput(
	session *webSocketSession,
	remoteAddr string,
	readErr chan<- error,
) {
	for {
		_, msg, err := session.conn.ReadMessage()
		if err != nil {
			readErr <- err
			return
		}

		envelope, err := inbound.DecodeClientPacketEnvelope(msg)
		if err != nil {
			context := session.sessionContext()
			logging.Emit(observability.Request{
				Event: observability.EventNamePacketEnvelopeDecodeFailed,
				Context: observability.Context{
					TraceID:   session.connectionTraceID,
					SessionID: session.sessionID,
					RoomID:    context.RoomID,
					PlayerID:  context.GamePlayerID,
				},
				Fields: observability.Fields{
					"error_code":   "packet_envelope_decode_failed",
					"failure_mode": "packet_envelope_decode_failed",
				},
			})
			continue
		}
		handleClientPacket(session, remoteAddr, msg, envelope)
	}
}
