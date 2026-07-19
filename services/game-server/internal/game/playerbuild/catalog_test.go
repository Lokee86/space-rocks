package playerbuild

import "testing"

func TestDefaultCatalogValidates(t *testing.T) {
	catalog := DefaultCatalog()
	if err := catalog.Validate(); err != nil {
		t.Fatalf("default catalog should validate: %v", err)
	}
	ship := catalog.Ships[ShipVWing]
	if ship.WeightClass != WeightStandard {
		t.Fatalf("expected v_wing to be standard, got %q", ship.WeightClass)
	}
	weapon := catalog.Weapons[WeaponPulse]
	if weapon.RuntimeID == "" {
		t.Fatal("pulse must map to a runtime weapon")
	}
}

func TestCatalogRejectsSecondaryDefaultPrimaryWeapon(t *testing.T) {
	catalog := DefaultCatalog()
	weapon := catalog.Weapons[WeaponPulse]
	weapon.Slot = "secondary"
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
