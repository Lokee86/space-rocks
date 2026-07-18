package game

import (
	"sync"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/measurement"
)

// AttachRuntimeMeasurement attaches one independent observer to this Game.
// The returned function is safe to call repeatedly and detaches only this observer.
func (game *Game) AttachRuntimeMeasurement(observer measurement.SimulationObserver) func() {
	if observer == nil {
		return func() {}
	}

	game.runtimeMeasurementMu.Lock()
	game.nextMeasurementObserverID++
	id := game.nextMeasurementObserverID
	game.runtimeMeasurements[id] = observer
	game.runtimeMeasurementMu.Unlock()

	var detachOnce sync.Once
	return func() {
		detachOnce.Do(func() {
			game.runtimeMeasurementMu.Lock()
			delete(game.runtimeMeasurements, id)
			game.runtimeMeasurementMu.Unlock()
		})
	}
}

func (game *Game) HasRuntimeMeasurements() bool {
	game.runtimeMeasurementMu.RLock()
	defer game.runtimeMeasurementMu.RUnlock()
	return len(game.runtimeMeasurements) > 0
}

func (game *Game) observeRuntimeMeasurement(duration time.Duration, entities measurement.EntityCounts) {
	game.runtimeMeasurementMu.RLock()
	observers := make([]measurement.SimulationObserver, 0, len(game.runtimeMeasurements))
	for _, observer := range game.runtimeMeasurements {
		observers = append(observers, observer)
	}
	game.runtimeMeasurementMu.RUnlock()

	for _, observer := range observers {
		observer.ObserveSimulation(duration, entities)
	}
}

func (game *Game) runtimeMeasurementEntityCountsLocked() measurement.EntityCounts {
	return measurement.EntityCounts{
		Players:               len(game.entities.Players),
		PlayerSessions:        len(game.playerSessions),
		Enemies:               len(game.entities.Enemies),
		Asteroids:             len(game.entities.Asteroids),
		Projectiles:           len(game.entities.Projectiles),
		Pickups:               len(game.entities.Pickups),
		RadialEffects:         game.radialEffects.Len(),
		AsteroidsSpawnedTotal: game.spawner.TotalAsteroidsSpawned(),
	}
}
