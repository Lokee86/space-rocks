package runtime

import (
	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func (view *CameraView) SetConfig(config ClientConfig) {
	view.Config = config
}

func (view *CameraView) SetPosition(position physics.Vector2) {
	view.X = position.X
	view.Y = position.Y
}

func (view *CameraView) Position() physics.Vector2 {
	return physics.Vector2{X: view.X, Y: view.Y}
}

func (view *CameraView) IsInside(position physics.Vector2) bool {
	width := view.VisibleWorldWidth()
	height := view.VisibleWorldHeight()
	left := view.X - width*0.5
	right := view.X + width*0.5
	top := view.Y - height*0.5
	bottom := view.Y + height*0.5

	return position.X >= left &&
		position.X <= right &&
		position.Y >= top &&
		position.Y <= bottom
}

func (view *CameraView) IsFarFrom(position physics.Vector2) bool {
	width := view.VisibleWorldWidth()
	height := view.VisibleWorldHeight()
	left := view.X - width*0.5 - constants.AsteroidDespawnMargin
	right := view.X + width*0.5 + constants.AsteroidDespawnMargin
	top := view.Y - height*0.5 - constants.AsteroidDespawnMargin
	bottom := view.Y + height*0.5 + constants.AsteroidDespawnMargin

	return position.X < left ||
		position.X > right ||
		position.Y < top ||
		position.Y > bottom
}

func (view *CameraView) VisibleWorldWidth() float64 {
	if view.Config.VisibleWorldWidth > 0 {
		return view.Config.VisibleWorldWidth
	}

	return constants.WorldWidth
}

func (view *CameraView) VisibleWorldHeight() float64 {
	if view.Config.VisibleWorldHeight > 0 {
		return view.Config.VisibleWorldHeight
	}

	return constants.WorldHeight
}
