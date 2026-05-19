package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestPausePlayerPacketClearsInputAndIgnoresNewInput(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	player := game.state.Players[playerID]
	player.SetInput(entities.InputState{Forward: true, Shoot: true})

	game.HandlePacket(playerID, ClientPacket{Type: PacketTypePausePlayer})

	if !player.Paused {
		t.Fatal("expected player to be paused")
	}
	if player.Input.Forward || player.Input.Shoot {
		t.Fatal("expected pause to clear player input")
	}
	if !game.statePacket(playerID).Players[playerID].Paused {
		t.Fatal("expected state packet to report paused player")
	}

	game.HandlePacket(playerID, ClientPacket{
		Type:  PacketTypeInput,
		Input: entities.InputState{Forward: true, Shoot: true},
	})

	if player.Input.Forward || player.Input.Shoot {
		t.Fatal("expected input packet to be ignored while paused")
	}
}

func TestFreshPlayerAcceptsInputAndMoves(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	player := game.state.Players[playerID]
	startX := player.X
	startY := player.Y

	game.HandlePacket(playerID, ClientPacket{
		Type:  PacketTypeInput,
		Input: entities.InputState{Forward: true},
	})
	game.Step(1.0 / float64(constants.ServerTickRate))

	if player.Paused {
		t.Fatal("expected fresh player not to be paused")
	}
	if player.X == startX && player.Y == startY {
		t.Fatalf("expected fresh player to move after input, still at (%v, %v)", player.X, player.Y)
	}
}

func TestFreshPlayerCanShoot(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()

	game.HandlePacket(playerID, ClientPacket{
		Type:  PacketTypeInput,
		Input: entities.InputState{Shoot: true},
	})
	game.Step(1.0 / float64(constants.ServerTickRate))

	if len(game.state.Projectiles) != 1 {
		t.Fatalf("expected fresh player to shoot, got %d projectiles", len(game.state.Projectiles))
	}
}

func TestPausedPlayerDoesNotMoveOrShoot(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	player := game.state.Players[playerID]
	player.Paused = true
	player.Velocity = physics.Vector2{X: 100, Y: 0}
	player.SetInput(entities.InputState{Forward: true, Shoot: true})
	startX := player.X
	startY := player.Y

	game.Step(1.0 / float64(constants.ServerTickRate))

	if player.X != startX || player.Y != startY {
		t.Fatalf("expected paused player to stay at (%v, %v), got (%v, %v)", startX, startY, player.X, player.Y)
	}
	if len(game.state.Projectiles) != 0 {
		t.Fatalf("expected paused player not to shoot, got %d projectiles", len(game.state.Projectiles))
	}
}

func TestResumePlayerPacketStartsInvulnerabilityAndBlocksShooting(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	player := game.state.Players[playerID]
	player.Paused = true

	game.HandlePacket(playerID, ClientPacket{Type: PacketTypeResumePlayer})

	if player.Paused {
		t.Fatal("expected player to resume")
	}
	if player.InvulnerabilityRemaining != constants.PlayerResumeInvulnerabilitySeconds {
		t.Fatalf("expected invulnerability %v, got %v", constants.PlayerResumeInvulnerabilitySeconds, player.InvulnerabilityRemaining)
	}

	game.HandlePacket(playerID, ClientPacket{
		Type:  PacketTypeInput,
		Input: entities.InputState{Shoot: true},
	})
	game.Step(1.0 / float64(constants.ServerTickRate))

	if len(game.state.Projectiles) != 0 {
		t.Fatalf("expected invulnerable player not to shoot, got %d projectiles", len(game.state.Projectiles))
	}
}

func TestResumePacketIgnoredForDeadPlayer(t *testing.T) {
	game := New()
	playerID := "player-1"
	game.playerSessions[playerID] = newPlayerSession(playerID, physics.Vector2{})

	game.HandlePacket(playerID, ClientPacket{Type: PacketTypeResumePlayer})

	if _, ok := game.state.Players[playerID]; ok {
		t.Fatal("expected resume to leave dead player inactive")
	}
}
