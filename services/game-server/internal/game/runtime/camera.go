package runtime

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
)

const (
	BaseVisibleWorldWidth  = 1280.0
	BaseVisibleWorldHeight = 720.0
	MinCameraZoom          = 0.5
	MaxCameraZoom          = 2.0

	MinVisibleWorldWidth  = BaseVisibleWorldWidth / MaxCameraZoom
	MinVisibleWorldHeight = BaseVisibleWorldHeight / MaxCameraZoom
	MaxVisibleWorldWidth  = BaseVisibleWorldWidth / MinCameraZoom
	MaxVisibleWorldHeight = BaseVisibleWorldHeight / MinCameraZoom
)

func DefaultCameraConfig() ClientConfig {
	return ClientConfig{
		VisibleWorldWidth:  BaseVisibleWorldWidth,
		VisibleWorldHeight: BaseVisibleWorldHeight,
	}
}

func ClampCameraConfig(config ClientConfig) ClientConfig {
	if config.VisibleWorldWidth <= 0 || config.VisibleWorldHeight <= 0 {
		return DefaultCameraConfig()
	}

	widthZoom := BaseVisibleWorldWidth / config.VisibleWorldWidth
	heightZoom := BaseVisibleWorldHeight / config.VisibleWorldHeight
	zoom := min(max(max(widthZoom, heightZoom), MinCameraZoom), MaxCameraZoom)
	return ClientConfig{
		VisibleWorldWidth:  BaseVisibleWorldWidth / zoom,
		VisibleWorldHeight: BaseVisibleWorldHeight / zoom,
	}
}

func (view *CameraView) SetConfig(config ClientConfig) {
	view.Config = ClampCameraConfig(config)
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

	return BaseVisibleWorldWidth
}

func (view *CameraView) VisibleWorldHeight() float64 {
	if view.Config.VisibleWorldHeight > 0 {
		return view.Config.VisibleWorldHeight
	}

	return BaseVisibleWorldHeight
}
