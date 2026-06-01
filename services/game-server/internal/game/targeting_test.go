package game

import "testing"

func TestSetPlayerTargetStoresExistingTarget(t *testing.T) {
	game := New()
	playerA := game.AddPlayer()
	playerB := game.AddPlayer()

	ok := game.SetPlayerTarget(playerA, playerB)
	if !ok {
		t.Fatal("expected SetPlayerTarget to succeed for existing requester/target")
	}
	if got := game.PlayerTarget(playerA); got != playerB {
		t.Fatalf("expected stored target %q, got %q", playerB, got)
	}
}

func TestClearPlayerTargetStoresEmptyString(t *testing.T) {
	game := New()
	playerA := game.AddPlayer()
	playerB := game.AddPlayer()

	if !game.SetPlayerTarget(playerA, playerB) {
		t.Fatal("expected setup target set to succeed")
	}
	if !game.ClearPlayerTarget(playerA) {
		t.Fatal("expected ClearPlayerTarget to succeed")
	}
	if got := game.PlayerTarget(playerA); got != "" {
		t.Fatalf("expected cleared target to be empty, got %q", got)
	}
}

func TestSetPlayerTargetMissingTargetDoesNotOverwriteExistingTarget(t *testing.T) {
	game := New()
	playerA := game.AddPlayer()
	playerB := game.AddPlayer()

	if !game.SetPlayerTarget(playerA, playerB) {
		t.Fatal("expected setup target set to succeed")
	}

	ok := game.SetPlayerTarget(playerA, "player-missing")
	if ok {
		t.Fatal("expected SetPlayerTarget to fail for missing target")
	}
	if got := game.PlayerTarget(playerA); got != playerB {
		t.Fatalf("expected existing target %q to remain unchanged, got %q", playerB, got)
	}
}

func TestStatePacketIncludesTargetPlayerID(t *testing.T) {
	game := New()
	playerA := game.AddPlayer()
	playerB := game.AddPlayer()

	if !game.SetPlayerTarget(playerA, playerB) {
		t.Fatal("expected SetPlayerTarget to succeed for existing requester/target")
	}

	packet := game.StatePacket(playerA)
	playerState, ok := packet.Players[playerA]
	if !ok {
		t.Fatalf("expected state packet to include player %q", playerA)
	}
	if playerState.TargetPlayerID != playerB {
		t.Fatalf("expected target_player_id %q, got %q", playerB, playerState.TargetPlayerID)
	}
}
