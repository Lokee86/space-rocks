package devtools

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
)

func TestWrapStatePacketPreservesServerSentMsec(t *testing.T) {
	state := game.StatePacket{
		Type:           game.PacketTypeState,
		ServerSentMsec: 123456789,
	}

	wrappedAny := WrapStatePacket(state, DebugStatus{}, map[string]DebugStatus{})
	wrapped, ok := wrappedAny.(statePacketWithDebugStatus)
	if !ok {
		t.Fatalf("WrapStatePacket returned %T, want statePacketWithDebugStatus", wrappedAny)
	}
	if wrapped.ServerSentMsec != state.ServerSentMsec {
		t.Fatalf("ServerSentMsec = %d, want %d", wrapped.ServerSentMsec, state.ServerSentMsec)
	}
}

func TestWrapStatePacketPreservesPlayerSessions(t *testing.T) {
	state := game.StatePacket{
		Type: game.PacketTypeState,
		PlayerSessions: map[string]game.PlayerSessionState{
			"Player-1": {ID: "Player-1", ShipType: "v_wing", Score: 12, Lives: 3},
		},
	}

	wrappedAny := WrapStatePacket(state, DebugStatus{}, map[string]DebugStatus{})
	wrapped, ok := wrappedAny.(statePacketWithDebugStatus)
	if !ok {
		t.Fatalf("WrapStatePacket returned %T, want statePacketWithDebugStatus", wrappedAny)
	}
	if len(wrapped.PlayerSessions) != len(state.PlayerSessions) {
		t.Fatalf("PlayerSessions len = %d, want %d", len(wrapped.PlayerSessions), len(state.PlayerSessions))
	}
	if wrapped.PlayerSessions["Player-1"].ID != state.PlayerSessions["Player-1"].ID {
		t.Fatalf("PlayerSessions[Player-1].ID = %q, want %q", wrapped.PlayerSessions["Player-1"].ID, state.PlayerSessions["Player-1"].ID)
	}
}

func TestWrapStatePacketPreservesTotalAsteroids(t *testing.T) {
	state := game.StatePacket{
		Type:           game.PacketTypeState,
		TotalAsteroids: 42,
	}

	wrappedAny := WrapStatePacket(state, DebugStatus{}, map[string]DebugStatus{})
	wrapped, ok := wrappedAny.(statePacketWithDebugStatus)
	if !ok {
		t.Fatalf("WrapStatePacket returned %T, want statePacketWithDebugStatus", wrappedAny)
	}
	if wrapped.TotalAsteroids != state.TotalAsteroids {
		t.Fatalf("TotalAsteroids = %d, want %d", wrapped.TotalAsteroids, state.TotalAsteroids)
	}
}
