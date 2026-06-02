package devtools

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
	player "github.com/Lokee86/space-rocks/server/internal/game/player"
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

func TestWrapStatePacketPreservesPlayerWorldStates(t *testing.T) {
	state := game.StatePacket{
		Type: game.PacketTypeState,
		PlayerWorldStates: map[string]player.WorldState{
			"Player-1": {ID: "Player-1", Status: player.StatusActive},
		},
	}

	wrappedAny := WrapStatePacket(state, DebugStatus{}, map[string]DebugStatus{})
	wrapped, ok := wrappedAny.(statePacketWithDebugStatus)
	if !ok {
		t.Fatalf("WrapStatePacket returned %T, want statePacketWithDebugStatus", wrappedAny)
	}
	if len(wrapped.PlayerWorldStates) != len(state.PlayerWorldStates) {
		t.Fatalf("PlayerWorldStates len = %d, want %d", len(wrapped.PlayerWorldStates), len(state.PlayerWorldStates))
	}
	if wrapped.PlayerWorldStates["Player-1"].ID != state.PlayerWorldStates["Player-1"].ID {
		t.Fatalf("PlayerWorldStates[Player-1].ID = %q, want %q", wrapped.PlayerWorldStates["Player-1"].ID, state.PlayerWorldStates["Player-1"].ID)
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
