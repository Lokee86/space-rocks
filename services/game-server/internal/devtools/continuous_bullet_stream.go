package devtools

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func handleDebugBeginContinuousBulletStream(target *game.Game, playerID string, command DebugCommand) bool {
	if !command.HasDirection {
		logging.Game.Info("debug begin continuous bullet stream ignored: has_direction is false",
			logging.FieldPlayerID, playerID,
		)
		return true
	}

	direction := physics.Vector2{X: command.DirectionX, Y: command.DirectionY}
	if direction.Length() == 0 {
		logging.Game.Info("debug begin continuous bullet stream ignored: direction is zero",
			logging.FieldPlayerID, playerID,
		)
		return true
	}

	origin := physics.Vector2{X: command.X, Y: command.Y}
	if !target.DevtoolsBeginContinuousBulletStream(playerID, origin, direction) {
		logging.Game.Info("debug begin continuous bullet stream ignored",
			logging.FieldPlayerID, playerID,
		)
		return true
	}

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
