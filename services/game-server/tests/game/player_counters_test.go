package gametests

import (
	"testing"

	servergame "github.com/Lokee86/space-rocks/services/game-server/internal/game"
)

func TestSetPlayerScoreSetsExactValue(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	change := game.SetPlayerScore(playerID, 42)
	assertPlayerCounterChange(t, change, 42)
}

func TestSetPlayerScoreClampsNegativeToZero(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	change := game.SetPlayerScore(playerID, -10)
	assertPlayerCounterChange(t, change, 0)
}

func TestAddPlayerScoreIncreases(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	setChange := game.SetPlayerScore(playerID, 10)
	assertPlayerCounterChange(t, setChange, 10)
	addChange := game.AddPlayerScore(playerID, 5)
	assertPlayerCounterChange(t, addChange, 15)
}

func TestAddPlayerScoreCanReduce(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	setChange := game.SetPlayerScore(playerID, 10)
	assertPlayerCounterChange(t, setChange, 10)
	addChange := game.AddPlayerScore(playerID, -3)
	assertPlayerCounterChange(t, addChange, 7)
}

func TestAddPlayerScoreClampsBelowZero(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	setChange := game.SetPlayerScore(playerID, 2)
	assertPlayerCounterChange(t, setChange, 2)
	addChange := game.AddPlayerScore(playerID, -5)
	assertPlayerCounterChange(t, addChange, 0)
}

func TestSetPlayerLivesSetsExactValue(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	change := game.SetPlayerLives(playerID, 5)
	assertPlayerCounterChange(t, change, 5)
}

func TestSetPlayerLivesClampsNegativeToZero(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	change := game.SetPlayerLives(playerID, -10)
	assertPlayerCounterChange(t, change, 0)
}

func TestAddPlayerLivesIncreases(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	setChange := game.SetPlayerLives(playerID, 3)
	assertPlayerCounterChange(t, setChange, 3)
	addChange := game.AddPlayerLives(playerID, 2)
	assertPlayerCounterChange(t, addChange, 5)
}

func TestAddPlayerLivesCanReduce(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	setChange := game.SetPlayerLives(playerID, 3)
	assertPlayerCounterChange(t, setChange, 3)
	addChange := game.AddPlayerLives(playerID, -1)
	assertPlayerCounterChange(t, addChange, 2)
}

func TestAddPlayerLivesClampsBelowZero(t *testing.T) {
	game := servergame.New()
	playerID := game.AddPlayer()

	setChange := game.SetPlayerLives(playerID, 1)
	assertPlayerCounterChange(t, setChange, 1)
	addChange := game.AddPlayerLives(playerID, -5)
	assertPlayerCounterChange(t, addChange, 0)
}

func TestPlayerSessionReportsCounterSeamUpdates(t *testing.T) {
	game := servergame.New()
	control := servergame.NewControl(game)
	playerID := game.AddPlayer()
	expectedScore := 77
	expectedLives := 4

	if !control.SetPlayerScore(playerID, expectedScore) {
		t.Fatalf("expected SetPlayerScore to find player %q", playerID)
	}
	if !control.SetPlayerLives(playerID, expectedLives) {
		t.Fatalf("expected SetPlayerLives to find player %q", playerID)
	}

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

func assertPlayerCounterChange(t *testing.T, change servergame.PlayerCounterChange, expectedAfter int) {
	t.Helper()

	if !change.Found {
		t.Fatalf("expected player counter change to find player")
	}
	if change.After != expectedAfter {
		t.Fatalf("expected player counter after %d, got %d", expectedAfter, change.After)
	}
}
