package game

import "github.com/Lokee86/space-rocks/server/internal/game/damage"
import "github.com/Lokee86/space-rocks/server/internal/game/runtime"

func applyDamageResultToAsteroid(asteroid *runtime.Asteroid, result damage.DamageResult) {
	asteroid.Health = result.RemainingHealth
}

func applyDamageResultToPlayer(player *runtime.Ship, result damage.DamageResult) {
	player.Health = result.RemainingHealth
	player.Shields = result.RemainingShield
}

func applyDamageResultToEnemy(enemy *runtime.Ship, result damage.DamageResult) {
	enemy.Health = result.RemainingHealth
	enemy.Shields = result.RemainingShield
}
