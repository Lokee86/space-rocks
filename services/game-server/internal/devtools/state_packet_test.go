package devtools

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
	runtime "github.com/Lokee86/space-rocks/server/internal/game/runtime"
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

func TestWrapStatePacketPreservesPickups(t *testing.T) {
	state := game.StatePacket{
		Type: game.PacketTypeState,
		Pickups: map[string]runtime.PickupState{
			"pickup-1": {ID: "pickup-1", Type: "shield", X: 12, Y: 34},
		},
	}

	wrappedAny := WrapStatePacket(state, DebugStatus{}, map[string]DebugStatus{})
	wrapped, ok := wrappedAny.(statePacketWithDebugStatus)
	if !ok {
		t.Fatalf("WrapStatePacket returned %T, want statePacketWithDebugStatus", wrappedAny)
	}
	if len(wrapped.Pickups) != len(state.Pickups) {
		t.Fatalf("Pickups len = %d, want %d", len(wrapped.Pickups), len(state.Pickups))
	}
	if wrapped.Pickups["pickup-1"].ID != state.Pickups["pickup-1"].ID {
		t.Fatalf("Pickups[pickup-1].ID = %q, want %q", wrapped.Pickups["pickup-1"].ID, state.Pickups["pickup-1"].ID)
	}
	if wrapped.Pickups["pickup-1"].Type != state.Pickups["pickup-1"].Type {
		t.Fatalf("Pickups[pickup-1].Type = %q, want %q", wrapped.Pickups["pickup-1"].Type, state.Pickups["pickup-1"].Type)
	}
}

func TestWrapStatePacketPreservesEntityMaps(t *testing.T) {
	state := game.StatePacket{
		Type: game.PacketTypeState,
		Players: map[string]runtime.ShipState{
			"Player-1": {ID: "Player-1"},
		},
		Bullets: map[string]runtime.BulletState{
			"bullet-1": {ID: "bullet-1"},
		},
		Asteroids: map[string]runtime.AsteroidState{
			"asteroid-1": {ID: "asteroid-1"},
		},
		Pickups: map[string]runtime.PickupState{
			"pickup-1": {ID: "pickup-1", Type: "shield"},
		},
	}

	wrappedAny := WrapStatePacket(state, DebugStatus{}, map[string]DebugStatus{})
	wrapped, ok := wrappedAny.(statePacketWithDebugStatus)
	if !ok {
		t.Fatalf("WrapStatePacket returned %T, want statePacketWithDebugStatus", wrappedAny)
	}
	if len(wrapped.Players) != 1 {
		t.Fatalf("Players len = %d, want 1", len(wrapped.Players))
	}
	if len(wrapped.Bullets) != 1 {
		t.Fatalf("Bullets len = %d, want 1", len(wrapped.Bullets))
	}
	if len(wrapped.Asteroids) != 1 {
		t.Fatalf("Asteroids len = %d, want 1", len(wrapped.Asteroids))
	}
	if len(wrapped.Pickups) != 1 {
		t.Fatalf("Pickups len = %d, want 1", len(wrapped.Pickups))
	}
}
