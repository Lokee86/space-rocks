package motion

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func TestApplyShipAsteroidRepulsionSeparatesSurvivors(t *testing.T) {
	ship := &runtime.Ship{
		X:     10,
		Stats: runtime.ShipStats{MaxSpeed: 100},
	}
	asteroid := &runtime.Asteroid{X: 0}
	ApplyShipAsteroidRepulsion(ship, asteroid, 50, 20)
	if ship.Velocity != (physics.Vector2{X: 50}) {
		t.Fatalf("ship velocity = %+v", ship.Velocity)
	}
	if asteroid.Velocity != (physics.Vector2{X: -20}) {
		t.Fatalf("asteroid velocity = %+v", asteroid.Velocity)
	}
}

func TestApplyShipAsteroidRepulsionUsesDeterministicOverlapFallback(t *testing.T) {
	ship := &runtime.Ship{Stats: runtime.ShipStats{MaxSpeed: 100}}
	asteroid := &runtime.Asteroid{}
	ApplyShipAsteroidRepulsion(ship, asteroid, 10, 5)
	if ship.Velocity != (physics.Vector2{Y: -10}) {
		t.Fatalf("ship velocity = %+v", ship.Velocity)
	}
	if asteroid.Velocity != (physics.Vector2{Y: 5}) {
		t.Fatalf("asteroid velocity = %+v", asteroid.Velocity)
	}
}
