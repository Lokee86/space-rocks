package game

import (
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
	"github.com/Lokee86/space-rocks/services/game-server/internal/measurement"
)

func (game *Game) runSimulation() {
	ticker := time.NewTicker(time.Second / time.Duration(constants.ServerTickRate))
	defer ticker.Stop()

	delta := 1.0 / float64(constants.ServerTickRate)
	for {
		select {
		case <-game.stopSimulation:
			return
		case <-ticker.C:
			game.Step(delta)
		}
	}
}

func (game *Game) Step(delta float64) {
	measure := game.HasRuntimeMeasurements()
	started := time.Time{}
	if measure {
		started = time.Now()
	}

	var entities measurement.EntityCounts
	game.mu.Lock()
	func() {
		defer game.mu.Unlock()
		defer game.publishPresentationFrameLocked()

		bounds := space.DefaultBounds()
		game.advanceEncounterLifecycle(delta)

		game.applyPendingPlayerInputsLocked()
		game.stepPlayerSessions(delta)
		if game.isMatchOverLocked() {
			game.encounterSpawn().Stop()
			game.stepPlayers(delta, bounds)
			game.removeReadyPlayers()
			game.stepAsteroids(delta, bounds)
			game.stepBullets(delta, bounds)
			game.stepPickups(delta)
			game.stepRadialEffects(delta)
			game.stepDamageOverTime(delta)
			for _, observer := range game.simulationStepObservers {
				observer(delta)
			}
		} else {
			game.stepBots()
			game.stepPlayerWeapons(delta)
			game.stepPlayers(delta, bounds)
			game.removeReadyPlayers()
			game.stepAsteroidSpawning(delta)
			game.stepAsteroids(delta, bounds)
			game.stepBullets(delta, bounds)
			game.stepPickups(delta)
			game.stepCollisions()
			game.stepRadialEffects(delta)
			game.stepDamageOverTime(delta)
			for _, observer := range game.simulationStepObservers {
				observer(delta)
			}
		}
		if measure {
			entities = game.runtimeMeasurementEntityCountsLocked()
		}
	}()

	if measure {
		game.observeRuntimeMeasurement(time.Since(started), entities)
	}
}

func (game *Game) stepCollisions() {
	if game.worldSimulationOptions.CanRunCollisions() {
		game.rebuildAsteroidSpatialIndex()
		game.handleShipAsteroidCollisions()
		game.handleBulletAsteroidCollisions()
		game.rebuildPickupSpatialIndex()
		game.handlePlayerPickupCollisions()
	}
}
