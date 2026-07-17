package devtools

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
)

func handleDebugBeginContinuousBulletStream(controller *Controller, playerID string, command DebugCommand) bool {
	if !command.HasDirection {
		return true
	}

	origin, direction := continuousBulletStreamRequestFromCommand(command)
	if direction.Length() == 0 {
		return true
	}

	if !controller.streams.BeginContinuousBulletStream(playerID, origin, direction) {
		return true
	}

	controller.observerRegistry.RegisterOnce(controller.target, func(delta float64, bulletsCanMove func() bool, spawnDebugBullet func(string, physics.Vector2, physics.Vector2) bool) {
		controller.streams.StepContinuousBulletStreams(delta, bulletsCanMove(), spawnDebugBullet)
	})

	return true
}

func continuousBulletStreamRequestFromCommand(command DebugCommand) (physics.Vector2, physics.Vector2) {
	origin := physics.Vector2{X: command.X, Y: command.Y}
	direction := physics.Vector2{X: command.DirectionX, Y: command.DirectionY}
	return origin, direction
}
