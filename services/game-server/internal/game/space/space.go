package space

import "github.com/Lokee86/space-rocks/server/internal/game/physics"

func Delta(from physics.Vector2, to physics.Vector2) physics.Vector2 {
	return to.Subtract(from)
}

func Distance(from physics.Vector2, to physics.Vector2) float64 {
	return Delta(from, to).Length()
}

func Direction(from physics.Vector2, to physics.Vector2) physics.Vector2 {
	return Delta(from, to).Normalized()
}

func NormalizePosition(position physics.Vector2) physics.Vector2 {
	return position
}
