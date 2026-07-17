package inbound

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/realtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

type contextCountSession struct {
	context  SessionContext
	calls    int
	outbound [][]byte
}

func (s *contextCountSession) CurrentSessionContext() SessionContext { s.calls++; return s.context }
func (s *contextCountSession) EnqueuePlayerPauseState()              {}
func (s *contextCountSession) EnqueueResyncRequest(SessionContext, realtime.ResyncRequest) bool {
	return true
}
func (s *contextCountSession) SessionID() string {
	return "session"
}
func (s *contextCountSession) ConnectionTraceID() string {
	return "00000000-0000-4000-8000-000000000001"
}
func (s *contextCountSession) EnqueueOutboundMessage(b []byte) { s.outbound = append(s.outbound, b) }

func TestHandlersCaptureSessionContextOnce(t *testing.T) {
	g := game.New()
	control := game.NewControl(g)
	if !control.EnsurePlayerSession("player-1", physics.Vector2{}) {
		t.Fatal("create player")
	}
	room := rooms.NewRoom("room", rooms.RoomStateInGame, g)
	s := &contextCountSession{context: SessionContext{Room: room, RoomID: room.ID, GamePlayerID: "player-1"}}
	if !HandleGameplayPacket(s, game.ClientPacket{Type: game.PacketTypeClientConfig, Config: runtime.ClientConfig{VisibleWorldWidth: 10, VisibleWorldHeight: 10}}) || s.calls != 1 {
		t.Fatalf("gameplay context calls=%d", s.calls)
	}
	s.calls = 0
	if !HandleSimpleDevtoolsPacket(s, "remote", []byte("bad"), ClientPacketEnvelope{Type: "debug_set_score"}) || s.calls != 1 {
		t.Fatalf("devtools context calls=%d", s.calls)
	}
	s.calls = 0
	if !HandleTelemetryPacket(s, "remote", game.ClientPacket{Type: game.PacketTypeTelemetryPing}) || s.calls != 1 {
		t.Fatalf("telemetry context calls=%d", s.calls)
	}
}
