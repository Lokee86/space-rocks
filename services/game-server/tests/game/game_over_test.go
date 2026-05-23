package gametests

import "testing"

func TestGameIsGameOverReturnsFalseWithoutPlayerSessions(t *testing.T) {
	scenario := newScenario(t)

	if scenario.game.IsGameOver() {
		t.Fatal("expected game without player sessions not to be game over")
	}
}

func TestGameIsGameOverReturnsFalseWhenAnySessionHasLives(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	scenario.removePlayerEntity(playerID)

	if scenario.game.IsGameOver() {
		t.Fatal("expected game with remaining player lives not to be game over")
	}
}

func TestGameIsGameOverReturnsFalseWhenActivePlayersRemain(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	scenario.setPlayerLives(playerID, 0)

	if scenario.game.IsGameOver() {
		t.Fatal("expected game with active player ships not to be game over")
	}
}

func TestGameIsGameOverReturnsTrueWhenAllSessionsOutOfLivesAndNoActivePlayers(t *testing.T) {
	scenario := newScenario(t)
	firstPlayerID := scenario.addPlayer()
	secondPlayerID := scenario.addPlayer()
	scenario.setPlayerLives(firstPlayerID, 0)
	scenario.setPlayerLives(secondPlayerID, 0)
	scenario.removePlayerEntity(firstPlayerID)
	scenario.removePlayerEntity(secondPlayerID)

	if !scenario.game.IsGameOver() {
		t.Fatal("expected game with all sessions out of lives and no active player ships to be game over")
	}
}
