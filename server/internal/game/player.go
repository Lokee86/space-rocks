package game

import (
	"math"

	"github.com/Lokee86/space-rocks/server/internal/constants"
)

type Player struct {
	X        float64
	Y        float64
	Rotation float64
	Velocity Vector2
	Input    InputState
}

func (player *Player) State() PlayerState {
	return PlayerState{
		X:        player.X,
		Y:        player.Y,
		Rotation: player.Rotation,
	}
}

func (player *Player) applyInput(delta float64) {
	rotationInput := axis(player.Input.Left, player.Input.Right)
	thrustInput := axis(player.Input.Back, player.Input.Forward)

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
