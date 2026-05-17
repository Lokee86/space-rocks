package game

import (
	"math/rand"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func (game *Game) randomAsteroidSpawnPosition(target *entities.Ship) physics.Vector2 {
	margin := constants.AsteroidSpawnMargin
	for attempts := 0; ; attempts++ {
		spawn := randomOffscreenPosition(target, margin)
		if !game.isOnscreenForAnyPlayer(spawn) {
			return spawn
		}

		if attempts > 0 && attempts%16 == 0 {
			margin += constants.AsteroidSpawnMargin
		}
	}
}

func randomOffscreenPosition(target *entities.Ship, margin float64) physics.Vector2 {
	width := target.VisibleWorldWidth()
	height := target.VisibleWorldHeight()
	left := target.X - width*0.5
	right := target.X + width*0.5
	top := target.Y - height*0.5
	bottom := target.Y + height*0.5

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

func (game *Game) isOnscreenForAnyPlayer(position physics.Vector2) bool {
	for _, player := range game.state.Players {
		if player.IsInsideView(position) {
			return true
		}
	}

	return false
}

func (game *Game) isAsteroidFarFromAllPlayers(asteroid *entities.Asteroid) bool {
	if len(game.state.Players) == 0 {
		return true
	}

	for _, player := range game.state.Players {
		if !player.IsFarFromView(asteroid.Position()) {
			return false
		}
	}

	return true
}

func (game *Game) isBulletFarFromAllPlayers(bullet *entities.Bullet) bool {
	if len(game.state.Players) == 0 {
		return true
	}

	for _, player := range game.state.Players {
		if !player.IsFarFromView(bullet.Position()) {
			return false
		}
	}

	return true
}

func randomRange(minValue float64, maxValue float64) float64 {
	return minValue + rand.Float64()*(maxValue-minValue)
}
