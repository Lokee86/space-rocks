package game

import "github.com/Lokee86/space-rocks/server/internal/game/weapons"

func (game *Game) stepPlayerWeapons(delta float64) {
	for _, player := range game.entities.Players {
		player.WeaponState = weapons.StepState(player.WeaponState, delta)
	}
}
