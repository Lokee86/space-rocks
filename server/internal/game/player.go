package game

import (
	"math"
	"time"

	"github.com/Lokee86/space-rocks/server/internal/constants"
)

type Player struct {
	X        float64
	Y        float64
	Rotation float64
	Velocity Vector2
	LastTick time.Time
}

func (player *Player) State() PlayerState {
	return PlayerState{
		X:        player.X,
		Y:        player.Y,
		Rotation: player.Rotation,
	}
}

func (player *Player) applyInput(input InputState) {
	delta := player.nextDelta()
	rotationInput := axis(input.Left, input.Right)
	thrustInput := axis(input.Back, input.Forward)

	player.Rotation += rotationInput * constants.PlayerRotationSpeed * delta

	if thrustInput != 0 {
		player.Velocity.X += math.Sin(player.Rotation) * constants.PlayerThrustForce * thrustInput * delta
		player.Velocity.Y += -math.Cos(player.Rotation) * constants.PlayerThrustForce * thrustInput * delta
	}

	damping := math.Pow(constants.PlayerDamping, delta/(1.0/60.0))
	player.Velocity.X *= damping
	player.Velocity.Y *= damping
	player.Velocity = player.Velocity.limitLength(constants.PlayerMaxSpeed)

	player.X += player.Velocity.X * delta
	player.Y += player.Velocity.Y * delta
}

func (player *Player) nextDelta() float64 {
	now := time.Now()
	if player.LastTick.IsZero() {
		player.LastTick = now
		return 1.0 / 60.0
	}

	delta := now.Sub(player.LastTick).Seconds()
	player.LastTick = now

	return min(delta, 0.05)
}

func axis(negative bool, positive bool) float64 {
	var value float64
	if negative {
		value -= 1
	}
	if positive {
		value += 1
	}

	return max(-1, min(value, 1))
}
