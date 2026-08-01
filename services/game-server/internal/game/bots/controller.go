package bots

import (
	"math"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
)

const (
	avoidLookaheadSeconds = 1.5
	avoidBaseRadius       = 90.0
	avoidSizeRadius       = 28.0
	turnDeadzone          = 0.08
	fireAimTolerance      = 0.18
	thrustAimTolerance    = 0.85
)

type AsteroidObservation struct {
	Position physics.Vector2
	Velocity physics.Vector2
	Size     int
}

type PlayerObservation struct {
	Position physics.Vector2
}

type Observation struct {
	Position  physics.Vector2
	Velocity  physics.Vector2
	Rotation  float64
	Asteroids []AsteroidObservation
	Players   []PlayerObservation
}

type Controller struct{}

func NewController() *Controller {
	return &Controller{}
}

func (controller *Controller) Decide(observation Observation) runtime.InputState {
	targetDelta := physics.Vector2{}
	targetDistance := math.MaxFloat64
	for _, player := range observation.Players {
		delta := space.Delta(observation.Position, player.Position)
		distance := delta.Length()
		if distance < targetDistance {
			targetDelta = delta
			targetDistance = distance
		}
	}
	if targetDistance == math.MaxFloat64 {
		targetDelta, targetDistance = nearestAsteroid(observation)
	}
	if targetDistance == math.MaxFloat64 {
		return runtime.InputState{Forward: true}
	}

	desiredDelta := targetDelta
	if avoidanceDelta, avoiding := imminentAvoidance(observation); avoiding {
		desiredDelta = avoidanceDelta
	}

	desiredAngle := math.Atan2(desiredDelta.X, -desiredDelta.Y)
	turnDelta := normalizeAngle(desiredAngle - observation.Rotation)
	input := runtime.InputState{
		Left:    turnDelta < -turnDeadzone,
		Right:   turnDelta > turnDeadzone,
		Forward: math.Abs(turnDelta) < thrustAimTolerance,
	}

	targetAngle := math.Atan2(targetDelta.X, -targetDelta.Y)
	targetTurnDelta := normalizeAngle(targetAngle - observation.Rotation)
	input.PrimaryFire = targetDistance > 0 && math.Abs(targetTurnDelta) < fireAimTolerance
	return input
}

func nearestAsteroid(observation Observation) (physics.Vector2, float64) {
	position := observation.Position
	nearestDelta := physics.Vector2{}
	nearestDistance := math.MaxFloat64
	for _, asteroid := range observation.Asteroids {
		delta := space.Delta(position, asteroid.Position)
		distance := delta.Length()
		if distance < nearestDistance {
			nearestDelta = delta
			nearestDistance = distance
		}
	}
	return nearestDelta, nearestDistance
}

func imminentAvoidance(observation Observation) (physics.Vector2, bool) {
	bestTime := math.MaxFloat64
	avoidance := physics.Vector2{}
	for _, asteroid := range observation.Asteroids {
		relativePosition := space.Delta(observation.Position, asteroid.Position)
		relativeVelocity := asteroid.Velocity.Subtract(observation.Velocity)
		velocitySquared := relativeVelocity.LengthSquared()
		if velocitySquared == 0 {
			continue
		}

		timeToClosest := -relativePosition.Dot(relativeVelocity) / velocitySquared
		if timeToClosest < 0 || timeToClosest > avoidLookaheadSeconds {
			continue
		}
		closestDelta := relativePosition.Add(relativeVelocity.Multiply(timeToClosest))
		avoidRadius := avoidBaseRadius + float64(max(asteroid.Size, 0))*avoidSizeRadius
		if closestDelta.Length() >= avoidRadius || timeToClosest >= bestTime {
			continue
		}

		bestTime = timeToClosest
		avoidance = closestDelta.Multiply(-1)
		if avoidance.LengthSquared() == 0 {
			avoidance = physics.Vector2{X: -relativePosition.Y, Y: relativePosition.X}
		}
	}
	return avoidance, bestTime != math.MaxFloat64
}

func normalizeAngle(angle float64) float64 {
	for angle > math.Pi {
		angle -= 2 * math.Pi
	}
	for angle < -math.Pi {
		angle += 2 * math.Pi
	}
	return angle
}
