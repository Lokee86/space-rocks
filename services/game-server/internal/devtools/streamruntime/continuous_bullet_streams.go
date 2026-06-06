package streamruntime

import (
	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
)

type ContinuousBulletStream struct {
	OwnerPlayerID     string
	Origin            physics.Vector2
	Direction         physics.Vector2
	CooldownRemaining float64
}

type ContinuousBulletStreams struct {
	streams []ContinuousBulletStream
}

func NewContinuousBulletStreams() *ContinuousBulletStreams {
	return &ContinuousBulletStreams{}
}

func (streams *ContinuousBulletStreams) Begin(ownerPlayerID string, origin physics.Vector2, direction physics.Vector2) bool {
	if ownerPlayerID == "" {
		return false
	}

	normalizedDirection := direction.Normalized()
	if normalizedDirection.Length() == 0 {
		return false
	}

	streams.streams = append(streams.streams, ContinuousBulletStream{
		OwnerPlayerID:     ownerPlayerID,
		Origin:            space.NormalizePosition(origin),
		Direction:         normalizedDirection,
		CooldownRemaining: constants.BasicCannonCooldown,
	})
	return true
}

func (streams *ContinuousBulletStreams) Active() []ContinuousBulletStream {
	active := make([]ContinuousBulletStream, len(streams.streams))
	copy(active, streams.streams)
	return active
}

func (streams *ContinuousBulletStreams) Clear() {
	streams.streams = nil
}

func (streams *ContinuousBulletStreams) Step(delta float64, bulletsCanMove bool, spawn func(string, physics.Vector2, physics.Vector2) bool) {
	for index := range streams.streams {
		stream := &streams.streams[index]
		stream.CooldownRemaining -= delta
		if stream.CooldownRemaining > 0 {
			continue
		}
		if bulletsCanMove {
			if spawn(stream.OwnerPlayerID, stream.Origin, stream.Direction) {
				stream.CooldownRemaining = constants.BasicCannonCooldown
			}
		}
	}
}
