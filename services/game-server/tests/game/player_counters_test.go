package gametests

import (
	"testing"

	servergame "github.com/Lokee86/space-rocks/server/internal/game"
)

func TestSetPlayerScoreSetsExactValue(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	game.SetPlayerScore(playerID, 42)

	assertPlayerSessionScore(t, game, playerID, 42)
}

func TestSetPlayerScoreClampsNegativeToZero(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	game.SetPlayerScore(playerID, -10)

	assertPlayerSessionScore(t, game, playerID, 0)
}

func TestAddPlayerScoreIncreases(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	game.SetPlayerScore(playerID, 10)
	game.AddPlayerScore(playerID, 5)

	assertPlayerSessionScore(t, game, playerID, 15)
}

func TestAddPlayerScoreCanReduce(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	game.SetPlayerScore(playerID, 10)
	game.AddPlayerScore(playerID, -3)

	assertPlayerSessionScore(t, game, playerID, 7)
}

func TestAddPlayerScoreClampsBelowZero(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	game.SetPlayerScore(playerID, 2)
	game.AddPlayerScore(playerID, -5)

	assertPlayerSessionScore(t, game, playerID, 0)
}

func TestSetPlayerLivesSetsExactValue(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	game.SetPlayerLives(playerID, 5)

	assertPlayerSessionLives(t, game, playerID, 5)
}

func TestSetPlayerLivesClampsNegativeToZero(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	game.SetPlayerLives(playerID, -10)

	assertPlayerSessionLives(t, game, playerID, 0)
}

func TestAddPlayerLivesIncreases(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	game.SetPlayerLives(playerID, 3)
	game.AddPlayerLives(playerID, 2)

	assertPlayerSessionLives(t, game, playerID, 5)
}

func TestAddPlayerLivesCanReduce(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	game.SetPlayerLives(playerID, 3)
	game.AddPlayerLives(playerID, -1)

	assertPlayerSessionLives(t, game, playerID, 2)
}

func TestAddPlayerLivesClampsBelowZero(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	game.SetPlayerLives(playerID, 1)
	game.AddPlayerLives(playerID, -5)

	assertPlayerSessionLives(t, game, playerID, 0)
}

func TestPlayerSessionReportsCounterSeamUpdates(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()
	expectedScore := 77
	expectedLives := 4

	game.SetPlayerScore(playerID, expectedScore)
	game.SetPlayerLives(playerID, expectedLives)

	snapshot := game.GameplayPresentationSnapshot(playerID)
	session, ok := snapshot.PlayerSessions[playerID]
	if !ok {
		t.Fatalf("expected player session %q in gameplay snapshot", playerID)
	}
	if session.Score != expectedScore {
		t.Fatalf("expected player score %d, got %d", expectedScore, session.Score)
	}
	if session.Lives != expectedLives {
		t.Fatalf("expected player lives %d, got %d", expectedLives, session.Lives)
	}
	if snapshot.Lives != expectedLives {
		t.Fatalf("expected snapshot lives %d, got %d", expectedLives, snapshot.Lives)
	}
}

func assertPlayerSessionScore(t *testing.T, game *servergame.Game, playerID string, expected int) {
	t.Helper()

	snapshot := game.GameplayPresentationSnapshot(playerID)
	session, ok := snapshot.PlayerSessions[playerID]
	if !ok {
		t.Fatalf("expected player session %q in gameplay snapshot", playerID)
	}
	if session.Score != expected {
		t.Fatalf("expected player score %d, got %d", expected, session.Score)
	}
}

func assertPlayerSessionLives(t *testing.T, game *servergame.Game, playerID string, expected int) {
	t.Helper()

	snapshot := game.GameplayPresentationSnapshot(playerID)
	session, ok := snapshot.PlayerSessions[playerID]
	if !ok {
		t.Fatalf("expected player session %q in gameplay snapshot", playerID)
	}
	if snapshot.Lives != expected {
		t.Fatalf("expected snapshot lives %d, got %d", expected, snapshot.Lives)
	}
	if session.Lives != expected {
		t.Fatalf("expected player lives %d, got %d", expected, session.Lives)
	}
}