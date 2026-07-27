package game

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/visibility"
)

func (game *Game) randomAsteroidSpawnPosition(targetView *runtime.CameraView) physics.Vector2 {
	margin := constants.AsteroidSpawnMargin
	for attempts := 0; ; attempts++ {
		spawn := game.randomOffscreenPosition(targetView, margin)
		if !game.isInsideAsteroidSpawnClearanceForAnyCamera(spawn) {
			return spawn
		}

		if attempts > 0 && attempts%16 == 0 {
			margin += constants.AsteroidSpawnMargin
		}
	}
}

func (game *Game) randomOffscreenPosition(view *runtime.CameraView, margin float64) physics.Vector2 {
	width := view.VisibleWorldWidth()
	height := view.VisibleWorldHeight()
	left := view.X - width*0.5
	right := view.X + width*0.5
	top := view.Y - height*0.5
	bottom := view.Y + height*0.5

	switch game.rngSource.Intn(4) {
	case 0:
		return physics.Vector2{X: game.randomRange(left, right), Y: top - margin}
	case 1:
		return physics.Vector2{
			X: right + margin,
			Y: game.randomRange(top, bottom),
		}
	case 2:
		return physics.Vector2{
			X: game.randomRange(left, right),
			Y: bottom + margin,
		}
	default:
		return physics.Vector2{X: left - margin, Y: game.randomRange(top, bottom)}
	}
}

func (game *Game) isInsideAsteroidSpawnClearanceForAnyCamera(position physics.Vector2) bool {
	for _, view := range game.cameraViews {
		if isInsideCameraViewWithMargin(view, position, constants.AsteroidSpawnMargin) {
			return true
		}
	}

	return false
}

func (game *Game) isAsteroidFarFromAllCameras(asteroid *runtime.Asteroid) bool {
	if !game.hasCameraViews() {
		return true
	}

	for _, view := range game.cameraViews {
		if !isFarFromCameraView(view, asteroid.Position()) {
			return false
		}
	}

	return true
}

func (game *Game) isBulletFarFromAllCameras(bullet *runtime.Bullet) bool {
	if !game.hasCameraViews() {
		return true
	}

	for _, view := range game.cameraViews {
		if !isFarFromCameraView(view, bullet.Position()) {
			return false
		}
	}

	return true
}

func (game *Game) hasCameraViews() bool {
	return len(game.cameraViews) > 0
}

func isInsideCameraView(view *runtime.CameraView, position physics.Vector2) bool {
	return view != nil && visibility.Contains(*view, position, 0)
}

func isInsideCameraViewWithMargin(view *runtime.CameraView, position physics.Vector2, margin float64) bool {
	return view != nil && visibility.ContainsStrict(*view, position, margin)
}

func isFarFromCameraView(view *runtime.CameraView, position physics.Vector2) bool {
	return view == nil || visibility.Outside(*view, position, constants.AsteroidDespawnMargin)
}

func (game *Game) randomRange(minValue float64, maxValue float64) float64 {
	return minValue + game.rngSource.Float64()*(maxValue-minValue)
}
