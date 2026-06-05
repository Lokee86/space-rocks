package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/player"
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

func TestStatePacketIncludesTargetKindAndTargetID(t *testing.T) {
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
	if playerState.TargetKind != "player" {
		t.Fatalf("expected target kind %q, got %q", "player", playerState.TargetKind)
	}
	if playerState.TargetID != playerB {
		t.Fatalf("expected target id %q, got %q", playerB, playerState.TargetID)
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
	asteroid := runtime.NewAsteroid("asteroid-1", physics.Vector2{}, physics.Vector2{}, 1, 0)
	game.entities.Asteroids[asteroid.ID] = asteroid

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
	bullet := runtime.NewBullet("bullet-1", playerA, physics.Vector2{}, 0, physics.Vector2{}, 1.0)
	game.entities.Projectiles[bullet.ID] = bullet

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

func TestClearTargetEmptiesGenericTargetFields(t *testing.T) {
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

	player := game.entities.Players[playerA]
	if player == nil {
		t.Fatalf("expected player %q to exist", playerA)
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

	targetPlayer := game.entities.Players[targetID]
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
	asteroid := runtime.NewAsteroid("asteroid-claim", physics.Vector2{X: 120, Y: 75}, physics.Vector2{}, 1, 0)
	game.entities.Asteroids[asteroid.ID] = asteroid

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
	bullet := runtime.NewBullet("bullet-claim", requesterID, physics.Vector2{X: 200, Y: 150}, 0, physics.Vector2{}, 1.0)
	game.entities.Projectiles[bullet.ID] = bullet

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

	newTarget := game.entities.Players[newTargetID]
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

func TestTargetLookupStatusLocked_PlayerActive(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()

	game.mu.Lock()
	status := game.targetLookupStatusLocked(targetpolicy.TargetRef{
		Kind: targetpolicy.TargetKindPlayer,
		ID:   playerID,
	})
	game.mu.Unlock()

	if status != player.TargetStatusActive {
		t.Fatalf("expected status %q, got %q", player.TargetStatusActive, status)
	}
}

func TestTargetLookupStatusLocked_PlayerPendingRespawn(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()

	game.mu.Lock()
	delete(game.entities.Players, playerID)
	status := game.targetLookupStatusLocked(targetpolicy.TargetRef{
		Kind: targetpolicy.TargetKindPlayer,
		ID:   playerID,
	})
	game.mu.Unlock()

	if status != player.TargetStatusInactive {
		t.Fatalf("expected status %q, got %q", player.TargetStatusInactive, status)
	}
}

func TestTargetLookupStatusLocked_PlayerMissing(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()

	game.RemovePlayer(playerID)

	game.mu.Lock()
	status := game.targetLookupStatusLocked(targetpolicy.TargetRef{
		Kind: targetpolicy.TargetKindPlayer,
		ID:   playerID,
	})
	game.mu.Unlock()

	if status != player.TargetStatusMissing {
		t.Fatalf("expected status %q, got %q", player.TargetStatusMissing, status)
	}
}

func TestClearTargetsForMissingPlayersLocked_KeepsInactivePendingRespawnTarget(t *testing.T) {
	game := New()
	requesterID := game.AddPlayer()
	targetID := game.AddPlayer()

	if !game.SetPlayerTarget(requesterID, targetID) {
		t.Fatal("expected setup SetPlayerTarget to succeed")
	}

	game.mu.Lock()
	delete(game.entities.Players, targetID)
	game.clearTargetsForMissingPlayersLocked()
	requester := game.entities.Players[requesterID]
	game.mu.Unlock()

	if requester == nil {
		t.Fatalf("expected requester %q to exist", requesterID)
	}
	if requester.TargetKind != string(targetpolicy.TargetKindPlayer) {
		t.Fatalf("expected target kind to remain %q, got %q", targetpolicy.TargetKindPlayer, requester.TargetKind)
	}
	if requester.TargetID != targetID {
		t.Fatalf("expected target id to remain %q, got %q", targetID, requester.TargetID)
	}
}

func TestClearTargetsForMissingPlayersLocked_ClearsMissingRemovedTarget(t *testing.T) {
	game := New()
	requesterID := game.AddPlayer()
	targetID := game.AddPlayer()

	if !game.SetPlayerTarget(requesterID, targetID) {
		t.Fatal("expected setup SetPlayerTarget to succeed")
	}

	game.RemovePlayer(targetID)

	game.mu.Lock()
	game.clearTargetsForMissingPlayersLocked()
	requester := game.entities.Players[requesterID]
	game.mu.Unlock()

	if requester == nil {
		t.Fatalf("expected requester %q to exist", requesterID)
	}
	if requester.TargetKind != "" || requester.TargetID != "" {
		t.Fatalf("expected missing target to be cleared, got kind=%q id=%q", requester.TargetKind, requester.TargetID)
	}
}

func TestSelectTargetAtPosition_DeadPlayerStoresTargetingAndRespawnMirrorsTarget(t *testing.T) {
	game := New()
	requesterID := game.AddPlayer()
	targetID := game.AddPlayer()

	targetPlayer := game.entities.Players[targetID]
	if targetPlayer == nil {
		t.Fatalf("expected target player %q to exist", targetID)
	}

	if !game.SelectTargetAtPosition(
		requesterID,
		targetPlayer.X,
		targetPlayer.Y,
		targetpolicy.TargetRef{Kind: targetpolicy.TargetKindPlayer, ID: targetID},
	) {
		t.Fatal("expected initial SelectTargetAtPosition to succeed")
	}

	game.mu.Lock()
	delete(game.entities.Players, requesterID)
	game.mu.Unlock()

	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("expected SelectTargetAtPosition while dead not to panic, got %v", recovered)
			}
		}()

		ok := game.SelectTargetAtPosition(
			requesterID,
			targetPlayer.X,
			targetPlayer.Y,
			targetpolicy.TargetRef{Kind: targetpolicy.TargetKindPlayer, ID: targetID},
		)
		if !ok {
			t.Fatal("expected SelectTargetAtPosition while dead to succeed")
		}
	}()

	game.mu.Lock()
	session := game.playerSessions[requesterID]
	game.mu.Unlock()
	if session == nil {
		t.Fatalf("expected player session %q to exist", requesterID)
	}
	if session.Targeting.Kind != string(targetpolicy.TargetKindPlayer) || session.Targeting.ID != targetID {
		t.Fatalf("expected session targeting player %q, got kind=%q id=%q", targetID, session.Targeting.Kind, session.Targeting.ID)
	}

	game.respawnPlayer(requesterID)
	ship := game.entities.Players[requesterID]
	if ship == nil {
		t.Fatalf("expected respawned ship %q to exist", requesterID)
	}
	if ship.TargetKind != string(targetpolicy.TargetKindPlayer) || ship.TargetID != targetID {
		t.Fatalf("expected respawned ship target player %q, got kind=%q id=%q", targetID, ship.TargetKind, ship.TargetID)
	}
}

func TestClearTarget_DeadPlayerClearsSessionTargetWithoutPanic(t *testing.T) {
	game := New()
	requesterID := game.AddPlayer()
	targetID := game.AddPlayer()

	if !game.SetPlayerTarget(requesterID, targetID) {
		t.Fatal("expected setup SetPlayerTarget to succeed")
	}

	game.mu.Lock()
	delete(game.entities.Players, requesterID)
	game.mu.Unlock()

	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("expected ClearTarget while dead not to panic, got %v", recovered)
			}
		}()
		game.ClearTarget(requesterID)
	}()

	target := game.Target(requesterID)
	if !target.IsEmpty() {
		t.Fatalf("expected target to be empty after clear while dead, got %#v", target)
	}
}
