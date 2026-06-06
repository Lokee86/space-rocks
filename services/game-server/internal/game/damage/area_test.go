package damage

import "testing"

func TestResolveAreaEmptyCandidates(t *testing.T) {
	result := ResolveArea(AreaDamageRequest{
		Radius: 10,
		Spec: DamageSpec{
			Amount: 4,
			Type:   DamageTypeThermal,
		},
	})

	if len(result.Results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(result.Results))
	}
}

func TestResolveAreaMultipleCandidates(t *testing.T) {
	result := ResolveArea(AreaDamageRequest{
		Radius: 10,
		Spec: DamageSpec{
			Amount: 4,
			Type:   DamageTypeThermal,
		},
		Candidates: []DamageTarget{
			{EntityID: "player-1", EntityType: EntityTypePlayer, Health: 10},
			{EntityID: "asteroid-1", EntityType: EntityTypeAsteroid, Health: 6},
		},
	})

	if len(result.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result.Results))
	}
	if result.Results[0].TargetEntityID != "player-1" {
		t.Fatalf("expected first target %q, got %q", "player-1", result.Results[0].TargetEntityID)
	}
	if result.Results[1].TargetEntityID != "asteroid-1" {
		t.Fatalf("expected second target %q, got %q", "asteroid-1", result.Results[1].TargetEntityID)
	}
}

func TestResolveAreaTargetSpecificResistance(t *testing.T) {
	result := ResolveArea(AreaDamageRequest{
		Radius: 10,
		Spec: DamageSpec{
			Amount: 4,
			Type:   DamageTypeThermal,
		},
		Candidates: []DamageTarget{
			{
				EntityID:   "player-1",
				EntityType: EntityTypePlayer,
				Health:     10,
				Modifiers: []DamageModifier{
					{Type: DamageTypeThermal, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.5},
				},
			},
		},
	})

	if len(result.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Results))
	}
	if result.Results[0].ModifiedAmount != 2 {
		t.Fatalf("expected modified amount %d, got %d", 2, result.Results[0].ModifiedAmount)
	}
}

func TestResolveAreaTargetSpecificVulnerability(t *testing.T) {
	result := ResolveArea(AreaDamageRequest{
		Radius: 10,
		Spec: DamageSpec{
			Amount: 4,
			Type:   DamageTypeThermal,
		},
		Candidates: []DamageTarget{
			{
				EntityID:   "asteroid-1",
				EntityType: EntityTypeAsteroid,
				Health:     10,
				Modifiers: []DamageModifier{
					{Type: DamageTypeThermal, Operation: DamageModifierOperationMultiply, Value: 2},
				},
			},
		},
	})

	if len(result.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Results))
	}
	if result.Results[0].ModifiedAmount != 8 {
		t.Fatalf("expected modified amount %d, got %d", 8, result.Results[0].ModifiedAmount)
	}
}

func TestResolveAreaShieldHandlingPerTarget(t *testing.T) {
	result := ResolveArea(AreaDamageRequest{
		Radius: 10,
		Spec: DamageSpec{
			Amount: 5,
			Type:   DamageTypeThermal,
		},
		Candidates: []DamageTarget{
			{
				EntityID:   "player-1",
				EntityType: EntityTypePlayer,
				Health:     10,
				Shield:     3,
			},
			{
				EntityID:   "player-2",
				EntityType: EntityTypePlayer,
				Health:     10,
				Shield:     3,
				Modifiers: []DamageModifier{
					{Operation: DamageModifierOperationSet, Value: 7},
				},
			},
		},
	})

	if result.Results[0].AbsorbedByShield != 3 {
		t.Fatalf("expected first absorbed shield %d, got %d", 3, result.Results[0].AbsorbedByShield)
	}
	if result.Results[0].AppliedToHealth != 2 {
		t.Fatalf("expected first applied health %d, got %d", 2, result.Results[0].AppliedToHealth)
	}
	if result.Results[1].AbsorbedByShield != 3 {
		t.Fatalf("expected second absorbed shield %d, got %d", 3, result.Results[1].AbsorbedByShield)
	}
	if result.Results[1].AppliedToHealth != 4 {
		t.Fatalf("expected second applied health %d, got %d", 4, result.Results[1].AppliedToHealth)
	}
}

func TestAreaDamageRequestConstruction(t *testing.T) {
	req := AreaDamageRequest{
		Source: DamageSource{
			EntityID:   "hazard-1",
			EntityType: EntityTypeAsteroid,
			Cause:      DamageCauseArea,
		},
		OriginX: 12.5,
		OriginY: 34.5,
		Radius:  20.0,
		Spec: DamageSpec{
			Amount:       4,
			Type:         DamageTypeThermal,
			Cause:        DamageCauseArea,
			BypassShield: false,
		},
		Modifiers: []DamageModifier{
			{Category: DamageModifierCategoryGeneric, Operation: DamageModifierOperationAdd, Value: 1},
		},
		Candidates: []DamageTarget{
			{
				EntityID:   "player-1",
				EntityType: EntityTypePlayer,
				Health:     10,
				Shield:     2,
			},
		},
	}

	if req.Source.EntityID != "hazard-1" {
		t.Fatalf("expected source entity id %q, got %q", "hazard-1", req.Source.EntityID)
	}
	if req.OriginX != 12.5 || req.OriginY != 34.5 {
		t.Fatalf("expected origin (12.5, 34.5), got (%v, %v)", req.OriginX, req.OriginY)
	}
	if req.Radius != 20.0 {
		t.Fatalf("expected radius %v, got %v", 20.0, req.Radius)
	}
	if req.Spec.Amount != 4 {
		t.Fatalf("expected spec amount %d, got %d", 4, req.Spec.Amount)
	}
	if len(req.Modifiers) != 1 {
		t.Fatalf("expected 1 modifier, got %d", len(req.Modifiers))
	}
	if len(req.Candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(req.Candidates))
	}
}

func TestAreaDamageResultConstruction(t *testing.T) {
	result := AreaDamageResult{
		Results: []DamageResult{
			{
				TargetEntityID:  "player-1",
				TargetEntityType: EntityTypePlayer,
				BaseAmount:      4,
				ModifiedAmount:  5,
			},
		},
	}

	if len(result.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Results))
	}
	if result.Results[0].TargetEntityID != "player-1" {
		t.Fatalf("expected result target id %q, got %q", "player-1", result.Results[0].TargetEntityID)
	}
	if result.Results[0].BaseAmount != 4 {
		t.Fatalf("expected base amount %d, got %d", 4, result.Results[0].BaseAmount)
	}
	if result.Results[0].ModifiedAmount != 5 {
		t.Fatalf("expected modified amount %d, got %d", 5, result.Results[0].ModifiedAmount)
	}
}
