package game

import (
	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
)

type DevtoolsContinuousBulletStream struct {
	OwnerPlayerID     string
	Origin            physics.Vector2
	Direction         physics.Vector2
	CooldownRemaining float64
}

func (game *Game) DevtoolsBeginContinuousBulletStream(ownerPlayerID string, origin physics.Vector2, direction physics.Vector2) bool {
	if ownerPlayerID == "" {
		return false
	}
	normalizedDirection := direction.Normalized()
	if normalizedDirection.Length() == 0 {
		return false
	}
	stream := DevtoolsContinuousBulletStream{
		OwnerPlayerID:     ownerPlayerID,
		Origin:            space.NormalizePosition(origin),
		Direction:         normalizedDirection,
		CooldownRemaining: constants.BulletCooldown,
	}
	game.activeDebugBulletStreams = append(game.activeDebugBulletStreams, stream)
	return true
}

func (game *Game) DevtoolsActiveContinuousBulletStreams() []DevtoolsContinuousBulletStream {
	streams := make([]DevtoolsContinuousBulletStream, len(game.activeDebugBulletStreams))
	copy(streams, game.activeDebugBulletStreams)
	return streams
}

func (game *Game) stepDevtoolsContinuousBulletStreams(delta float64) {
	for index := range game.activeDebugBulletStreams {
		stream := &game.activeDebugBulletStreams[index]
		stream.CooldownRemaining -= delta
		if stream.CooldownRemaining > 0 {
			continue
		}
		if game.worldSimulationOptions.BulletsCanMove() {
			_, spawned := game.spawnDebugBullet(stream.OwnerPlayerID, stream.Origin, stream.Direction)
			if spawned {
				stream.CooldownRemaining = constants.BulletCooldown
			}
		}
	}
}
