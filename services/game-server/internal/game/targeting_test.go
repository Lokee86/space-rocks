package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	targetpolicy "github.com/Lokee86/space-rocks/server/internal/game/targeting"
)

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

func TestSetPlayerTargetReflectedInStatePacket(t *testing.T) {
	game := New()
	shooterID := game.AddPlayer()
	targetID := game.AddPlayer()

	if !game.SetPlayerTarget(shooterID, targetID) {
		t.Fatal("expected SetPlayerTarget to succeed for existing requester/target")
	}

	packet := game.StatePacket(shooterID)
	shooterState, ok := packet.Players[shooterID]
	if !ok {
		t.Fatalf("expected state packet to include shooter %q", shooterID)
	}
	if shooterState.TargetKind != "player" {
		t.Fatalf("expected target kind %q, got %q", "player", shooterState.TargetKind)
	}
	if shooterState.TargetID != targetID {
		t.Fatalf("expected target id %q, got %q", targetID, shooterState.TargetID)
	}
	if shooterState.TargetPlayerID != targetID {
		t.Fatalf("expected target_player_id %q, got %q", targetID, shooterState.TargetPlayerID)
	}
}

func TestSetTargetStoresPlayerKindAndID(t *testing.T) {
	game := New()
	playerA := game.AddPlayer()
	playerB := game.AddPlayer()

	ok := game.SetTarget(playerA, targetpolicy.TargetRef{
		Kind: targetpolicy.TargetKindPlayer,
		ID:   playerB,
	})
	if !ok {
		t.Fatal("expected SetTarget player target to succeed")
	}

	target := game.Target(playerA)
	if target.Kind != targetpolicy.TargetKindPlayer {
		t.Fatalf("expected target kind %q, got %q", targetpolicy.TargetKindPlayer, target.Kind)
	}
	if target.ID != playerB {
		t.Fatalf("expected target id %q, got %q", playerB, target.ID)
	}
}

func TestSetTargetStoresAsteroidKindAndID(t *testing.T) {
	game := New()
	playerA := game.AddPlayer()
	asteroid := entities.NewAsteroid("asteroid-1", physics.Vector2{}, physics.Vector2{}, 1, 0)
	game.state.Asteroids[asteroid.ID] = asteroid

	ok := game.SetTarget(playerA, targetpolicy.TargetRef{
		Kind: targetpolicy.TargetKindAsteroid,
		ID:   asteroid.ID,
	})
	if !ok {
		t.Fatal("expected SetTarget asteroid target to succeed")
	}

	target := game.Target(playerA)
	if target.Kind != targetpolicy.TargetKindAsteroid {
		t.Fatalf("expected target kind %q, got %q", targetpolicy.TargetKindAsteroid, target.Kind)
	}
	if target.ID != asteroid.ID {
		t.Fatalf("expected target id %q, got %q", asteroid.ID, target.ID)
	}
}

func TestSetTargetStoresBulletKindAndID(t *testing.T) {
	game := New()
	playerA := game.AddPlayer()
	bullet := entities.NewBullet("bullet-1", playerA, physics.Vector2{}, 0, physics.Vector2{}, 1.0)
	game.state.Projectiles[bullet.ID] = bullet

	ok := game.SetTarget(playerA, targetpolicy.TargetRef{
		Kind: targetpolicy.TargetKindBullet,
		ID:   bullet.ID,
	})
	if !ok {
		t.Fatal("expected SetTarget bullet target to succeed")
	}

	target := game.Target(playerA)
	if target.Kind != targetpolicy.TargetKindBullet {
		t.Fatalf("expected target kind %q, got %q", targetpolicy.TargetKindBullet, target.Kind)
	}
	if target.ID != bullet.ID {
		t.Fatalf("expected target id %q, got %q", bullet.ID, target.ID)
	}
}

func TestClearTargetEmptiesGenericAndCompatibilityTargetFields(t *testing.T) {
	game := New()
	playerA := game.AddPlayer()
	playerB := game.AddPlayer()

	if !game.SetTarget(playerA, targetpolicy.TargetRef{
		Kind: targetpolicy.TargetKindPlayer,
		ID:   playerB,
	}) {
		t.Fatal("expected setup SetTarget to succeed")
	}

	game.ClearTarget(playerA)

	target := game.Target(playerA)
	if !target.IsEmpty() {
		t.Fatalf("expected cleared generic target to be empty, got %#v", target)
	}

	player := game.state.Players[playerA]
	if player == nil {
		t.Fatalf("expected player %q to exist", playerA)
	}
	if player.TargetPlayerID != "" {
		t.Fatalf("expected compatibility target_player_id to be empty, got %q", player.TargetPlayerID)
	}
}

func TestSetTargetInvalidTargetDoesNotOverwriteExistingTarget(t *testing.T) {
	game := New()
	playerA := game.AddPlayer()
	playerB := game.AddPlayer()

	if !game.SetTarget(playerA, targetpolicy.TargetRef{
		Kind: targetpolicy.TargetKindPlayer,
		ID:   playerB,
	}) {
		t.Fatal("expected setup SetTarget to succeed")
	}

	ok := game.SetTarget(playerA, targetpolicy.TargetRef{
		Kind: targetpolicy.TargetKindAsteroid,
		ID:   "asteroid-missing",
	})
	if ok {
		t.Fatal("expected SetTarget with missing asteroid target to fail")
	}

	target := game.Target(playerA)
	if target.Kind != targetpolicy.TargetKindPlayer || target.ID != playerB {
		t.Fatalf("expected existing target to remain player %q, got %#v", playerB, target)
	}
}

func TestSelectTargetAtPositionStoresPlayerKindAndIDWhenOverlapping(t *testing.T) {
	game := New()
	requesterID := game.AddPlayer()
	targetID := game.AddPlayer()

	targetPlayer := game.state.Players[targetID]
	if targetPlayer == nil {
		t.Fatalf("expected target player %q to exist", targetID)
	}

	ok := game.SelectTargetAtPosition(
		requesterID,
		targetPlayer.X,
		targetPlayer.Y,
		targetpolicy.TargetRef{Kind: targetpolicy.TargetKindPlayer, ID: targetID},
	)
	if !ok {
		t.Fatal("expected SelectTargetAtPosition player target to succeed")
	}

	target := game.Target(requesterID)
	if target.Kind != targetpolicy.TargetKindPlayer || target.ID != targetID {
		t.Fatalf("expected player target %q, got %#v", targetID, target)
	}
}

func TestSelectTargetAtPositionStoresAsteroidKindAndIDWhenOverlapping(t *testing.T) {
	game := New()
	requesterID := game.AddPlayer()
	asteroid := entities.NewAsteroid("asteroid-claim", physics.Vector2{X: 120, Y: 75}, physics.Vector2{}, 1, 0)
	game.state.Asteroids[asteroid.ID] = asteroid

	ok := game.SelectTargetAtPosition(
		requesterID,
		asteroid.X,
		asteroid.Y,
		targetpolicy.TargetRef{Kind: targetpolicy.TargetKindAsteroid, ID: asteroid.ID},
	)
	if !ok {
		t.Fatal("expected SelectTargetAtPosition asteroid target to succeed")
	}

	target := game.Target(requesterID)
	if target.Kind != targetpolicy.TargetKindAsteroid || target.ID != asteroid.ID {
		t.Fatalf("expected asteroid target %q, got %#v", asteroid.ID, target)
	}
}

func TestSelectTargetAtPositionStoresBulletKindAndIDWhenOverlapping(t *testing.T) {
	game := New()
	requesterID := game.AddPlayer()
	bullet := entities.NewBullet("bullet-claim", requesterID, physics.Vector2{X: 200, Y: 150}, 0, physics.Vector2{}, 1.0)
	game.state.Projectiles[bullet.ID] = bullet

	ok := game.SelectTargetAtPosition(
		requesterID,
		bullet.X,
		bullet.Y,
		targetpolicy.TargetRef{Kind: targetpolicy.TargetKindBullet, ID: bullet.ID},
	)
	if !ok {
		t.Fatal("expected SelectTargetAtPosition bullet target to succeed")
	}

	target := game.Target(requesterID)
	if target.Kind != targetpolicy.TargetKindBullet || target.ID != bullet.ID {
		t.Fatalf("expected bullet target %q, got %#v", bullet.ID, target)
	}
}

func TestSelectTargetAtPositionMissingTargetDoesNotOverwriteExistingTarget(t *testing.T) {
	game := New()
	requesterID := game.AddPlayer()
	existingTargetID := game.AddPlayer()

	if !game.SetTarget(requesterID, targetpolicy.TargetRef{
		Kind: targetpolicy.TargetKindPlayer,
		ID:   existingTargetID,
	}) {
		t.Fatal("expected setup SetTarget to succeed")
	}

	ok := game.SelectTargetAtPosition(
		requesterID,
		0,
		0,
		targetpolicy.TargetRef{Kind: targetpolicy.TargetKindAsteroid, ID: "asteroid-missing"},
	)
	if ok {
		t.Fatal("expected SelectTargetAtPosition to fail for missing target")
	}

	target := game.Target(requesterID)
	if target.Kind != targetpolicy.TargetKindPlayer || target.ID != existingTargetID {
		t.Fatalf("expected existing target to remain player %q, got %#v", existingTargetID, target)
	}
}

func TestSelectTargetAtPositionNonOverlappingPointDoesNotOverwriteExistingTarget(t *testing.T) {
	game := New()
	requesterID := game.AddPlayer()
	existingTargetID := game.AddPlayer()
	newTargetID := game.AddPlayer()

	if !game.SetTarget(requesterID, targetpolicy.TargetRef{
		Kind: targetpolicy.TargetKindPlayer,
		ID:   existingTargetID,
	}) {
		t.Fatal("expected setup SetTarget to succeed")
	}

	newTarget := game.state.Players[newTargetID]
	if newTarget == nil {
		t.Fatalf("expected new target player %q to exist", newTargetID)
	}

	ok := game.SelectTargetAtPosition(
		requesterID,
		newTarget.X+5000,
		newTarget.Y+5000,
		targetpolicy.TargetRef{Kind: targetpolicy.TargetKindPlayer, ID: newTargetID},
	)
	if ok {
		t.Fatal("expected SelectTargetAtPosition to fail for non-overlapping point")
	}

	target := game.Target(requesterID)
	if target.Kind != targetpolicy.TargetKindPlayer || target.ID != existingTargetID {
		t.Fatalf("expected existing target to remain player %q, got %#v", existingTargetID, target)
	}
}
