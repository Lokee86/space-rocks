package gametests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/rules"
)

func TestGameplayPresentationSnapshotIncludesPlayerLifecycleForAllPlayers(t *testing.T) {
	scenario := newScenario(t)
	activePlayerID := scenario.addPlayer()
	pendingPlayerID := scenario.addPlayer()
	eliminatedPlayerID := scenario.addPlayer()

	scenario.removePlayerEntity(pendingPlayerID)
	scenario.setPlayerLives(eliminatedPlayerID, 0)
	scenario.removePlayerEntity(eliminatedPlayerID)

	snapshot := scenario.presentationSnapshot(activePlayerID)

	wantLifecycle := map[string]string{
		activePlayerID:     string(rules.PlayerActive),
		pendingPlayerID:    string(rules.PlayerPendingRespawn),
		eliminatedPlayerID: string(rules.PlayerEliminated),
	}
	if len(snapshot.PlayerLifecycle) != len(wantLifecycle) {
		t.Fatalf("expected %d lifecycle entries, got %d", len(wantLifecycle), len(snapshot.PlayerLifecycle))
	}
	for playerID, wantStatus := range wantLifecycle {
		if gotStatus := snapshot.PlayerLifecycle[playerID]; gotStatus != wantStatus {
			t.Fatalf("expected %q lifecycle %q, got %q", playerID, wantStatus, gotStatus)
		}
	}
	if _, ok := snapshot.Players[pendingPlayerID]; ok {
		t.Fatalf("expected pending player %q not to have active world ship state", pendingPlayerID)
	}
	if _, ok := snapshot.Players[eliminatedPlayerID]; ok {
		t.Fatalf("expected eliminated player %q not to have active world ship state", eliminatedPlayerID)
	}
}