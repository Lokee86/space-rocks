package devtools

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
)

func handleDebugBeginContinuousBulletStream(controller *Controller, playerID string, command DebugCommand) bool {
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

	if !controller.streams.BeginContinuousBulletStream(playerID, origin, direction) {
		logging.Game.Info("debug begin continuous bullet stream ignored",
			logging.FieldPlayerID, playerID,
		)
		return true
	}

	controller.observerRegistry.RegisterOnce(controller.target, func(delta float64) {
		controller.streams.StepContinuousBulletStreams(delta, controller.target.BulletsCanMove(), controller.target.SpawnDebugBullet)
	})

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
