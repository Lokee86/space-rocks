package playerbuild

import "testing"

func TestValidateSelectionAcceptsFallback(t *testing.T) {
	inventory := testInventory()
	catalog := testCatalog()
	options := ComputeEligibility("player-1", inventory, catalog, Rules{ModeID: "arcade_survival"})
	if err := ValidateSelection(options.FallbackLoadout, inventory, catalog, options); err != nil {
		t.Fatalf("fallback should validate: %v", err)
	}
}

func TestValidateSelectionRejectsWeaponOnUnavailablePoint(t *testing.T) {
	inventory := testInventory()
	catalog := testCatalog()
	options := ComputeEligibility("player-1", inventory, catalog, Rules{})
	selection := options.FallbackLoadout
	selection.SelectedWeaponsByPoint[Primary2] = "weapon-1"
	if err := ValidateSelection(selection, inventory, catalog, options); err == nil {
		t.Fatal("expected unavailable primary_2 to fail")
	}
}

func TestValidateSelectionRejectsMissingPrimary(t *testing.T) {
	inventory := testInventory()
	catalog := testCatalog()
	options := ComputeEligibility("player-1", inventory, catalog, Rules{})
	selection := options.FallbackLoadout
	delete(selection.SelectedWeaponsByPoint, Primary1)
	if err := ValidateSelection(selection, inventory, catalog, options); err == nil {
		t.Fatal("expected missing primary_1 to fail")
	}
}
