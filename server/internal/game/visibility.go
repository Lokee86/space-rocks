package game

import (
	"math/rand"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func (game *Game) randomAsteroidSpawnPosition(targetView *entities.CameraView) physics.Vector2 {
	margin := constants.AsteroidSpawnMargin
	for attempts := 0; ; attempts++ {
		spawn := randomOffscreenPosition(targetView, margin)
		if !game.isOnscreenForAnyCamera(spawn) {
			return spawn
		}

		if attempts > 0 && attempts%16 == 0 {
			margin += constants.AsteroidSpawnMargin
		}
	}
}

func randomOffscreenPosition(view *entities.CameraView, margin float64) physics.Vector2 {
	width := view.VisibleWorldWidth()
	height := view.VisibleWorldHeight()
	left := view.X - width*0.5
	right := view.X + width*0.5
	top := view.Y - height*0.5
	bottom := view.Y + height*0.5

	switch rand.Intn(4) {
	case 0:
		return physics.Vector2{X: randomRange(left, right), Y: top - margin}
	case 1:
		return physics.Vector2{
			X: right + margin,
			Y: randomRange(top, bottom),
		}
	case 2:
		return physics.Vector2{
			X: randomRange(left, right),
			Y: bottom + margin,
		}
	default:
		return physics.Vector2{X: left - margin, Y: randomRange(top, bottom)}
	}
}

func (game *Game) isOnscreenForAnyCamera(position physics.Vector2) bool {
	for _, view := range game.cameraViews {
		if view.IsInside(position) {
			return true
		}
	}

	return false
}

func (game *Game) isAsteroidFarFromAllCameras(asteroid *entities.Asteroid) bool {
	if !game.hasCameraViews() {
		return true
	}

	for _, view := range game.cameraViews {
		if !view.IsFarFrom(asteroid.Position()) {
			return false
		}
	}

	return true
}

func (game *Game) isBulletFarFromAllCameras(bullet *entities.Bullet) bool {
	if !game.hasCameraViews() {
		return true
	}

	for _, view := range game.cameraViews {
		if !view.IsFarFrom(bullet.Position()) {
			return false
		}
	}

	return true
}

func (game *Game) hasCameraViews() bool {
	return len(game.cameraViews) > 0
}

func randomRange(minValue float64, maxValue float64) float64 {
	return minValue + rand.Float64()*(maxValue-minValue)
}
