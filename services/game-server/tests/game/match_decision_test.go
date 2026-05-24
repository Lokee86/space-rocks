package gametests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/rules"
)

func TestGameMatchDecisionReportsPlayerParticipation(t *testing.T) {
	scenario := newScenario(t)
	activePlayerID := scenario.addPlayer()
	pendingPlayerID := scenario.addPlayer()
	eliminatedPlayerID := scenario.addPlayer()

	scenario.removePlayerEntity(pendingPlayerID)
	scenario.setPlayerLives(eliminatedPlayerID, 0)
	scenario.removePlayerEntity(eliminatedPlayerID)

	decision := scenario.game.MatchDecision()
	if decision.IsOver {
		t.Fatal("expected mixed participating match not to be over")
	}

	statuses := decisionByID(decision)
	wantStatuses := map[string]rules.PlayerParticipationStatus{
		activePlayerID:     rules.PlayerActive,
		pendingPlayerID:    rules.PlayerPendingRespawn,
		eliminatedPlayerID: rules.PlayerEliminated,
	}
	if len(statuses) != len(wantStatuses) {
		t.Fatalf("expected %d player decisions, got %d", len(wantStatuses), len(statuses))
	}
	for playerID, wantStatus := range wantStatuses {
		if gotStatus := statuses[playerID]; gotStatus != wantStatus {
			t.Fatalf("expected %q status %q, got %q", playerID, wantStatus, gotStatus)
		}
	}
}

func TestGameMatchDecisionReportsEliminatedMatchOver(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	scenario.setPlayerLives(playerID, 0)
	scenario.removePlayerEntity(playerID)

	decision := scenario.game.MatchDecision()
	if !decision.IsOver {
		t.Fatal("expected eliminated player match to be over")
	}

	statuses := decisionByID(decision)
	if gotStatus := statuses[playerID]; gotStatus != rules.PlayerEliminated {
		t.Fatalf("expected %q status %q, got %q", playerID, rules.PlayerEliminated, gotStatus)
	}
}

func decisionByID(decision rules.MatchDecision) map[string]rules.PlayerParticipationStatus {
	statuses := make(map[string]rules.PlayerParticipationStatus, len(decision.Players))
	for _, player := range decision.Players {
		statuses[player.ID] = player.Status
	}
	return statuses
}
