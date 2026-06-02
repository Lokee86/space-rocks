package game

import (
	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
)

type continuousBulletStreamEntry struct {
	OwnerPlayerID     string
	Origin            physics.Vector2
	Direction         physics.Vector2
	CooldownRemaining float64
}

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
	game.devtoolsAppendContinuousBulletStream(continuousBulletStreamEntry(stream))
	return true
}

func (game *Game) DevtoolsActiveContinuousBulletStreams() []DevtoolsContinuousBulletStream {
	entries := game.devtoolsContinuousBulletStreams()
	streams := make([]DevtoolsContinuousBulletStream, len(entries))
	for index, entry := range entries {
		streams[index] = DevtoolsContinuousBulletStream(entry)
	}
	return streams
}

func (game *Game) DevtoolsClearContinuousBulletStreams() {
	game.devtoolsResetContinuousBulletStreams()
}

func (game *Game) devtoolsContinuousBulletStreams() []continuousBulletStreamEntry {
	return game.streamEmitters
}

func (game *Game) devtoolsAppendContinuousBulletStream(stream continuousBulletStreamEntry) {
	game.streamEmitters = append(game.streamEmitters, stream)
}

func (game *Game) devtoolsResetContinuousBulletStreams() {
	game.streamEmitters = nil
}

func (game *Game) stepContinuousBulletStreams(delta float64) {
	for index := range game.devtoolsContinuousBulletStreams() {
		stream := &game.streamEmitters[index]
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
