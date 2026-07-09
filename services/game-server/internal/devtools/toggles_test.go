package devtools

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
)

func TestHandleToggleDebugInvincibleTargetsAllPlayers(t *testing.T) {
	target := game.New()
	control := game.NewControl(target)
	callerID := target.AddPlayer()
	otherPlayerID := target.AddPlayer()

	ok := HandleCommand(control, callerID, DebugCommand{
		Type:        PacketTypeToggleDebugInvincible,
		TargetScope: targetScopeAllPlayers,
	})
	if !ok {
		t.Fatal("expected HandleCommand to return true")
	}

	assertPlayerInvincible(t, target, callerID, true)
	assertPlayerInvincible(t, target, otherPlayerID, true)
}

func TestHandleToggleDebugInvincibleTargetsAllPlayersKeepsEveryoneEnabledUntilAllAreEnabled(t *testing.T) {
	target := game.New()
	control := game.NewControl(target)
	callerID := target.AddPlayer()
	otherPlayerID := target.AddPlayer()
	control.SetPlayerInvincible(callerID, true)

	ok := HandleCommand(control, callerID, DebugCommand{
		Type:        PacketTypeToggleDebugInvincible,
		TargetScope: targetScopeAllPlayers,
	})
	if !ok {
		t.Fatal("expected HandleCommand to return true")
	}

	assertPlayerInvincible(t, target, callerID, true)
	assertPlayerInvincible(t, target, otherPlayerID, true)
}

func TestHandleToggleDebugInvincibleTargetsAllPlayersDisablesEveryoneWhenAllAreEnabled(t *testing.T) {
	target := game.New()
	control := game.NewControl(target)
	callerID := target.AddPlayer()
	otherPlayerID := target.AddPlayer()
	control.SetPlayerInvincible(callerID, true)
	control.SetPlayerInvincible(otherPlayerID, true)

	ok := HandleCommand(control, callerID, DebugCommand{
		Type:        PacketTypeToggleDebugInvincible,
		TargetScope: targetScopeAllPlayers,
	})
	if !ok {
		t.Fatal("expected HandleCommand to return true")
	}

	assertPlayerInvincible(t, target, callerID, false)
	assertPlayerInvincible(t, target, otherPlayerID, false)
}

func TestHandleToggleDebugInfiniteLivesTargetsAllPlayers(t *testing.T) {
	target := game.New()
	control := game.NewControl(target)
	callerID := target.AddPlayer()
	otherPlayerID := target.AddPlayer()

	ok := HandleCommand(control, callerID, DebugCommand{
		Type:        PacketTypeToggleDebugInfiniteLives,
		TargetScope: targetScopeAllPlayers,
	})
	if !ok {
		t.Fatal("expected HandleCommand to return true")
	}

	assertPlayerInfiniteLives(t, target, callerID, true)
	assertPlayerInfiniteLives(t, target, otherPlayerID, true)
}

func TestHandleToggleDebugInfiniteLivesTargetsAllPlayersKeepsEveryoneEnabledUntilAllAreEnabled(t *testing.T) {
	target := game.New()
	control := game.NewControl(target)
	callerID := target.AddPlayer()
	otherPlayerID := target.AddPlayer()
	control.SetInfiniteLives(callerID, true)

	ok := HandleCommand(control, callerID, DebugCommand{
		Type:        PacketTypeToggleDebugInfiniteLives,
		TargetScope: targetScopeAllPlayers,
	})
	if !ok {
		t.Fatal("expected HandleCommand to return true")
	}

	assertPlayerInfiniteLives(t, target, callerID, true)
	assertPlayerInfiniteLives(t, target, otherPlayerID, true)
}

func TestHandleToggleDebugInfiniteLivesTargetsAllPlayersDisablesEveryoneWhenAllAreEnabled(t *testing.T) {
	target := game.New()
	control := game.NewControl(target)
	callerID := target.AddPlayer()
	otherPlayerID := target.AddPlayer()
	control.SetInfiniteLives(callerID, true)
	control.SetInfiniteLives(otherPlayerID, true)

	ok := HandleCommand(control, callerID, DebugCommand{
		Type:        PacketTypeToggleDebugInfiniteLives,
		TargetScope: targetScopeAllPlayers,
	})
	if !ok {
		t.Fatal("expected HandleCommand to return true")
	}

	assertPlayerInfiniteLives(t, target, callerID, false)
	assertPlayerInfiniteLives(t, target, otherPlayerID, false)
}

func TestHandleToggleDebugFreezePlayerTargetsAllPlayers(t *testing.T) {
	target := game.New()
	control := game.NewControl(target)
	callerID := target.AddPlayer()
	otherPlayerID := target.AddPlayer()

	ok := HandleCommand(control, callerID, DebugCommand{
		Type:        PacketTypeToggleDebugFreezePlayer,
		TargetScope: targetScopeAllPlayers,
	})
	if !ok {
		t.Fatal("expected HandleCommand to return true")
	}

	assertPlayerFrozen(t, target, callerID, true)
	assertPlayerFrozen(t, target, otherPlayerID, true)
}

func TestHandleToggleDebugFreezePlayerTargetsAllPlayersKeepsEveryoneEnabledUntilAllAreEnabled(t *testing.T) {
	target := game.New()
	control := game.NewControl(target)
	callerID := target.AddPlayer()
	otherPlayerID := target.AddPlayer()
	control.SetPlayerFrozen(callerID, true)

	ok := HandleCommand(control, callerID, DebugCommand{
		Type:        PacketTypeToggleDebugFreezePlayer,
		TargetScope: targetScopeAllPlayers,
	})
	if !ok {
		t.Fatal("expected HandleCommand to return true")
	}

	assertPlayerFrozen(t, target, callerID, true)
	assertPlayerFrozen(t, target, otherPlayerID, true)
}

func TestHandleToggleDebugFreezePlayerTargetsAllPlayersDisablesEveryoneWhenAllAreEnabled(t *testing.T) {
	target := game.New()
	control := game.NewControl(target)
	callerID := target.AddPlayer()
	otherPlayerID := target.AddPlayer()
	control.SetPlayerFrozen(callerID, true)
	control.SetPlayerFrozen(otherPlayerID, true)

	ok := HandleCommand(control, callerID, DebugCommand{
		Type:        PacketTypeToggleDebugFreezePlayer,
		TargetScope: targetScopeAllPlayers,
	})
	if !ok {
		t.Fatal("expected HandleCommand to return true")
	}

	assertPlayerFrozen(t, target, callerID, false)
	assertPlayerFrozen(t, target, otherPlayerID, false)
}

func TestHandleDebugKillPlayerTargetsAllPlayers(t *testing.T) {
	target := game.New()
	control := game.NewControl(target)
	callerID := target.AddPlayer()
	otherPlayerID := target.AddPlayer()

	ok := HandleCommand(control, callerID, DebugCommand{
		Type:        PacketTypeDebugKillPlayer,
		TargetScope: targetScopeAllPlayers,
	})
	if !ok {
		t.Fatal("expected HandleCommand to return true")
	}

	assertPlayerInactive(t, target, callerID)
	assertPlayerInactive(t, target, otherPlayerID)
}

func assertPlayerInvincible(t *testing.T, target *game.Game, playerID string, expected bool) {
	t.Helper()

	control := game.NewControl(target)
	actual, found := control.PlayerInvincible(playerID)
	if !found {
		t.Fatalf("expected player %q to be found", playerID)
	}
	if actual != expected {
		t.Fatalf("expected player %q invincible=%v, got %v", playerID, expected, actual)
	}
}

func assertPlayerInfiniteLives(t *testing.T, target *game.Game, playerID string, expected bool) {
	t.Helper()

	control := game.NewControl(target)
	actual, found := control.InfiniteLives(playerID)
	if !found {
		t.Fatalf("expected player %q to be found", playerID)
	}
	if actual != expected {
		t.Fatalf("expected player %q infinite_lives=%v, got %v", playerID, expected, actual)
	}
}

func assertPlayerFrozen(t *testing.T, target *game.Game, playerID string, expected bool) {
	t.Helper()

	control := game.NewControl(target)
	actual, found := control.PlayerFrozen(playerID)
	if !found {
		t.Fatalf("expected player %q to be found", playerID)
	}
	if actual != expected {
		t.Fatalf("expected player %q frozen=%v, got %v", playerID, expected, actual)
	}
}

func assertPlayerInactive(t *testing.T, target *game.Game, playerID string) {
	t.Helper()

	snapshot := target.GameplayPresentationSnapshot(playerID)
	status, ok := snapshot.PlayerLifecycle[playerID]
	if !ok {
		t.Fatalf("expected lifecycle for player %q", playerID)
	}
	if status == "active" {
		t.Fatalf("expected player %q to be inactive after debug kill", playerID)
	}
}
