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

type continuousBulletStreams struct {
	streamEmitters []continuousBulletStreamEntry
}

type DevtoolsContinuousBulletStream struct {
	OwnerPlayerID     string
	Origin            physics.Vector2
	Direction         physics.Vector2
	CooldownRemaining float64
}

func (streams *continuousBulletStreams) start(ownerPlayerID string, origin physics.Vector2, direction physics.Vector2) bool {
	if ownerPlayerID == "" {
		return false
	}
	normalizedDirection := direction.Normalized()
	if normalizedDirection.Length() == 0 {
		return false
	}
	streams.add(continuousBulletStreamEntry{
		OwnerPlayerID:     ownerPlayerID,
		Origin:            space.NormalizePosition(origin),
		Direction:         normalizedDirection,
		CooldownRemaining: constants.BulletCooldown,
	})
	return true
}

func (streams *continuousBulletStreams) active() []DevtoolsContinuousBulletStream {
	active := make([]DevtoolsContinuousBulletStream, len(streams.streamEmitters))
	for index, entry := range streams.streamEmitters {
		active[index] = DevtoolsContinuousBulletStream(entry)
	}
	return active
}

func (streams *continuousBulletStreams) add(stream continuousBulletStreamEntry) {
	streams.streamEmitters = append(streams.streamEmitters, stream)
}

func (streams *continuousBulletStreams) reset() {
	streams.streamEmitters = nil
}

func (streams *continuousBulletStreams) step(delta float64, bulletsCanMove bool, spawn func(string, physics.Vector2, physics.Vector2) bool) {
	for index := range streams.streamEmitters {
		stream := &streams.streamEmitters[index]
		stream.CooldownRemaining -= delta
		if stream.CooldownRemaining > 0 {
			continue
		}
		if bulletsCanMove {
			if spawn(stream.OwnerPlayerID, stream.Origin, stream.Direction) {
				stream.CooldownRemaining = constants.BulletCooldown
			}
		}
	}
}

func (game *Game) DevtoolsBeginContinuousBulletStream(ownerPlayerID string, origin physics.Vector2, direction physics.Vector2) bool {
	return game.continuousBulletStreams.start(ownerPlayerID, origin, direction)
}

func (game *Game) DevtoolsActiveContinuousBulletStreams() []DevtoolsContinuousBulletStream {
	return game.continuousBulletStreams.active()
}

func (game *Game) DevtoolsClearContinuousBulletStreams() {
	game.devtoolsResetContinuousBulletStreams()
}

func (game *Game) devtoolsContinuousBulletStreams() []continuousBulletStreamEntry {
	return game.continuousBulletStreams.streamEmitters
}

func (game *Game) devtoolsAppendContinuousBulletStream(stream continuousBulletStreamEntry) {
	game.continuousBulletStreams.add(stream)
}

func (game *Game) devtoolsResetContinuousBulletStreams() {
	game.continuousBulletStreams.reset()
}

func (game *Game) stepContinuousBulletStreams(delta float64) {
	game.continuousBulletStreams.step(delta, game.worldSimulationOptions.BulletsCanMove(), func(ownerPlayerID string, origin physics.Vector2, direction physics.Vector2) bool {
		_, spawned := game.spawnDebugBullet(ownerPlayerID, origin, direction)
		return spawned
	})
}
