package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/weapons"
)

func (game *Game) firePlayerPrimaryWeapon(playerID string, player *runtime.Ship) bool {
	result := weapons.Fire(weapons.FireRequest{
		Equipped: player.ShipWeapons.Primary,
		State:    player.WeaponState.Primary,
		Position: player.Position(),
		Forward:  player.Forward(),
		Rotation: player.Rotation,
	})
	if !result.Fired {
		return false
	}

	id := game.spawner.NextBulletID()
	bullet := runtime.NewBulletFromWeaponSpawn(id, playerID, result.Projectile)
	game.entities.Projectiles[id] = bullet
	player.WeaponState.Primary = result.NewState
	game.refreshDepletedPrimaryWeapon(playerID, player)
	return true
}

func (game *Game) firePlayerSecondaryWeapon(playerID string, player *runtime.Ship) bool {
	result := weapons.Fire(weapons.FireRequest{
		Equipped: player.ShipWeapons.Secondary,
		State:    player.WeaponState.Secondary,
		Position: player.Position(),
		Forward:  player.Forward(),
		Rotation: player.Rotation,
	})
	if !result.Fired {
		return false
	}

	id := game.spawner.NextBulletID()
	bullet := runtime.NewBulletFromWeaponSpawn(id, playerID, result.Projectile)
	game.entities.Projectiles[id] = bullet
	player.WeaponState.Secondary = result.NewState
	game.refreshDepletedSecondaryWeapon(playerID, player)
	return true
}

func (game *Game) refreshDepletedPrimaryWeapon(playerID string, player *runtime.Ship) {
	if player.ShipWeapons.Primary.AmmoPolicy != weapons.LimitedAmmo {
		return
	}
	if player.WeaponState.Primary.AmmoRemaining > 0 {
		return
	}

	session, ok := game.playerSessions[playerID]
	if !ok || session == nil {
		return
	}

	player.ShipWeapons.Primary = session.PlayerArmory.Primary
	player.WeaponState.Primary = weapons.SlotState{}
}

func (game *Game) refreshDepletedSecondaryWeapon(playerID string, player *runtime.Ship) {
	if player.ShipWeapons.Secondary.AmmoPolicy != weapons.LimitedAmmo {
		return
	}
	if player.WeaponState.Secondary.AmmoRemaining > 0 {
		return
	}

	session, ok := game.playerSessions[playerID]
	if !ok || session == nil {
		return
	}

	player.ShipWeapons.Secondary = session.PlayerArmory.Secondary
	player.WeaponState.Secondary = weapons.SlotState{}
}
