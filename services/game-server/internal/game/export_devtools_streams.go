package game

import (
	"github.com/Lokee86/space-rocks/server/internal/devtools/streamruntime"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

type DevtoolsContinuousBulletStream = streamruntime.ContinuousBulletStream

func (game *Game) DevtoolsBeginContinuousBulletStream(ownerPlayerID string, origin physics.Vector2, direction physics.Vector2) bool {
	return game.devtoolsRuntime.BeginContinuousBulletStream(ownerPlayerID, origin, direction)
}

func (game *Game) DevtoolsActiveContinuousBulletStreams() []DevtoolsContinuousBulletStream {
	return game.devtoolsRuntime.ActiveContinuousBulletStreams()
}

func (game *Game) DevtoolsClearContinuousBulletStreams() {
	game.devtoolsRuntime.ClearContinuousBulletStreams()
}

func (game *Game) DevtoolsStepContinuousBulletStreams(delta float64) {
	game.devtoolsRuntime.StepContinuousBulletStreams(delta, game.worldSimulationOptions.BulletsCanMove(), func(ownerPlayerID string, origin physics.Vector2, direction physics.Vector2) bool {
		_, spawned := game.spawnDebugBullet(ownerPlayerID, origin, direction)
		return spawned
	})
}
