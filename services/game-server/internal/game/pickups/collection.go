package pickups

import "github.com/Lokee86/space-rocks/server/internal/game/weapons"

const EffectTypeAddLives = "add_lives"
const EffectTypeEquipWeapon = "equip_weapon"

type CollectionRequest struct {
	PlayerID   string
	PickupID   string
	PickupType string
	X          float64
	Y          float64
}

type EffectIntent struct {
	PlayerID   string
	PickupID   string
	PickupType string
	EffectType string
	Amount     int
	WeaponID   weapons.ID
	Slot       weapons.Slot
	Ammo       int
}

type CollectionResult struct {
	Collected    bool
	PlayerID     string
	PickupID     string
	PickupType   string
	X            float64
	Y            float64
	EffectIntent EffectIntent
}

func ResolveCollection(req CollectionRequest) CollectionResult {
	if req.PlayerID == "" || req.PickupID == "" || req.PickupType == "" {
		return CollectionResult{Collected: false}
	}

	effectIntent := resolveEffectIntent(req)
	return CollectionResult{
		Collected:    true,
		PlayerID:     req.PlayerID,
		PickupID:     req.PickupID,
		PickupType:   req.PickupType,
		X:            req.X,
		Y:            req.Y,
		EffectIntent: effectIntent,
	}
}

func resolveEffectIntent(req CollectionRequest) EffectIntent {
	switch req.PickupType {
	case "1_up":
		return EffectIntent{
			PlayerID:   req.PlayerID,
			PickupID:   req.PickupID,
			PickupType: req.PickupType,
			EffectType: EffectTypeAddLives,
			Amount:     1,
		}
	case "torpedo":
		return EffectIntent{
			PlayerID:   req.PlayerID,
			PickupID:   req.PickupID,
			PickupType: req.PickupType,
			EffectType: EffectTypeEquipWeapon,
			WeaponID:   weapons.Torpedo,
			Slot:       weapons.Secondary,
			Ammo:       1,
		}
	default:
		return EffectIntent{}
	}
}
