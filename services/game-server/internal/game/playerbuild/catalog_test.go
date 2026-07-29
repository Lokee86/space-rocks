package playerbuild

import (
	"testing"

	"github.com/Lokee86/space-rocks/player-data/playerdata"
	"github.com/Lokee86/space-rocks/player-data/protocol"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/weapons"
)

func TestDefaultCatalogValidates(t *testing.T) {
	catalog := DefaultCatalog()
	if err := catalog.Validate(); err != nil {
		t.Fatalf("default catalog should validate: %v", err)
	}
	if len(catalog.Ships) != 2 || len(catalog.Weapons) != 2 || len(catalog.Modules) != 4 {
		t.Fatalf("unexpected production catalog sizes: ships=%d weapons=%d modules=%d", len(catalog.Ships), len(catalog.Weapons), len(catalog.Modules))
	}
	standard := catalog.Ships[ShipVWing]
	scout := catalog.Ships[ShipVWingScout]
	if standard.WeightClass != WeightStandard || scout.WeightClass != WeightLight {
		t.Fatalf("unexpected ship weight classes: standard=%q scout=%q", standard.WeightClass, scout.WeightClass)
	}
	if scout.Stats.MaxHealth >= standard.Stats.MaxHealth || scout.Stats.MaxSpeed <= standard.Stats.MaxSpeed {
		t.Fatalf("scout should trade survivability for mobility: standard=%#v scout=%#v", standard.Stats, scout.Stats)
	}
	if catalog.Weapons[WeaponPulse].RuntimeID != weapons.BasicCannon {
		t.Fatal("pulse must map to the basic cannon runtime weapon")
	}
	torpedo := catalog.Weapons[WeaponTorpedo]
	if torpedo.RuntimeID != weapons.Torpedo || torpedo.Slot != weapons.Secondary || torpedo.StartingAmmo != 3 {
		t.Fatalf("unexpected torpedo catalog profile: %#v", torpedo)
	}
	if catalog.Modules[ModuleShieldCapacitor].Adjustment.MaxShieldsDelta <= 0 {
		t.Fatal("shield capacitor must provide shields")
	}
}

func TestDefaultCatalogCoversStarterInventory(t *testing.T) {
	inventory := playerdata.StarterHangarInventory(protocol.PlayerDataIdentity{IdentityKind: playerdata.IdentityKindGuest})
	options := ComputeEligibility("player-1", inventory, DefaultCatalog(), Rules{})
	if len(options.BlockedOptions) != 0 {
		t.Fatalf("starter catalog contains blocked inventory: %#v", options.BlockedOptions)
	}
	if len(options.EligibleShips) != 2 || len(options.WeaponsByPoint[Primary1]) != 1 || len(options.WeaponsByPoint[Secondary1]) != 1 {
		t.Fatalf("starter catalog eligibility mismatch: %#v", options)
	}
	for _, slot := range []ModuleSlot{ShieldMod, ArmorMod, EngineMod, UtilityMod} {
		if len(options.ModulesBySlot[slot]) != 1 {
			t.Fatalf("starter catalog missing module option for %q: %#v", slot, options.ModulesBySlot)
		}
	}
	if options.FallbackLoadout.SelectedOwnedShipID != inventory.DefaultOwnedShipID {
		t.Fatalf("starter fallback did not preserve default ship")
	}
}

func TestCatalogRejectsSecondaryDefaultPrimaryWeapon(t *testing.T) {
	catalog := DefaultCatalog()
	weapon := catalog.Weapons[WeaponPulse]
	weapon.Slot = weapons.Secondary
	catalog.Weapons[WeaponPulse] = weapon
	if err := catalog.Validate(); err == nil {
		t.Fatal("expected secondary default primary weapon to fail validation")
	}
}

func TestCatalogRejectsMissingPrimaryHardpoint(t *testing.T) {
	catalog := DefaultCatalog()
	ship := catalog.Ships[ShipVWing]
	ship.WeaponPoints[Primary1] = PointSoftpoint
	catalog.Ships[ShipVWing] = ship
	if err := catalog.Validate(); err == nil {
		t.Fatal("expected missing primary hardpoint to fail validation")
	}
}
