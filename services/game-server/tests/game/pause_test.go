package gametests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	servergame "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestShipIsSuspendedReflectsPauseAndFreeze(t *testing.T) {
	ship := entities.Ship{}
	if ship.IsSuspended() {
		t.Fatal("expected ship without pause or freeze to be active")
	}

	ship.Pause()
	if !ship.IsSuspended() {
		t.Fatal("expected paused ship to be suspended")
	}

	ship.Resume(0)
	ship.Suspension.SetDevFrozen(true)
	if !ship.IsSuspended() {
		t.Fatal("expected frozen ship to be suspended")
	}
}

func TestFrozenShipCannotReceiveInputOrMove(t *testing.T) {
	ship := entities.Ship{}
	if !ship.CanReceiveInput() {
		t.Fatal("expected active ship to receive input")
	}
	if !ship.CanMove() {
		t.Fatal("expected active ship to move")
	}

	ship.Suspension.SetDevFrozen(true)
	if ship.CanReceiveInput() {
		t.Fatal("expected frozen ship not to receive input")
	}
	if ship.CanMove() {
		t.Fatal("expected frozen ship not to move")
	}
}

func TestPausedAndFrozenShipRequiresBothCausesCleared(t *testing.T) {
	ship := entities.Ship{}
	ship.Pause()
	ship.Suspension.SetDevFrozen(true)

	ship.Resume(0)
	if !ship.Suspension.DevFrozen {
		t.Fatal("expected resume not to clear player freeze")
	}
	if !ship.IsSuspended() {
		t.Fatal("expected resumed ship to remain suspended while frozen")
	}

	ship.Pause()
	ship.Suspension.SetDevFrozen(false)
	if !ship.Suspension.Paused {
		t.Fatal("expected unfreeze not to clear pause")
	}
	if !ship.IsSuspended() {
		t.Fatal("expected unfrozen ship to remain suspended while paused")
	}

	ship.Resume(0)
	if ship.IsSuspended() {
		t.Fatal("expected ship to be active after pause and freeze are cleared")
	}
}

func TestPausePlayerPacketClearsInputAndIgnoresNewInput(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	start := scenario.playerState(playerID, playerID)

	scenario.send(playerID, servergame.ClientPacket{
		Type:  servergame.PacketTypeInput,
		Input: entities.InputState{Forward: true, Shoot: true},
	})
	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypePausePlayer})

	paused := scenario.playerState(playerID, playerID)
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

	paused := scenario.playerState(playerID, playerID)
	if !paused.Paused {
		t.Fatal("expected pause request to pause player")
	}

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypePauseRequest})

	player := scenario.player(playerID)
	if player.Suspension.Paused {
		t.Fatal("expected second pause request to resume player")
	}
}

func TestPausePlayerPacketClearsVelocityBeforeResume(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	ship := scenario.player(playerID)
	start := ship.Position()
	ship.Velocity = physics.Vector2{X: 25, Y: -10}

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypePausePlayer})
	if ship.Velocity.X != 0 || ship.Velocity.Y != 0 {
		t.Fatalf("expected pause to clear velocity, got (%v, %v)", ship.Velocity.X, ship.Velocity.Y)
	}

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeResumePlayer})
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

	player := scenario.player(playerID)
	if player.Suspension.Paused {
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

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypePausePlayer})
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

func TestResumePlayerPacketResumesAndBlocksImmediateShooting(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypePausePlayer})
	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeResumePlayer})

	player := scenario.player(playerID)
	if player.Suspension.Paused {
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

func TestResumePacketIgnoredForDeadPlayer(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	scenario.removePlayerEntity(playerID)

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeResumePlayer})

	if scenario.playerEntityExists(playerID) {
		t.Fatal("expected resume to leave dead player inactive")
	}
}
