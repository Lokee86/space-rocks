package gametests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/devtools/streamruntime"
)

func TestDevtoolsContinuousBulletStreamSpawnsBulletAfterCooldown(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	origin := physics.Vector2{X: 10, Y: 20}
	direction := physics.Vector2{X: 0, Y: -1}
	runtime := streamruntime.NewRuntime()

	if !runtime.BeginContinuousBulletStream(playerID, origin, direction) {
		t.Fatal("expected continuous bullet stream to start")
	}
	runtime.StepContinuousBulletStreams(constants.BasicCannonCooldown, scenario.game.DevtoolsBulletsCanMove(), scenario.game.DevtoolsSpawnDebugBullet)

	if bulletCount := scenario.bullets().Len(); bulletCount != 1 {
		t.Fatalf("expected 1 spawned bullet, got %d", bulletCount)
	}

	bullet := scenario.bullet("bullet-1")
	if bullet.OwnerID != playerID {
		t.Fatalf("expected bullet owner %q, got %q", playerID, bullet.OwnerID)
	}
}
