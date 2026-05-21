package gametests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestPlayerCrossingRightEdgeWrapsToLeft(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	ship := scenario.player(playerID)
	ship.Velocity = physics.Vector2{X: 5}
	ship.Stats.Damping = 1

	scenario.setPlayerPosition(playerID, physics.Vector2{X: constants.WorldWidth - 2, Y: 100})
	scenario.step(1)

	player := scenario.playerState(playerID, playerID)
	if player.X != 3 || player.Y != 100 {
		t.Fatalf("expected player to wrap to (3, 100), got (%v, %v)", player.X, player.Y)
	}
}

func TestPlayerCrossingLeftEdgeWrapsToRight(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	ship := scenario.player(playerID)
	ship.Velocity = physics.Vector2{X: -5}
	ship.Stats.Damping = 1

	scenario.setPlayerPosition(playerID, physics.Vector2{X: 2, Y: 100})
	scenario.step(1)

	player := scenario.playerState(playerID, playerID)
	expectedX := constants.WorldWidth - 3
	if player.X != expectedX || player.Y != 100 {
		t.Fatalf("expected player to wrap to (%v, 100), got (%v, %v)", expectedX, player.X, player.Y)
	}
}

func TestAsteroidCrossingEdgeWraps(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	scenario.addCameraView("observer", physics.Vector2{X: 5, Y: 100}, entities.ClientConfig{
		VisibleWorldWidth:  400,
		VisibleWorldHeight: 400,
	})
	scenario.placeMovingAsteroid(
		"asteroid-1",
		physics.Vector2{X: constants.WorldWidth - 2, Y: 100},
		physics.Vector2{X: 5},
		1,
	)

	scenario.step(1)

	asteroid, ok := scenario.state(playerID).Asteroids["asteroid-1"]
	if !ok {
		t.Fatal("expected wrapped asteroid to remain in state")
	}
	if asteroid.X != 3 || asteroid.Y != 100 {
		t.Fatalf("expected asteroid to wrap to (3, 100), got (%v, %v)", asteroid.X, asteroid.Y)
	}
}

func TestBulletCrossingEdgeWraps(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	scenario.addCameraView("observer", physics.Vector2{X: 5, Y: 100}, entities.ClientConfig{
		VisibleWorldWidth:  400,
		VisibleWorldHeight: 400,
	})
	scenario.placeBullet(
		"bullet-1",
		playerID,
		physics.Vector2{X: constants.WorldWidth - 2, Y: 100},
		physics.Vector2{X: 5},
	)
	scenario.bullet("bullet-1").Life = 10

	scenario.step(1)

	bullet, ok := scenario.state(playerID).Bullets["bullet-1"]
	if !ok {
		t.Fatal("expected wrapped bullet to remain in state")
	}
	if bullet.X != 3 || bullet.Y != 100 {
		t.Fatalf("expected bullet to wrap to (3, 100), got (%v, %v)", bullet.X, bullet.Y)
	}
}
