package game

import (
	"math/rand"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func (game *Game) randomAsteroidSpawnPosition(target *Ship) physics.Vector2 {
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

func randomOffscreenPosition(target *Ship, margin float64) physics.Vector2 {
	width := target.visibleWorldWidth()
	height := target.visibleWorldHeight()
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
		if player.isInsideView(position) {
			return true
		}
	}

	return false
}

func (game *Game) isAsteroidFarFromAllPlayers(asteroid *Asteroid) bool {
	if len(game.state.Players) == 0 {
		return true
	}

	for _, player := range game.state.Players {
		if !player.isFarFromView(asteroid.Position()) {
			return false
		}
	}

	return true
}

func (game *Game) isBulletFarFromAllPlayers(bullet *Bullet) bool {
	if len(game.state.Players) == 0 {
		return true
	}

	for _, player := range game.state.Players {
		if !player.isFarFromView(bullet.Position()) {
			return false
		}
	}

	return true
}

func (ship *Ship) isInsideView(position physics.Vector2) bool {
	width := ship.visibleWorldWidth()
	height := ship.visibleWorldHeight()
	left := ship.X - width*0.5
	right := ship.X + width*0.5
	top := ship.Y - height*0.5
	bottom := ship.Y + height*0.5

	return position.X >= left &&
		position.X <= right &&
		position.Y >= top &&
		position.Y <= bottom
}

func (ship *Ship) isFarFromView(position physics.Vector2) bool {
	width := ship.visibleWorldWidth()
	height := ship.visibleWorldHeight()
	left := ship.X - width*0.5 - constants.AsteroidDespawnMargin
	right := ship.X + width*0.5 + constants.AsteroidDespawnMargin
	top := ship.Y - height*0.5 - constants.AsteroidDespawnMargin
	bottom := ship.Y + height*0.5 + constants.AsteroidDespawnMargin

	return position.X < left ||
		position.X > right ||
		position.Y < top ||
		position.Y > bottom
}

func (ship *Ship) visibleWorldWidth() float64 {
	if ship.Config.VisibleWorldWidth > 0 {
		return ship.Config.VisibleWorldWidth
	}

	return constants.WorldWidth
}

func (ship *Ship) visibleWorldHeight() float64 {
	if ship.Config.VisibleWorldHeight > 0 {
		return ship.Config.VisibleWorldHeight
	}

	return constants.WorldHeight
}

func randomRange(minValue float64, maxValue float64) float64 {
	return minValue + rand.Float64()*(maxValue-minValue)
}
