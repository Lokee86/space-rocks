package devtools

import (
	"sync"

	"github.com/Lokee86/space-rocks/server/internal/devtools/streamruntime"
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

var continuousBulletStreamStepObservers = struct {
	mu         sync.Mutex
	registered map[*game.Game]struct{}
}{
	registered: make(map[*game.Game]struct{}),
}

func handleDebugBeginContinuousBulletStream(target *game.Game, playerID string, command DebugCommand) bool {
	if !command.HasDirection {
		logging.Game.Info("debug begin continuous bullet stream ignored: has_direction is false",
			logging.FieldPlayerID, playerID,
		)
		return true
	}

	origin, direction := continuousBulletStreamRequestFromCommand(command)
	if direction.Length() == 0 {
		logging.Game.Info("debug begin continuous bullet stream ignored: direction is zero",
			logging.FieldPlayerID, playerID,
		)
		return true
	}

	if !streamruntime.DefaultRuntime.BeginContinuousBulletStream(playerID, origin, direction) {
		logging.Game.Info("debug begin continuous bullet stream ignored",
			logging.FieldPlayerID, playerID,
		)
		return true
	}
	ensureContinuousBulletStreamStepObserver(target)

	normalizedDirection := direction.Normalized()
	logging.Game.Info("debug continuous bullet stream started",
		logging.FieldPlayerID, playerID,
		"x", command.X,
		"y", command.Y,
		"direction_x", normalizedDirection.X,
		"direction_y", normalizedDirection.Y,
	)
	return true
}

func continuousBulletStreamRequestFromCommand(command DebugCommand) (physics.Vector2, physics.Vector2) {
	origin := physics.Vector2{X: command.X, Y: command.Y}
	direction := physics.Vector2{X: command.DirectionX, Y: command.DirectionY}
	return origin, direction
}

func ensureContinuousBulletStreamStepObserver(target *game.Game) {
	continuousBulletStreamStepObservers.mu.Lock()
	defer continuousBulletStreamStepObservers.mu.Unlock()

	if _, ok := continuousBulletStreamStepObservers.registered[target]; ok {
		return
	}
	target.DevtoolsRegisterSimulationStepObserver(func(delta float64) {
		streamruntime.StepContinuousBulletStreams(delta, target.DevtoolsBulletsCanMove(), target.DevtoolsSpawnDebugBullet)
	})
	continuousBulletStreamStepObservers.registered[target] = struct{}{}
}
