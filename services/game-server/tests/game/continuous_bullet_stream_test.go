package gametests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestDevtoolsContinuousBulletStreamSpawnsBulletAfterCooldown(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	origin := physics.Vector2{X: 10, Y: 20}
	direction := physics.Vector2{X: 0, Y: -1}

	if !scenario.game.DevtoolsBeginContinuousBulletStream(playerID, origin, direction) {
		t.Fatal("expected continuous bullet stream to start")
	}
	scenario.step(constants.BulletCooldown)

	if bulletCount := scenario.bullets().Len(); bulletCount != 1 {
		t.Fatalf("expected 1 spawned bullet, got %d", bulletCount)
	}

	bullet := scenario.bullet("bullet-1")
	if bullet.OwnerID != playerID {
		t.Fatalf("expected bullet owner %q, got %q", playerID, bullet.OwnerID)
	}
}
