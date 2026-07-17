package inbound

import (
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/packetcodec"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

type telemetrySession interface {
	CurrentSessionContext() SessionContext
	SessionID() string
	ConnectionTraceID() string
	EnqueueOutboundMessage([]byte)
}

func HandleTelemetryPacket(session telemetrySession, remoteAddr string, packet game.ClientPacket) bool {
	context := session.CurrentSessionContext()
	if packet.Type != game.PacketTypeTelemetryPing {
		return false
	}
	serverReceivedMsec := time.Now().UnixMilli()
	pong := game.ClientPacket{Type: game.PacketTypeTelemetryPong, Sequence: packet.Sequence, ClientSentMsec: packet.ClientSentMsec, ServerReceivedMsec: int(serverReceivedMsec)}
	pong.ServerSentMsec = int(time.Now().UnixMilli())
	response, err := packetcodec.Encode(pong)
	if err != nil {
		traceID := packet.TraceID
		if traceID == "" {
			traceID = session.ConnectionTraceID()
		}
		logging.Emit(observability.Request{
			Event: observability.EventNameOutboundPacketEncodeFailed,
			Context: observability.Context{
				TraceID:    traceID,
				SessionID:  session.SessionID(),
				RoomID:     context.RoomID,
				PlayerID:   context.GamePlayerID,
				PacketType: game.PacketTypeTelemetryPong,
			},
			Fields: observability.Fields{
				"error_code":   "telemetry_pong_encode_failed",
				"failure_mode": "telemetry_pong_encode_failed",
			},
		})
		return true
	}
	session.EnqueueOutboundMessage(response)
	return true
}
