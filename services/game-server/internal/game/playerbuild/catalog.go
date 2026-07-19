package playerbuild

import (
	"fmt"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/weapons"
)

const (
	ShipVWing   = "v_wing"
	WeaponPulse = "pulse"
)

func DefaultCatalog() Catalog {
	return Catalog{
		Ships: map[string]ShipVariant{
			ShipVWing: {
				ID:          ShipVWing,
				WeightClass: WeightStandard,
				Stats:       runtime.ResolveShipStats(runtime.DefaultShipTypeID),
				WeaponPoints: map[WeaponPoint]PointKind{
					Primary1:   PointHardpoint,
					Primary2:   PointNone,
					Secondary1: PointHardpoint,
					Secondary2: PointNone,
				},
				ModuleSlots:            []ModuleSlot{ShieldMod, ArmorMod, EngineMod, UtilityMod},
				DefaultPrimaryWeaponID: WeaponPulse,
			},
		},
		Weapons: map[string]WeaponProfile{
			WeaponPulse: {
				ID:              WeaponPulse,
				RuntimeID:       weapons.BasicCannon,
				Slot:            weapons.Primary,
				Size:            WeaponStandard,
				DeliveryClass:   "ballistic",
				TargetingPolicy: "skill_shot",
				EffectFlags:     []EffectFlag{"direct"},
				AmmoPolicy:      weapons.InfiniteAmmo,
			},
		},
		Modules: map[string]ModuleProfile{},
	}
}

func (catalog Catalog) Validate() error {
	if len(catalog.Ships) == 0 {
		return fmt.Errorf("player build catalog requires at least one ship")
	}
	for id, ship := range catalog.Ships {
		if id == "" || ship.ID != id {
			return fmt.Errorf("ship catalog key must match ship ID")
		}
		if ship.WeightClass == "" {
			return fmt.Errorf("ship %q requires a weight class", id)
		}
		if ship.WeaponPoints[Primary1] != PointHardpoint {
			return fmt.Errorf("ship %q requires primary_1 hardpoint", id)
		}
		defaultWeapon, ok := catalog.Weapons[ship.DefaultPrimaryWeaponID]
		if !ok {
			return fmt.Errorf("ship %q default primary weapon is unknown", id)
		}
		if defaultWeapon.Slot != weapons.Primary {
			return fmt.Errorf("ship %q default primary weapon is not primary", id)
		}
		for point, kind := range ship.WeaponPoints {
			if point == "" {
				return fmt.Errorf("ship %q has an empty weapon point", id)
			}
			if kind != PointNone && kind != PointHardpoint && kind != PointSoftpoint {
				return fmt.Errorf("ship %q has unsupported point kind %q", id, kind)
			}
		}
	}
	for id, weapon := range catalog.Weapons {
		if id == "" || weapon.ID != id || weapon.RuntimeID == "" {
			return fmt.Errorf("weapon catalog key, ID, and runtime ID are required")
		}
		if weapon.Slot != weapons.Primary && weapon.Slot != weapons.Secondary {
			return fmt.Errorf("weapon %q has unsupported slot %q", id, weapon.Slot)
		}
		if weapon.AmmoPolicy == weapons.LimitedAmmo && weapon.StartingAmmo <= 0 {
			return fmt.Errorf("limited-ammo weapon %q requires positive starting ammo", id)
		}
	}
	for id, module := range catalog.Modules {
		if id == "" || module.ID != id || module.Slot == "" {
			return fmt.Errorf("module catalog key, ID, and slot are required")
		}
		if module.Activation != ModulePassive && module.Activation != ModuleActive {
			return fmt.Errorf("module %q has unsupported activation %q", id, module.Activation)
		}
	}
	return nil
}
