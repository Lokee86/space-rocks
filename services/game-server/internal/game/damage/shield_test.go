package damage

import "testing"

func TestResolveSingleFullShieldAbsorption(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "player-1",
			EntityType: EntityTypePlayer,
			Health:     10,
			Shield:     5,
		},
		Spec: DamageSpec{
			Amount: 3,
			Type:   DamageTypeThermal,
		},
	})

	if result.AbsorbedByShield != 3 {
		t.Fatalf("expected absorbed shield %d, got %d", 3, result.AbsorbedByShield)
	}
	if result.AppliedToHealth != 0 {
		t.Fatalf("expected applied health %d, got %d", 0, result.AppliedToHealth)
	}
	if result.RemainingShield != 2 {
		t.Fatalf("expected remaining shield %d, got %d", 2, result.RemainingShield)
	}
}

func TestResolveSinglePartialShieldAbsorption(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "player-1",
			EntityType: EntityTypePlayer,
			Health:     10,
			Shield:     2,
		},
		Spec: DamageSpec{
			Amount: 5,
			Type:   DamageTypeThermal,
		},
	})

	if result.AbsorbedByShield != 2 {
		t.Fatalf("expected absorbed shield %d, got %d", 2, result.AbsorbedByShield)
	}
	if result.AppliedToHealth != 3 {
		t.Fatalf("expected applied health %d, got %d", 3, result.AppliedToHealth)
	}
	if result.RemainingShield != 0 {
		t.Fatalf("expected remaining shield %d, got %d", 0, result.RemainingShield)
	}
}

func TestResolveSingleBypassShield(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "player-1",
			EntityType: EntityTypePlayer,
			Health:     10,
			Shield:     5,
		},
		Spec: DamageSpec{
			Amount:       3,
			Type:         DamageTypeThermal,
			BypassShield: true,
		},
	})

	if result.AbsorbedByShield != 0 {
		t.Fatalf("expected absorbed shield %d, got %d", 0, result.AbsorbedByShield)
	}
	if result.AppliedToHealth != 3 {
		t.Fatalf("expected applied health %d, got %d", 3, result.AppliedToHealth)
	}
	if result.RemainingShield != 5 {
		t.Fatalf("expected remaining shield %d, got %d", 5, result.RemainingShield)
	}
}
