package space

import (
	"math"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

type Bounds struct {
	Width  float64
	Height float64
}

func DefaultBounds() Bounds {
	return Bounds{
		Width:  constants.WorldWidth,
		Height: constants.WorldHeight,
	}
}

func WrapPosition(pos physics.Vector2, bounds Bounds) physics.Vector2 {
	return physics.Vector2{
		X: wrapCoordinate(pos.X, bounds.Width),
		Y: wrapCoordinate(pos.Y, bounds.Height),
	}
}

func ShortestDelta(from physics.Vector2, to physics.Vector2, bounds Bounds) physics.Vector2 {
	return physics.Vector2{
		X: shortestCoordinateDelta(to.X-from.X, bounds.Width),
		Y: shortestCoordinateDelta(to.Y-from.Y, bounds.Height),
	}
}

func WrappedDistance(from physics.Vector2, to physics.Vector2, bounds Bounds) float64 {
	return ShortestDelta(from, to, bounds).Length()
}

func wrapCoordinate(value float64, size float64) float64 {
	if size <= 0 {
		return value
	}

	wrapped := math.Mod(value, size)
	if wrapped < 0 {
		wrapped += size
	}
	return wrapped
}

func shortestCoordinateDelta(delta float64, size float64) float64 {
	if size <= 0 {
		return delta
	}

	halfSize := size / 2
	if delta > halfSize {
		return delta - size
	}
	if delta < -halfSize {
		return delta + size
	}
	return delta
}

func Delta(from physics.Vector2, to physics.Vector2) physics.Vector2 {
	return ShortestDelta(from, to, DefaultBounds())
}

func Distance(from physics.Vector2, to physics.Vector2) float64 {
	return WrappedDistance(from, to, DefaultBounds())
}

func Direction(from physics.Vector2, to physics.Vector2) physics.Vector2 {
	return Delta(from, to).Normalized()
}

func NormalizePosition(position physics.Vector2) physics.Vector2 {
	return WrapPosition(position, DefaultBounds())
}
