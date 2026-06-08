package game

import (
	pickuprules "github.com/Lokee86/space-rocks/server/internal/game/pickups"
	"github.com/Lokee86/space-rocks/server/internal/game/events"
	"github.com/Lokee86/space-rocks/server/internal/game/weapons"
)

func (game *Game) applyPickupEffectIntentLocked(intent pickuprules.EffectIntent) bool {
	if intent.EffectType == "" {
		return false
	}

	switch intent.EffectType {
	case pickuprules.EffectTypeAddLives:
		change := game.addPlayerLivesLocked(intent.PlayerID, intent.Amount)
		if !change.Found {
			return false
		}
		game.recordDomainEvent(events.Event{
			Type:       events.EventPickupEffectApplied,
			PlayerID:   intent.PlayerID,
			PickupID:   intent.PickupID,
			PickupType: intent.PickupType,
			EffectType: intent.EffectType,
			Amount:     intent.Amount,
			LivesAfter: change.After,
		})
		return true
	case pickuprules.EffectTypeEquipWeapon:
		player, ok := game.entities.Players[intent.PlayerID]
		if !ok || player == nil {
			return false
		}
		if intent.WeaponID == "" {
			return false
		}

		equipped := weapons.Equipped{
			ID:         intent.WeaponID,
			AmmoPolicy: weapons.LimitedAmmo,
		}
		switch intent.Slot {
		case weapons.Primary:
			player.ShipWeapons.Primary = equipped
			player.WeaponState.Primary.AmmoRemaining += intent.Ammo
		case weapons.Secondary:
			player.ShipWeapons.Secondary = equipped
			player.WeaponState.Secondary.AmmoRemaining += intent.Ammo
		default:
			return false
		}

		game.recordDomainEvent(events.Event{
			Type:       events.EventPickupEffectApplied,
			PlayerID:   intent.PlayerID,
			PickupID:   intent.PickupID,
			PickupType: intent.PickupType,
			EffectType: intent.EffectType,
		})
		return true
	default:
		return false
	}
}
