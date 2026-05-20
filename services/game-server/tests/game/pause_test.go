package gametests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	servergame "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
)

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

func TestFreshPlayerAcceptsInputAndMoves(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	start := scenario.playerState(playerID, playerID)

	scenario.send(playerID, servergame.ClientPacket{
		Type:  servergame.PacketTypeInput,
		Input: entities.InputState{Forward: true},
	})
	scenario.step(1.0 / float64(constants.ServerTickRate))

	player := scenario.playerState(playerID, playerID)
	if player.Paused {
		t.Fatal("expected fresh player not to be paused")
	}
	if player.X == start.X && player.Y == start.Y {
		t.Fatalf("expected fresh player to move after input, still at (%v, %v)", player.X, player.Y)
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

	player := scenario.playerState(playerID, playerID)
	if player.Paused {
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
