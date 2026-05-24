package gametests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/rules"
)

func TestStatePacketIncludesPlayerLifecycleForAllSessions(t *testing.T) {
	scenario := newScenario(t)
	activePlayerID := scenario.addPlayer()
	pendingPlayerID := scenario.addPlayer()
	eliminatedPlayerID := scenario.addPlayer()

	scenario.removePlayerEntity(pendingPlayerID)
	scenario.setPlayerLives(eliminatedPlayerID, 0)
	scenario.removePlayerEntity(eliminatedPlayerID)

	packet := scenario.state(activePlayerID)

	wantLifecycle := map[string]string{
		activePlayerID:     string(rules.PlayerActive),
		pendingPlayerID:    string(rules.PlayerPendingRespawn),
		eliminatedPlayerID: string(rules.PlayerEliminated),
	}
	if len(packet.PlayerLifecycle) != len(wantLifecycle) {
		t.Fatalf("expected %d lifecycle entries, got %d", len(wantLifecycle), len(packet.PlayerLifecycle))
	}
	for playerID, wantStatus := range wantLifecycle {
		if gotStatus := packet.PlayerLifecycle[playerID]; gotStatus != wantStatus {
			t.Fatalf("expected %q lifecycle %q, got %q", playerID, wantStatus, gotStatus)
		}
	}
	if _, ok := packet.Players[pendingPlayerID]; ok {
		t.Fatalf("expected pending player %q not to have active ship state", pendingPlayerID)
	}
	if _, ok := packet.Players[eliminatedPlayerID]; ok {
		t.Fatalf("expected eliminated player %q not to have active ship state", eliminatedPlayerID)
	}
}
