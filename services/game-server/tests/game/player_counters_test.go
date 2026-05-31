package gametests

import (
	"testing"

	servergame "github.com/Lokee86/space-rocks/server/internal/game"
)

func TestSetPlayerScoreSetsExactValue(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	game.SetPlayerScore(playerID, 42)

	assertPlayerPacketScore(t, game, playerID, 42)
}

func TestSetPlayerScoreClampsNegativeToZero(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	game.SetPlayerScore(playerID, -10)

	assertPlayerPacketScore(t, game, playerID, 0)
}

func TestAddPlayerScoreIncreases(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	game.SetPlayerScore(playerID, 10)
	game.AddPlayerScore(playerID, 5)

	assertPlayerPacketScore(t, game, playerID, 15)
}

func TestAddPlayerScoreCanReduce(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	game.SetPlayerScore(playerID, 10)
	game.AddPlayerScore(playerID, -3)

	assertPlayerPacketScore(t, game, playerID, 7)
}

func TestAddPlayerScoreClampsBelowZero(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	game.SetPlayerScore(playerID, 2)
	game.AddPlayerScore(playerID, -5)

	assertPlayerPacketScore(t, game, playerID, 0)
}

func TestSetPlayerLivesSetsExactValue(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	game.SetPlayerLives(playerID, 5)

	assertPlayerPacketLives(t, game, playerID, 5)
}

func TestSetPlayerLivesClampsNegativeToZero(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	game.SetPlayerLives(playerID, -10)

	assertPlayerPacketLives(t, game, playerID, 0)
}

func TestAddPlayerLivesIncreases(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	game.SetPlayerLives(playerID, 3)
	game.AddPlayerLives(playerID, 2)

	assertPlayerPacketLives(t, game, playerID, 5)
}

func TestAddPlayerLivesCanReduce(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	game.SetPlayerLives(playerID, 3)
	game.AddPlayerLives(playerID, -1)

	assertPlayerPacketLives(t, game, playerID, 2)
}

func TestAddPlayerLivesClampsBelowZero(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	game.SetPlayerLives(playerID, 1)
	game.AddPlayerLives(playerID, -5)

	assertPlayerPacketLives(t, game, playerID, 0)
}

func TestStatePacketReportsCounterSeamUpdates(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()
	expectedScore := 77
	expectedLives := 4

	game.SetPlayerScore(playerID, expectedScore)
	game.SetPlayerLives(playerID, expectedLives)

	packet := game.StatePacket(playerID)
	player, ok := packet.Players[playerID]
	if !ok {
		t.Fatalf("expected player %q in state packet", playerID)
	}
	if player.Score != expectedScore {
		t.Fatalf("expected player score %d, got %d", expectedScore, player.Score)
	}
	if player.Lives != expectedLives {
		t.Fatalf("expected player lives %d, got %d", expectedLives, player.Lives)
	}
	if packet.Lives != expectedLives {
		t.Fatalf("expected packet lives %d, got %d", expectedLives, packet.Lives)
	}
}

func assertPlayerPacketScore(t *testing.T, game *servergame.Game, playerID string, expected int) {
	t.Helper()

	packet := game.StatePacket(playerID)
	player, ok := packet.Players[playerID]
	if !ok {
		t.Fatalf("expected player %q in state packet", playerID)
	}
	if player.Score != expected {
		t.Fatalf("expected player score %d, got %d", expected, player.Score)
	}
}

func assertPlayerPacketLives(t *testing.T, game *servergame.Game, playerID string, expected int) {
	t.Helper()

	packet := game.StatePacket(playerID)
	player, ok := packet.Players[playerID]
	if !ok {
		t.Fatalf("expected player %q in state packet", playerID)
	}
	if packet.Lives != expected {
		t.Fatalf("expected packet lives %d, got %d", expected, packet.Lives)
	}
	if player.Lives != expected {
		t.Fatalf("expected player lives %d, got %d", expected, player.Lives)
	}
}
