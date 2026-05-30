package gametests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	servergame "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestSuspensionStateReflectsPauseAndFreeze(t *testing.T) {
	state := entities.SuspensionState{}
	if state.IsSuspended() {
		t.Fatal("expected state without pause or freeze to be active")
	}

	state.SetPaused(true)
	if !state.IsSuspended() {
		t.Fatal("expected paused state to be suspended")
	}

	state.SetPaused(false)
	state.SetDevFrozen(true)
	if !state.IsSuspended() {
		t.Fatal("expected frozen state to be suspended")
	}
}

func TestPausedAndFrozenSuspensionRequiresBothCausesCleared(t *testing.T) {
	state := entities.SuspensionState{}
	state.SetPaused(true)
	state.SetDevFrozen(true)

	state.SetPaused(false)
	if !state.DevFrozen {
		t.Fatal("expected resume not to clear player freeze")
	}
	if !state.IsSuspended() {
		t.Fatal("expected resumed state to remain suspended while frozen")
	}

	state.SetPaused(true)
	state.SetDevFrozen(false)
	if !state.Paused {
		t.Fatal("expected unfreeze not to clear pause")
	}
	if !state.IsSuspended() {
		t.Fatal("expected unfrozen state to remain suspended while paused")
	}

	state.SetPaused(false)
	if state.IsSuspended() {
		t.Fatal("expected state to be active after pause and freeze are cleared")
	}
}

func TestPauseRequestToggleClearsInputAndIgnoresNewInput(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	start := scenario.playerState(playerID, playerID)

	scenario.send(playerID, servergame.ClientPacket{
		Type:  servergame.PacketTypeInput,
		Input: entities.InputState{Forward: true, Shoot: true},
	})
	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypePauseRequest})

	paused, ok := scenario.game.PlayerPauseStatePacket(playerID)
	if !ok {
		t.Fatal("expected pause state packet after pause")
	}
	if !paused.Paused {
		t.Fatal("expected player to be paused")
	}

	scenario.send(playerID, servergame.ClientPacket{
		Type:  servergame.PacketTypeInput,
		Input: entities.InputState{Forward: true, Shoot: true},
	})
	scenario.step(1.0 / float64(constants.ServerTickRate))

	packet := scenario.state(playerID)
	player := packet.Players[playerID]
	if player.X != start.X || player.Y != start.Y {
		t.Fatalf("expected paused player to stay at (%v, %v), got (%v, %v)", start.X, start.Y, player.X, player.Y)
	}
	if len(packet.Bullets) != 0 {
		t.Fatalf("expected paused player not to shoot, got %d bullets", len(packet.Bullets))
	}
}

func TestPauseRequestPacketTogglesPauseState(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypePauseRequest})

	paused, ok := scenario.game.PlayerPauseStatePacket(playerID)
	if !ok {
		t.Fatal("expected pause state packet after pause request")
	}
	if !paused.Paused {
		t.Fatal("expected pause request to pause player")
	}

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypePauseRequest})

	resumed, ok := scenario.game.PlayerPauseStatePacket(playerID)
	if !ok {
		t.Fatal("expected pause state packet after second pause request")
	}
	if resumed.Paused {
		t.Fatal("expected second pause request to resume player")
	}
}

func TestPlayerPauseStatePacketReflectsPauseRequestToggle(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	fresh, ok := scenario.game.PlayerPauseStatePacket(playerID)
	if !ok {
		t.Fatal("expected pause state packet for fresh player")
	}
	if fresh.Type != servergame.PacketTypePlayerPauseState {
		t.Fatalf("expected packet type %q, got %q", servergame.PacketTypePlayerPauseState, fresh.Type)
	}
	if fresh.PlayerID != playerID {
		t.Fatalf("expected player id %q, got %q", playerID, fresh.PlayerID)
	}
	if fresh.Paused {
		t.Fatal("expected fresh player pause state to be false")
	}

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypePauseRequest})
	paused, ok := scenario.game.PlayerPauseStatePacket(playerID)
	if !ok {
		t.Fatal("expected pause state packet after pause request")
	}
	if !paused.Paused {
		t.Fatal("expected pause state packet to report paused true")
	}

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypePauseRequest})
	resumed, ok := scenario.game.PlayerPauseStatePacket(playerID)
	if !ok {
		t.Fatal("expected pause state packet after second pause request")
	}
	if resumed.Paused {
		t.Fatal("expected pause state packet to report paused false")
	}
}

func TestPauseRequestToggleClearsVelocityBeforeResume(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	ship := scenario.player(playerID)
	start := ship.Position()
	ship.Velocity = physics.Vector2{X: 25, Y: -10}

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypePauseRequest})
	if ship.Velocity.X != 0 || ship.Velocity.Y != 0 {
		t.Fatalf("expected pause to clear velocity, got (%v, %v)", ship.Velocity.X, ship.Velocity.Y)
	}

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypePauseRequest})
	scenario.step(1.0 / float64(constants.ServerTickRate))

	if ship.X != start.X || ship.Y != start.Y {
		t.Fatalf("expected resumed player not to drift from (%v, %v), got (%v, %v)", start.X, start.Y, ship.X, ship.Y)
	}
}

func TestFreshPlayerAcceptsInputAndMoves(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	start := scenario.playerState(playerID, playerID)

	scenario.send(playerID, servergame.ClientPacket{
		Type:  servergame.PacketTypeInput,
		Input: entities.InputState{Forward: true},
	})
	scenario.step(1.0 / float64(constants.ServerTickRate))

	fresh, ok := scenario.game.PlayerPauseStatePacket(playerID)
	if !ok {
		t.Fatal("expected pause state packet for fresh player")
	}
	if fresh.Paused {
		t.Fatal("expected fresh player not to be paused")
	}
	state := scenario.playerState(playerID, playerID)
	if state.X == start.X && state.Y == start.Y {
		t.Fatalf("expected fresh player to move after input, still at (%v, %v)", state.X, state.Y)
	}
}

func TestFreshPlayerCanShoot(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	scenario.send(playerID, servergame.ClientPacket{
		Type:  servergame.PacketTypeInput,
		Input: entities.InputState{Shoot: true},
	})
	scenario.step(1.0 / float64(constants.ServerTickRate))

	packet := scenario.state(playerID)
	if len(packet.Bullets) != 1 {
		t.Fatalf("expected fresh player to shoot, got %d bullets", len(packet.Bullets))
	}
}

func TestPausedPlayerDoesNotMoveOrShoot(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	start := scenario.playerState(playerID, playerID)

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypePauseRequest})
	scenario.send(playerID, servergame.ClientPacket{
		Type:  servergame.PacketTypeInput,
		Input: entities.InputState{Forward: true, Shoot: true},
	})
	scenario.step(1.0 / float64(constants.ServerTickRate))

	packet := scenario.state(playerID)
	player := packet.Players[playerID]
	if player.X != start.X || player.Y != start.Y {
		t.Fatalf("expected paused player to stay at (%v, %v), got (%v, %v)", start.X, start.Y, player.X, player.Y)
	}
	if len(packet.Bullets) != 0 {
		t.Fatalf("expected paused player not to shoot, got %d bullets", len(packet.Bullets))
	}
}

func TestPauseRequestSecondToggleResumesAndBlocksImmediateShooting(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypePauseRequest})
	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypePauseRequest})

	resumed, ok := scenario.game.PlayerPauseStatePacket(playerID)
	if !ok {
		t.Fatal("expected pause state packet after resume")
	}
	if resumed.Paused {
		t.Fatal("expected player to resume")
	}

	scenario.send(playerID, servergame.ClientPacket{
		Type:  servergame.PacketTypeInput,
		Input: entities.InputState{Shoot: true},
	})
	scenario.step(1.0 / float64(constants.ServerTickRate))

	packet := scenario.state(playerID)
	if len(packet.Bullets) != 0 {
		t.Fatalf("expected immediately resumed player not to shoot, got %d bullets", len(packet.Bullets))
	}
}

func TestPauseRequestToggleIgnoredForDeadPlayer(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	scenario.removePlayerEntity(playerID)

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypePauseRequest})

	if scenario.playerEntityExists(playerID) {
		t.Fatal("expected resume to leave dead player inactive")
	}
}
