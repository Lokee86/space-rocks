package visibility

import (
	"math"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
)

// Contains reports whether position is inside the wrap-aware camera rectangle
// expanded by margin on each edge.
func Contains(view runtime.CameraView, position physics.Vector2, margin float64) bool {
	delta := space.Delta(view.Position(), position)
	return math.Abs(delta.X) <= view.VisibleWorldWidth()*0.5+margin &&
		math.Abs(delta.Y) <= view.VisibleWorldHeight()*0.5+margin
}

// ContainsStrict is the exclusive-boundary variant used by spawn clearance.
func ContainsStrict(view runtime.CameraView, position physics.Vector2, margin float64) bool {
	delta := space.Delta(view.Position(), position)
	return math.Abs(delta.X) < view.VisibleWorldWidth()*0.5+margin &&
		math.Abs(delta.Y) < view.VisibleWorldHeight()*0.5+margin
}

// Outside reports whether position is outside the wrap-aware camera rectangle
// expanded by margin on each edge.
func Outside(view runtime.CameraView, position physics.Vector2, margin float64) bool {
	delta := space.Delta(view.Position(), position)
	return math.Abs(delta.X) > view.VisibleWorldWidth()*0.5+margin ||
		math.Abs(delta.Y) > view.VisibleWorldHeight()*0.5+margin
}
