package playerbuild

import "testing"

func TestComputeEligibilityBuildsFallbackFromOwnedInventory(t *testing.T) {
	inventory := testInventory()
	options := ComputeEligibility("player-1", inventory, testCatalog(), Rules{ModeID: "arcade_survival"})
	if len(options.EligibleShips) != 1 {
		t.Fatalf("expected one eligible ship, got %d", len(options.EligibleShips))
	}
	if got := options.FallbackLoadout.SelectedOwnedShipID; got != inventory.DefaultOwnedShipID {
		t.Fatalf("expected default ship %q, got %q", inventory.DefaultOwnedShipID, got)
	}
	if got := options.FallbackLoadout.SelectedWeaponsByPoint[Primary1]; got != "weapon-1" {
		t.Fatalf("expected starter primary weapon, got %q", got)
	}
	if len(options.WeaponsByPoint[Primary2]) != 0 {
		t.Fatalf("unavailable primary_2 should not expose options: %#v", options.WeaponsByPoint[Primary2])
	}
	if len(options.BlockedOptions) != 1 || options.BlockedOptions[0].ReasonCode != ReasonUnknownCatalog {
		t.Fatalf("expected unknown weapon block, got %#v", options.BlockedOptions)
	}
}

func TestComputeEligibilityAppliesModeRestrictions(t *testing.T) {
	options := ComputeEligibility("player-1", testInventory(), testCatalog(), Rules{
		ModeID:                 "restricted",
		BannedShipIDs:          []string{ShipVWing},
		ModuleActivationPolicy: ModulesPassiveOnly,
	})
	if len(options.EligibleShips) != 0 {
		t.Fatal("banned ship should not be eligible")
	}
	assertBlockedReason(t, options, "ship-1", ReasonShipBanned)
	assertBlockedReason(t, options, "module-active", ReasonActiveModuleBlocked)
}

func TestComputeEligibilityRequiresConfiguredWeaponEffects(t *testing.T) {
	options := ComputeEligibility("player-1", testInventory(), testCatalog(), Rules{
		RequiredWeaponEffectFlags: []EffectFlag{"area"},
	})
	assertBlockedReason(t, options, "weapon-1", ReasonEffectFlagMissing)
}

func assertBlockedReason(t *testing.T, options EligibleBuildOptions, ownedID, reason string) {
	t.Helper()
	for _, blocked := range options.BlockedOptions {
		if blocked.OwnedInstanceID == ownedID && blocked.ReasonCode == reason {
			return
		}
	}
	t.Fatalf("missing blocked reason %q for %q: %#v", reason, ownedID, options.BlockedOptions)
}
