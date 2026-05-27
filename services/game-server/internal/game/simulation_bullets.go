package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/motion"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
)

func (game *Game) stepBullets(delta float64, bounds space.Bounds) {
	for id, bullet := range game.state.Projectiles {
		if game.worldSimulationOptions.BulletsCanMove() {
			motion.AdvanceBullet(bullet, delta, bounds)
		}
		if bullet.ReadyForRemoval() {
			delete(game.state.Projectiles, id)
			continue
		}
		if bullet.IsExpired() || game.isBulletFarFromAllCameras(bullet) {
			delete(game.state.Projectiles, id)
		}
	}
}
