package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func TestHandlePacketCoalescesPlayerInputUntilSimulationTick(t *testing.T) {
	game := NewWithSeed(1)
	playerID := game.AddPlayer()
	player := game.entities.Players[playerID]

	game.HandlePacket(playerID, ClientPacket{
		Type:  PacketTypeInput,
		Input: runtime.InputState{Forward: true},
	})
	game.HandlePacket(playerID, ClientPacket{
		Type:  PacketTypeInput,
		Input: runtime.InputState{Left: true, PrimaryFire: true},
	})

	if player.Input.Forward || player.Input.Left || player.Input.PrimaryFire {
		t.Fatal("expected inbound input to remain pending before the simulation tick")
	}

	game.mu.Lock()
	game.applyPendingPlayerInputsLocked()
	game.mu.Unlock()

	if player.Input.Forward {
		t.Fatal("expected the superseded input state to be discarded")
	}
	if !player.Input.Left || !player.Input.PrimaryFire {
		t.Fatalf("applied input = %+v, want latest queued state", player.Input)
	}
	if len(game.pendingPlayerInputs) != 0 {
		t.Fatalf("pending input count = %d, want 0 after tick application", len(game.pendingPlayerInputs))
	}
}

func TestRemovePlayerClearsPendingInput(t *testing.T) {
	game := NewWithSeed(1)
	playerID := game.AddPlayer()

	game.HandlePacket(playerID, ClientPacket{
		Type:  PacketTypeInput,
		Input: runtime.InputState{Forward: true},
	})
	game.RemovePlayer(playerID)

	game.inputMu.Lock()
	_, pending := game.pendingPlayerInputs[playerID]
	game.inputMu.Unlock()
	if pending {
		t.Fatal("expected removing a player to clear its pending input")
	}
}
