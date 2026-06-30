package gametests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestAsteroidAcrossWorldEdgeIsNearCameraView(t *testing.T) {
	scenario := newScenario(t)
	scenario.addCameraView("observer", physics.Vector2{X: 5, Y: 100}, runtime.ClientConfig{
		VisibleWorldWidth:  200,
		VisibleWorldHeight: 200,
	})
	scenario.placeAsteroid("asteroid-1", physics.Vector2{X: constants.WorldWidth - 5, Y: 100}, 1)

	scenario.step(0)

	if !scenario.asteroidExists("asteroid-1") {
		t.Fatal("expected cross-edge nearby asteroid to remain visible")
	}
}

func TestBulletAcrossWorldEdgeIsNearCameraView(t *testing.T) {
	scenario := newScenario(t)
	scenario.addCameraView("observer", physics.Vector2{X: 5, Y: 100}, runtime.ClientConfig{
		VisibleWorldWidth:  200,
		VisibleWorldHeight: 200,
	})
	scenario.placeBullet(
		"bullet-1",
		"player-1",
		physics.Vector2{X: constants.WorldWidth - 5, Y: 100},
		physics.Vector2{},
	)
	scenario.bullet("bullet-1").Life = 10

	scenario.step(0)

	if _, ok := scenario.bulletSnapshot("observer", "bullet-1"); !ok {
		t.Fatal("expected cross-edge nearby bullet to remain visible")
	}
}

func TestFarAsteroidStillDespawns(t *testing.T) {
	scenario := newScenario(t)
	scenario.addCameraView("observer", physics.Vector2{X: 100, Y: 100}, runtime.ClientConfig{
		VisibleWorldWidth:  200,
		VisibleWorldHeight: 200,
	})
	scenario.placeAsteroid("asteroid-1", physics.Vector2{X: 600, Y: 100}, 1)

	scenario.step(0)

	if scenario.asteroidExists("asteroid-1") {
		t.Fatal("expected far asteroid to despawn")
	}
}

func TestFarBulletStillDespawns(t *testing.T) {
	scenario := newScenario(t)
	scenario.addCameraView("observer", physics.Vector2{X: 100, Y: 100}, runtime.ClientConfig{
		VisibleWorldWidth:  200,
		VisibleWorldHeight: 200,
	})
	scenario.placeBullet(
		"bullet-1",
		"player-1",
		physics.Vector2{X: 600, Y: 100},
		physics.Vector2{},
	)
	scenario.bullet("bullet-1").Life = 10

	scenario.step(0)

	if _, ok := scenario.bulletSnapshot("observer", "bullet-1"); ok {
		t.Fatal("expected far bullet to despawn")
	}
}
