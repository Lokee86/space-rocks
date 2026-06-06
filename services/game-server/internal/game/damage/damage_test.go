package damage

import "testing"

func TestResolveSingleNoDotCreatesNoDoT(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Source: DamageSource{
			EntityID:   "hazard-1",
			EntityType: EntityTypeAsteroid,
			Cause:      DamageCauseArea,
		},
		Target: DamageTarget{
			EntityID:   "player-1",
			EntityType: EntityTypePlayer,
			Health:     10,
		},
		Spec: DamageSpec{
			Amount: 1,
			Type:   DamageTypeThermal,
		},
	})

	if len(result.CreatedDamageOverTime) != 0 {
		t.Fatalf("expected no created dot effects, got %d", len(result.CreatedDamageOverTime))
	}
}

func TestResolveSingleEnabledDotCreatesEffect(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Source: DamageSource{
			EntityID:   "hazard-1",
			EntityType: EntityTypeAsteroid,
			Cause:      DamageCauseArea,
		},
		Target: DamageTarget{
			EntityID:   "player-1",
			EntityType: EntityTypePlayer,
			Health:     10,
		},
		Spec: DamageSpec{
			Amount: 1,
			Type:   DamageTypeThermal,
			DoT: DamageOverTimeSpec{
				Enabled:         true,
				AmountPerTick:   2,
				TickSeconds:     0.5,
				DurationSeconds: 3.0,
				Type:            DamageTypeRadioactive,
				Modifiers: []DamageModifier{
					{Operation: DamageModifierOperationAdd, Value: 1},
				},
			},
		},
	})

	if len(result.CreatedDamageOverTime) != 1 {
		t.Fatalf("expected 1 created dot effect, got %d", len(result.CreatedDamageOverTime))
	}
	effect := result.CreatedDamageOverTime[0]
	if effect.Source.EntityID != "hazard-1" {
		t.Fatalf("expected source entity id %q, got %q", "hazard-1", effect.Source.EntityID)
	}
	if effect.Target.EntityID != "player-1" {
		t.Fatalf("expected target entity id %q, got %q", "player-1", effect.Target.EntityID)
	}
	if effect.Type != DamageTypeRadioactive {
		t.Fatalf("expected kind %q, got %q", DamageTypeRadioactive, effect.Type)
	}
	if effect.Source.Cause != DamageCauseDot {
		t.Fatalf("expected source cause %q, got %q", DamageCauseDot, effect.Source.Cause)
	}
	if effect.AmountPerTick != 2 {
		t.Fatalf("expected amount per tick %d, got %d", 2, effect.AmountPerTick)
	}
	if effect.TickSeconds != 0.5 {
		t.Fatalf("expected tick seconds %v, got %v", 0.5, effect.TickSeconds)
	}
	if effect.DurationSeconds != 3.0 {
		t.Fatalf("expected duration seconds %v, got %v", 3.0, effect.DurationSeconds)
	}
	if len(effect.Modifiers) != 1 {
		t.Fatalf("expected 1 modifier, got %d", len(effect.Modifiers))
	}
}

func TestResolveSingleIgnoredDamageCreatesNoDoT(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "player-1",
			EntityType: EntityTypePlayer,
			Health:     10,
		},
		Spec: DamageSpec{
			Amount: 0,
			Type:   DamageTypeThermal,
			DoT: DamageOverTimeSpec{
				Enabled:         true,
				AmountPerTick:   2,
				TickSeconds:     0.5,
				DurationSeconds: 3.0,
				Type:            DamageTypeRadioactive,
			},
		},
	})

	if !result.Ignored {
		t.Fatal("expected zero damage result to be ignored")
	}
	if len(result.CreatedDamageOverTime) != 0 {
		t.Fatalf("expected no created dot effects, got %d", len(result.CreatedDamageOverTime))
	}
}

func TestResolveSingleDotEffectFieldPreservation(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Source: DamageSource{
			EntityID:   "hazard-1",
			EntityType: EntityTypeAsteroid,
			Cause:      DamageCauseArea,
		},
		Target: DamageTarget{
			EntityID:   "player-1",
			EntityType: EntityTypePlayer,
			Health:     10,
		},
		Spec: DamageSpec{
			Amount: 1,
			Type:   DamageTypeThermal,
			DoT: DamageOverTimeSpec{
				Enabled:         true,
				AmountPerTick:   2,
				TickSeconds:     0.5,
				DurationSeconds: 3.0,
				Type:            DamageTypeRadioactive,
				Modifiers: []DamageModifier{
					{Operation: DamageModifierOperationAdd, Value: 1},
				},
			},
		},
	})

	effect := result.CreatedDamageOverTime[0]
	if effect.Source.EntityID != "hazard-1" || effect.Source.EntityType != EntityTypeAsteroid {
		t.Fatal("expected source ref to be preserved")
	}
	if effect.Target.EntityID != "player-1" || effect.Target.EntityType != EntityTypePlayer {
		t.Fatal("expected target ref to be preserved")
	}
	if effect.Type != DamageTypeRadioactive {
		t.Fatalf("expected kind %q, got %q", DamageTypeRadioactive, effect.Type)
	}
	if effect.AmountPerTick != 2 || effect.TickSeconds != 0.5 || effect.DurationSeconds != 3.0 {
		t.Fatal("expected tick fields to be preserved")
	}
	if len(effect.Modifiers) != 1 {
		t.Fatalf("expected 1 modifier, got %d", len(effect.Modifiers))
	}
}

func TestTickDamageOverTimeNoTickBeforeInterval(t *testing.T) {
	effect := ActiveDamageOverTime{
		Source: DamageSource{EntityID: "hazard-1", EntityType: EntityTypeAsteroid, Cause: DamageCauseArea},
		Target: DamageTargetRef{EntityID: "player-1", EntityType: EntityTypePlayer},
		AmountPerTick: 1,
		TickSeconds: 1,
		DurationSeconds: 5,
		Type: DamageTypeRadioactive,
	}

	result := TickDamageOverTime(effect, DamageTarget{EntityID: "player-1", EntityType: EntityTypePlayer, Health: 10}, 0.5)

	if len(result.Results) != 0 {
		t.Fatalf("expected 0 tick results, got %d", len(result.Results))
	}
	if result.TickRemaining != 0.5 {
		t.Fatalf("expected tick remaining %v, got %v", 0.5, result.TickRemaining)
	}
}

func TestTickDamageOverTimeOneTickAtInterval(t *testing.T) {
	effect := ActiveDamageOverTime{
		Source: DamageSource{EntityID: "hazard-1", EntityType: EntityTypeAsteroid, Cause: DamageCauseArea},
		Target: DamageTargetRef{EntityID: "player-1", EntityType: EntityTypePlayer},
		AmountPerTick: 2,
		TickSeconds: 1,
		DurationSeconds: 5,
		Type: DamageTypeRadioactive,
	}

	result := TickDamageOverTime(effect, DamageTarget{EntityID: "player-1", EntityType: EntityTypePlayer, Health: 10}, 1)

	if len(result.Results) != 1 {
		t.Fatalf("expected 1 tick result, got %d", len(result.Results))
	}
	if result.Results[0].ModifiedAmount != 2 {
		t.Fatalf("expected modified amount %d, got %d", 2, result.Results[0].ModifiedAmount)
	}
	if result.TickRemaining != 1 {
		t.Fatalf("expected tick remaining %v, got %v", 1, result.TickRemaining)
	}
}

func TestTickDamageOverTimeMultipleTicksForLargeDelta(t *testing.T) {
	effect := ActiveDamageOverTime{
		Source: DamageSource{EntityID: "hazard-1", EntityType: EntityTypeAsteroid, Cause: DamageCauseArea},
		Target: DamageTargetRef{EntityID: "player-1", EntityType: EntityTypePlayer},
		AmountPerTick: 2,
		TickSeconds: 1,
		DurationSeconds: 5,
		Type: DamageTypeRadioactive,
	}

	result := TickDamageOverTime(effect, DamageTarget{EntityID: "player-1", EntityType: EntityTypePlayer, Health: 10}, 2.5)

	if len(result.Results) != 2 {
		t.Fatalf("expected 2 tick results, got %d", len(result.Results))
	}
	if result.DurationRemaining != 2.5 {
		t.Fatalf("expected duration remaining %v, got %v", 2.5, result.DurationRemaining)
	}
	if result.TickRemaining != 0.5 {
		t.Fatalf("expected tick remaining %v, got %v", 0.5, result.TickRemaining)
	}
}

func TestTickDamageOverTimeResistanceAffectsTickDamage(t *testing.T) {
	effect := ActiveDamageOverTime{
		Source: DamageSource{EntityID: "hazard-1", EntityType: EntityTypeAsteroid, Cause: DamageCauseArea},
		Target: DamageTargetRef{EntityID: "player-1", EntityType: EntityTypePlayer},
		AmountPerTick: 4,
		TickSeconds: 1,
		DurationSeconds: 5,
		Type: DamageTypeThermal,
		Modifiers: []DamageModifier{
			{Type: DamageTypeThermal, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.5},
		},
	}

	result := TickDamageOverTime(effect, DamageTarget{EntityID: "player-1", EntityType: EntityTypePlayer, Health: 10}, 1)

	if len(result.Results) != 1 {
		t.Fatalf("expected 1 tick result, got %d", len(result.Results))
	}
	if result.Results[0].ModifiedAmount != 2 {
		t.Fatalf("expected modified amount %d, got %d", 2, result.Results[0].ModifiedAmount)
	}
}

func TestTickDamageOverTimeExpiredEffect(t *testing.T) {
	effect := ActiveDamageOverTime{
		Source: DamageSource{EntityID: "hazard-1", EntityType: EntityTypeAsteroid, Cause: DamageCauseArea},
		Target: DamageTargetRef{EntityID: "player-1", EntityType: EntityTypePlayer},
		AmountPerTick: 2,
		TickSeconds: 1,
		DurationSeconds: 1,
		Type: DamageTypeRadioactive,
	}

	result := TickDamageOverTime(effect, DamageTarget{EntityID: "player-1", EntityType: EntityTypePlayer, Health: 10}, 1.5)

	if !result.Expired {
		t.Fatal("expected effect to be expired")
	}
	if result.DurationRemaining != 0 {
		t.Fatalf("expected duration remaining %v, got %v", 0, result.DurationRemaining)
	}
}

func TestActiveDamageOverTimeConstruction(t *testing.T) {
	active := ActiveDamageOverTime{
		Source: DamageSource{
			EntityID:   "hazard-1",
			EntityType: EntityTypeAsteroid,
			Cause:      DamageCauseArea,
		},
		Target: DamageTargetRef{
			EntityID:   "player-1",
			EntityType: EntityTypePlayer,
		},
		AmountPerTick:  2,
		TickSeconds:    0.5,
		DurationSeconds: 3.0,
		Type:           DamageTypeRadioactive,
		Modifiers: []DamageModifier{
			{Operation: DamageModifierOperationAdd, Value: 1},
		},
	}

	if active.Source.EntityID != "hazard-1" {
		t.Fatalf("expected source entity id %q, got %q", "hazard-1", active.Source.EntityID)
	}
	if active.Target.EntityID != "player-1" {
		t.Fatalf("expected target entity id %q, got %q", "player-1", active.Target.EntityID)
	}
	if active.AmountPerTick != 2 {
		t.Fatalf("expected amount per tick %d, got %d", 2, active.AmountPerTick)
	}
	if active.TickSeconds != 0.5 {
		t.Fatalf("expected tick seconds %v, got %v", 0.5, active.TickSeconds)
	}
	if active.DurationSeconds != 3.0 {
		t.Fatalf("expected duration seconds %v, got %v", 3.0, active.DurationSeconds)
	}
	if active.Type != DamageTypeRadioactive {
		t.Fatalf("expected kind %q, got %q", DamageTypeRadioactive, active.Type)
	}
	if len(active.Modifiers) != 1 {
		t.Fatalf("expected 1 modifier, got %d", len(active.Modifiers))
	}
}

func TestDamageOverTimeTickResultConstruction(t *testing.T) {
	result := DamageOverTimeTickResult{
		Source: DamageSource{
			EntityID:   "hazard-1",
			EntityType: EntityTypeAsteroid,
			Cause:      DamageCauseArea,
		},
		Target: DamageTargetRef{
			EntityID:   "player-1",
			EntityType: EntityTypePlayer,
		},
		AmountPerTick:  2,
		TickSeconds:    0.5,
		DurationSeconds: 3.0,
		Type:           DamageTypeRadioactive,
		Modifiers: []DamageModifier{
			{Operation: DamageModifierOperationAdd, Value: 1},
		},
	}

	if result.Source.EntityID != "hazard-1" {
		t.Fatalf("expected source entity id %q, got %q", "hazard-1", result.Source.EntityID)
	}
	if result.Target.EntityID != "player-1" {
		t.Fatalf("expected target entity id %q, got %q", "player-1", result.Target.EntityID)
	}
	if result.AmountPerTick != 2 {
		t.Fatalf("expected amount per tick %d, got %d", 2, result.AmountPerTick)
	}
	if result.TickSeconds != 0.5 {
		t.Fatalf("expected tick seconds %v, got %v", 0.5, result.TickSeconds)
	}
	if result.DurationSeconds != 3.0 {
		t.Fatalf("expected duration seconds %v, got %v", 3.0, result.DurationSeconds)
	}
	if result.Type != DamageTypeRadioactive {
		t.Fatalf("expected kind %q, got %q", DamageTypeRadioactive, result.Type)
	}
	if len(result.Modifiers) != 1 {
		t.Fatalf("expected 1 modifier, got %d", len(result.Modifiers))
	}
}

func TestDamageSpecConstructionWithDot(t *testing.T) {
	spec := DamageSpec{
		Amount: 5,
		Type:   DamageTypeThermal,
		DoT: DamageOverTimeSpec{
			Enabled:         true,
			AmountPerTick:   1,
			TickSeconds:     0.5,
			DurationSeconds: 3.0,
			Type:            DamageTypeRadioactive,
			Modifiers: []DamageModifier{
				{Operation: DamageModifierOperationAdd, Value: 1},
			},
		},
	}

	if !spec.DoT.Enabled {
		t.Fatal("expected dot to be enabled")
	}
	if spec.DoT.AmountPerTick != 1 {
		t.Fatalf("expected amount per tick %d, got %d", 1, spec.DoT.AmountPerTick)
	}
	if spec.DoT.TickSeconds != 0.5 {
		t.Fatalf("expected tick seconds %v, got %v", 0.5, spec.DoT.TickSeconds)
	}
	if spec.DoT.DurationSeconds != 3.0 {
		t.Fatalf("expected duration seconds %v, got %v", 3.0, spec.DoT.DurationSeconds)
	}
	if spec.DoT.Type != DamageTypeRadioactive {
		t.Fatalf("expected dot kind %q, got %q", DamageTypeRadioactive, spec.DoT.Type)
	}
	if len(spec.DoT.Modifiers) != 1 {
		t.Fatalf("expected 1 dot modifier, got %d", len(spec.DoT.Modifiers))
	}
}

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

func TestResolveSingleHealthOneDamageOneDestroysTarget(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "asteroid-1",
			EntityType: EntityTypeAsteroid,
			Health:     1,
		},
		Spec: DamageSpec{
			Amount: 1,
			Type:   DamageTypeKinetic,
		},
	})

	if result.BaseAmount != 1 {
		t.Fatalf("expected base amount %d, got %d", 1, result.BaseAmount)
	}
	if result.ModifiedAmount != 1 {
		t.Fatalf("expected modified amount %d, got %d", 1, result.ModifiedAmount)
	}
	if result.AppliedToHealth != 1 {
		t.Fatalf("expected applied health damage 1, got %d", result.AppliedToHealth)
	}
	if result.AbsorbedByShield != 0 {
		t.Fatalf("expected absorbed shield 0, got %d", result.AbsorbedByShield)
	}
	if result.RemainingHealth != 0 {
		t.Fatalf("expected remaining health 0, got %d", result.RemainingHealth)
	}
	if !result.Destroyed {
		t.Fatal("expected target to be destroyed")
	}
}

func TestResolveSinglePlayerDestroyedIsFatal(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "player-1",
			EntityType: EntityTypePlayer,
			Health:     1,
		},
		Spec: DamageSpec{
			Amount: 1,
			Type:   DamageTypeKinetic,
		},
	})

	if !result.Destroyed {
		t.Fatal("expected target to be destroyed")
	}
	if !result.Fatal {
		t.Fatal("expected destroyed player result to be fatal")
	}
}

func TestResolveSingleModifiedAmountUsesModifiers(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "asteroid-1",
			EntityType: EntityTypeAsteroid,
			Health:     10,
		},
		Spec: DamageSpec{
			Amount: 5,
			Type:   DamageTypeThermal,
		},
		Modifiers: []DamageModifier{
			{Operation: DamageModifierOperationAdd, Value: 3},
			{Operation: DamageModifierOperationMultiply, Value: 2},
		},
	})

	if result.BaseAmount != 5 {
		t.Fatalf("expected base amount %d, got %d", 5, result.BaseAmount)
	}
	if result.ModifiedAmount != 16 {
		t.Fatalf("expected modified amount %d, got %d", 16, result.ModifiedAmount)
	}
	if result.Type != DamageTypeThermal {
		t.Fatalf("expected kind %q, got %q", DamageTypeThermal, result.Type)
	}
	if result.Cause != "" {
		t.Fatalf("expected empty cause, got %q", result.Cause)
	}
	if result.AppliedToHealth != 10 {
		t.Fatalf("expected applied health damage 10, got %d", result.AppliedToHealth)
	}
	if result.RemainingHealth != 0 {
		t.Fatalf("expected remaining health 0, got %d", result.RemainingHealth)
	}
	if !result.Destroyed {
		t.Fatal("expected target to be destroyed")
	}
}

func TestResolveSingleFullShieldAbsorption(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "asteroid-1",
			EntityType: EntityTypeAsteroid,
			Health:     10,
			Shield:     6,
		},
		Spec: DamageSpec{
			Amount: 5,
			Type:   DamageTypeThermal,
		},
	})

	if result.AbsorbedByShield != 5 {
		t.Fatalf("expected absorbed shield 5, got %d", result.AbsorbedByShield)
	}
	if result.AppliedToHealth != 0 {
		t.Fatalf("expected applied health 0, got %d", result.AppliedToHealth)
	}
	if result.RemainingShield != 1 {
		t.Fatalf("expected remaining shield 1, got %d", result.RemainingShield)
	}
	if result.RemainingHealth != 10 {
		t.Fatalf("expected remaining health 10, got %d", result.RemainingHealth)
	}
	if result.Destroyed {
		t.Fatal("expected target not to be destroyed")
	}
}

func TestResolveSinglePartialShieldAbsorption(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "asteroid-1",
			EntityType: EntityTypeAsteroid,
			Health:     10,
			Shield:     2,
		},
		Spec: DamageSpec{
			Amount: 5,
			Type:   DamageTypeThermal,
		},
	})

	if result.AbsorbedByShield != 2 {
		t.Fatalf("expected absorbed shield 2, got %d", result.AbsorbedByShield)
	}
	if result.AppliedToHealth != 3 {
		t.Fatalf("expected applied health 3, got %d", result.AppliedToHealth)
	}
	if result.RemainingShield != 0 {
		t.Fatalf("expected remaining shield 0, got %d", result.RemainingShield)
	}
	if result.RemainingHealth != 7 {
		t.Fatalf("expected remaining health 7, got %d", result.RemainingHealth)
	}
}

func TestResolveSingleBypassShield(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "asteroid-1",
			EntityType: EntityTypeAsteroid,
			Health:     10,
			Shield:     4,
		},
		Spec: DamageSpec{
			Amount:       5,
			Type:         DamageTypeThermal,
			BypassShield: true,
		},
	})

	if result.AbsorbedByShield != 0 {
		t.Fatalf("expected absorbed shield 0, got %d", result.AbsorbedByShield)
	}
	if result.AppliedToHealth != 5 {
		t.Fatalf("expected applied health 5, got %d", result.AppliedToHealth)
	}
	if result.RemainingShield != 4 {
		t.Fatalf("expected remaining shield 4, got %d", result.RemainingShield)
	}
	if result.RemainingHealth != 5 {
		t.Fatalf("expected remaining health 5, got %d", result.RemainingHealth)
	}
}

func TestResolveSingleZeroDamageIgnored(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "asteroid-1",
			EntityType: EntityTypeAsteroid,
			Health:     10,
			Shield:     4,
		},
		Spec: DamageSpec{
			Amount: 0,
			Type:   DamageTypeThermal,
		},
	})

	if !result.Ignored {
		t.Fatal("expected zero damage result to be ignored")
	}
}

func TestResolveSingleDeadTargetIgnored(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "asteroid-1",
			EntityType: EntityTypeAsteroid,
			Health:     0,
			Shield:     4,
		},
		Spec: DamageSpec{
			Amount: 5,
			Type:   DamageTypeThermal,
		},
	})

	if !result.Ignored {
		t.Fatal("expected dead target result to be ignored")
	}
}

func TestResolveSingleFatalPlayer(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "player-1",
			EntityType: EntityTypePlayer,
			Health:     1,
			Shield:     0,
		},
		Spec: DamageSpec{
			Amount: 1,
			Type:   DamageTypeThermal,
		},
	})

	if !result.Destroyed {
		t.Fatal("expected target to be destroyed")
	}
	if !result.Fatal {
		t.Fatal("expected destroyed player result to be fatal")
	}
}

func TestResolveSingleNonfatalAsteroidDestruction(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "asteroid-1",
			EntityType: EntityTypeAsteroid,
			Health:     1,
			Shield:     0,
		},
		Spec: DamageSpec{
			Amount: 1,
			Type:   DamageTypeThermal,
		},
	})

	if !result.Destroyed {
		t.Fatal("expected target to be destroyed")
	}
	if result.Fatal {
		t.Fatal("expected destroyed asteroid result not to be fatal")
	}
}

func TestResolveSingleReportsAppliedResistanceModifier(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "asteroid-1",
			EntityType: EntityTypeAsteroid,
			Health:     10,
		},
		Spec: DamageSpec{
			Amount: 5,
			Type:   DamageTypeThermal,
		},
		Modifiers: []DamageModifier{
			{Type: DamageTypeThermal, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.5},
		},
	})

	if len(result.AppliedModifiers) != 1 {
		t.Fatalf("expected 1 applied modifier, got %d", len(result.AppliedModifiers))
	}
	if result.AppliedModifiers[0].Modifier.Category != DamageModifierCategoryResistance {
		t.Fatalf("expected resistance modifier, got %q", result.AppliedModifiers[0].Modifier.Category)
	}
}

func TestResolveSingleReportsAppliedVulnerabilityModifier(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "asteroid-1",
			EntityType: EntityTypeAsteroid,
			Health:     10,
		},
		Spec: DamageSpec{
			Amount: 5,
			Type:   DamageTypeThermal,
		},
		Modifiers: []DamageModifier{
			{Type: DamageTypeThermal, Category: DamageModifierCategoryVulnerability, Operation: DamageModifierOperationMultiply, Value: 2},
		},
	})

	if len(result.AppliedModifiers) != 1 {
		t.Fatalf("expected 1 applied modifier, got %d", len(result.AppliedModifiers))
	}
	if result.AppliedModifiers[0].Modifier.Category != DamageModifierCategoryVulnerability {
		t.Fatalf("expected vulnerability modifier, got %q", result.AppliedModifiers[0].Modifier.Category)
	}
}

func TestDamageResolutionRequestConstruction(t *testing.T) {
	req := DamageResolutionRequest{
		Source: DamageSource{
			EntityID:   "bullet-1",
			EntityType: EntityTypeProjectile,
			Cause:      DamageCauseProjectile,
		},
		Target: DamageTarget{
			EntityID:   "asteroid-1",
			EntityType: EntityTypeAsteroid,
			Health:     3,
			Shield:     1,
		},
		Spec: DamageSpec{
			Amount:       2,
			Type:         DamageTypeExplosive,
			Cause:        DamageCauseProjectile,
			BypassShield: true,
		},
		Modifiers: []DamageModifier{
			{Type: DamageTypeExplosive, Category: DamageModifierCategoryOutgoing, Operation: DamageModifierOperationAdd, Value: 1},
		},
	}

	if req.Source.EntityID != "bullet-1" {
		t.Fatalf("expected source entity id %q, got %q", "bullet-1", req.Source.EntityID)
	}
	if req.Target.EntityID != "asteroid-1" {
		t.Fatalf("expected target entity id %q, got %q", "asteroid-1", req.Target.EntityID)
	}
	if req.Spec.Amount != 2 {
		t.Fatalf("expected spec amount %d, got %d", 2, req.Spec.Amount)
	}
	if len(req.Modifiers) != 1 {
		t.Fatalf("expected 1 modifier, got %d", len(req.Modifiers))
	}
}

func TestResolveModifiedAmountNoModifiers(t *testing.T) {
	result := ResolveModifiedAmount(5, nil, DamageTypeThermal)

	if result.BaseAmount != 5 {
		t.Fatalf("expected base amount %v, got %v", 5, result.BaseAmount)
	}
	if result.ModifiedAmount != 5 {
		t.Fatalf("expected modified amount %d, got %d", 5, result.ModifiedAmount)
	}
	if len(result.AppliedModifiers) != 0 {
		t.Fatalf("expected no applied modifiers, got %d", len(result.AppliedModifiers))
	}
}

func TestResolveModifiedAmountAdd(t *testing.T) {
	result := ResolveModifiedAmount(5, []DamageModifier{
		{Operation: DamageModifierOperationAdd, Value: 2},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 7 {
		t.Fatalf("expected modified amount %d, got %d", 7, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountMultiply(t *testing.T) {
	result := ResolveModifiedAmount(5, []DamageModifier{
		{Operation: DamageModifierOperationMultiply, Value: 2},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 10 {
		t.Fatalf("expected modified amount %d, got %d", 10, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountthermalResistance025(t *testing.T) {
	result := ResolveModifiedAmount(100, []DamageModifier{
		{Type: DamageTypeThermal, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.25},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 75 {
		t.Fatalf("expected modified amount %d, got %d", 75, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountthermalResistance025And020(t *testing.T) {
	result := ResolveModifiedAmount(100, []DamageModifier{
		{Type: DamageTypeThermal, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.25},
		{Type: DamageTypeThermal, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.20},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 60 {
		t.Fatalf("expected modified amount %d, got %d", 60, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountthermalVulnerability125(t *testing.T) {
	result := ResolveModifiedAmount(100, []DamageModifier{
		{Type: DamageTypeThermal, Category: DamageModifierCategoryVulnerability, Operation: DamageModifierOperationMultiply, Value: 1.25},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 125 {
		t.Fatalf("expected modified amount %d, got %d", 125, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountthermalResistanceDoesNotAffectradioactive(t *testing.T) {
	result := ResolveModifiedAmount(100, []DamageModifier{
		{Type: DamageTypeThermal, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.25},
	}, DamageTypeRadioactive)

	if result.ModifiedAmount != 100 {
		t.Fatalf("expected modified amount %d, got %d", 100, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountthermalVulnerabilityDoesNotAffectradioactive(t *testing.T) {
	result := ResolveModifiedAmount(100, []DamageModifier{
		{Type: DamageTypeThermal, Category: DamageModifierCategoryVulnerability, Operation: DamageModifierOperationMultiply, Value: 1.25},
	}, DamageTypeRadioactive)

	if result.ModifiedAmount != 100 {
		t.Fatalf("expected modified amount %d, got %d", 100, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountEmptyKindResistanceAppliesGlobally(t *testing.T) {
	result := ResolveModifiedAmount(100, []DamageModifier{
		{Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.25},
	}, DamageTypeRadioactive)

	if result.ModifiedAmount != 75 {
		t.Fatalf("expected modified amount %d, got %d", 75, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountRadioactiveIgnoresThermalResistance(t *testing.T) {
	result := ResolveModifiedAmount(100, []DamageModifier{
		{Type: DamageTypeThermal, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.25},
	}, DamageTypeRadioactive)

	if result.ModifiedAmount != 100 {
		t.Fatalf("expected modified amount %d, got %d", 100, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountEmptyTypeResistanceAppliesGlobally(t *testing.T) {
	result := ResolveModifiedAmount(100, []DamageModifier{
		{Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.25},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 75 {
		t.Fatalf("expected modified amount %d, got %d", 75, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountMultiplyHalf(t *testing.T) {
	result := ResolveModifiedAmount(5, []DamageModifier{
		{Operation: DamageModifierOperationMultiply, Value: 0.5},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 3 {
		t.Fatalf("expected modified amount %d, got %d", 3, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountMultiplyOnePointFive(t *testing.T) {
	result := ResolveModifiedAmount(5, []DamageModifier{
		{Operation: DamageModifierOperationMultiply, Value: 1.5},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 8 {
		t.Fatalf("expected modified amount %d, got %d", 8, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountAddBeforeMultiply(t *testing.T) {
	result := ResolveModifiedAmount(5, []DamageModifier{
		{Operation: DamageModifierOperationAdd, Value: 3},
		{Operation: DamageModifierOperationMultiply, Value: 2},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 16 {
		t.Fatalf("expected modified amount %d, got %d", 16, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountSetLast(t *testing.T) {
	result := ResolveModifiedAmount(5, []DamageModifier{
		{Operation: DamageModifierOperationAdd, Value: 3},
		{Operation: DamageModifierOperationMultiply, Value: 2},
		{Operation: DamageModifierOperationSet, Value: 4},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 4 {
		t.Fatalf("expected modified amount %d, got %d", 4, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountWrongKindIgnored(t *testing.T) {
	result := ResolveModifiedAmount(5, []DamageModifier{
		{Type: DamageTypeRadioactive, Operation: DamageModifierOperationAdd, Value: 3},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 5 {
		t.Fatalf("expected modified amount %d, got %d", 5, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountZeroMultiplier(t *testing.T) {
	result := ResolveModifiedAmount(5, []DamageModifier{
		{Operation: DamageModifierOperationMultiply, Value: 0},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 0 {
		t.Fatalf("expected modified amount %d, got %d", 0, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountNegativeClamp(t *testing.T) {
	result := ResolveModifiedAmount(5, []DamageModifier{
		{Operation: DamageModifierOperationAdd, Value: -10},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 0 {
		t.Fatalf("expected modified amount %d, got %d", 0, result.ModifiedAmount)
	}
}

func TestFilterDamageModifiersByKindEmptyKindApplies(t *testing.T) {
	modifiers := []DamageModifier{
		{Category: DamageModifierCategoryGeneric, Operation: DamageModifierOperationAdd, Value: 1},
	}

	filtered := FilterDamageModifiersByKind(modifiers, DamageTypeThermal)

	if len(filtered) != 1 {
		t.Fatalf("expected 1 modifier, got %d", len(filtered))
	}
	if filtered[0] != modifiers[0] {
		t.Fatal("expected empty kind modifier to be preserved")
	}
}

func TestFilterDamageModifiersByKindMatchingKindApplies(t *testing.T) {
	modifiers := []DamageModifier{
		{Type: DamageTypeThermal, Category: DamageModifierCategoryOutgoing, Operation: DamageModifierOperationMultiply, Value: 2},
	}

	filtered := FilterDamageModifiersByKind(modifiers, DamageTypeThermal)

	if len(filtered) != 1 {
		t.Fatalf("expected 1 modifier, got %d", len(filtered))
	}
	if filtered[0] != modifiers[0] {
		t.Fatal("expected matching kind modifier to be preserved")
	}
}

func TestFilterDamageModifiersByKindNonMatchingKindDoesNotApply(t *testing.T) {
	modifiers := []DamageModifier{
		{Type: DamageTypeRadioactive, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.5},
	}

	filtered := FilterDamageModifiersByKind(modifiers, DamageTypeThermal)

	if len(filtered) != 0 {
		t.Fatalf("expected 0 modifiers, got %d", len(filtered))
	}
}

func TestFilterDamageModifiersByKindPreservesInputOrder(t *testing.T) {
	modifiers := []DamageModifier{
		{Category: DamageModifierCategoryGeneric, Operation: DamageModifierOperationAdd, Value: 1},
		{Type: DamageTypeThermal, Category: DamageModifierCategoryOutgoing, Operation: DamageModifierOperationMultiply, Value: 2},
		{Type: DamageTypeThermal, Category: DamageModifierCategoryGeneric, Operation: DamageModifierOperationSet, Value: 3},
		{Type: DamageTypeRadioactive, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.5},
	}

	filtered := FilterDamageModifiersByKind(modifiers, DamageTypeThermal)

	if len(filtered) != 3 {
		t.Fatalf("expected 3 modifiers, got %d", len(filtered))
	}
	if filtered[0] != modifiers[0] || filtered[1] != modifiers[1] || filtered[2] != modifiers[2] {
		t.Fatal("expected filtered modifiers to preserve input order")
	}
}

func TestFilterDamageModifiersByKindSkipsInvalidResistanceModifiers(t *testing.T) {
	modifiers := []DamageModifier{
		{Type: DamageTypeThermal, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 1.0},
		{Type: DamageTypeThermal, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 1.25},
	}

	filtered := FilterDamageModifiersByKind(modifiers, DamageTypeThermal)

	if len(filtered) != 0 {
		t.Fatalf("expected 0 modifiers, got %d", len(filtered))
	}
}

func TestFilterDamageModifiersByKindSkipsInvalidVulnerabilityModifiers(t *testing.T) {
	modifiers := []DamageModifier{
		{Type: DamageTypeThermal, Category: DamageModifierCategoryVulnerability, Operation: DamageModifierOperationMultiply, Value: 1.0},
		{Type: DamageTypeThermal, Category: DamageModifierCategoryVulnerability, Operation: DamageModifierOperationMultiply, Value: 0.75},
	}

	filtered := FilterDamageModifiersByKind(modifiers, DamageTypeThermal)

	if len(filtered) != 0 {
		t.Fatalf("expected 0 modifiers, got %d", len(filtered))
	}
}

func TestFilterDamageModifiersByKindOutgoingValidation(t *testing.T) {
	modifiers := []DamageModifier{
		{Type: DamageTypeThermal, Category: "", Operation: DamageModifierOperationAdd, Value: 1},
		{Type: DamageTypeThermal, Category: DamageModifierCategoryGeneric, Operation: DamageModifierOperationMultiply, Value: 2},
		{Type: DamageTypeThermal, Category: DamageModifierCategoryOutgoing, Operation: DamageModifierOperationSet, Value: 3},
	}

	filtered := FilterDamageModifiersByKind(modifiers, DamageTypeThermal)

	if len(filtered) != len(modifiers) {
		t.Fatalf("expected %d modifiers, got %d", len(modifiers), len(filtered))
	}
}

func TestDamageModifierCategoryStringValues(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{name: "outgoing", got: string(DamageModifierCategoryOutgoing), want: "outgoing"},
		{name: "resistance", got: string(DamageModifierCategoryResistance), want: "resistance"},
		{name: "vulnerability", got: string(DamageModifierCategoryVulnerability), want: "vulnerability"},
		{name: "generic", got: string(DamageModifierCategoryGeneric), want: "generic"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, tc.got)
			}
		})
	}
}

func TestDamageModifierOperationStringValues(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{name: "add", got: string(DamageModifierOperationAdd), want: "add"},
		{name: "multiply", got: string(DamageModifierOperationMultiply), want: "multiply"},
		{name: "set", got: string(DamageModifierOperationSet), want: "set"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, tc.got)
			}
		})
	}
}

func TestDamageSpecConstructionKineticCollision(t *testing.T) {
	spec := DamageSpec{
		Amount:       1,
		Type:         DamageTypeKinetic,
		Cause:        DamageCauseCollision,
		BypassShield: false,
	}

	if spec.Amount != 1 {
		t.Fatalf("expected amount %d, got %d", 1, spec.Amount)
	}
	if spec.Type != DamageTypeKinetic {
		t.Fatalf("expected kind %q, got %q", DamageTypeKinetic, spec.Type)
	}
	if spec.Cause != DamageCauseCollision {
		t.Fatalf("expected cause %q, got %q", DamageCauseCollision, spec.Cause)
	}
	if spec.BypassShield {
		t.Fatal("expected bypass shield to be false")
	}
}

func TestDamageSpecConstructionExplosiveProjectile(t *testing.T) {
	spec := DamageSpec{
		Amount:       3,
		Type:         DamageTypeExplosive,
		Cause:        DamageCauseProjectile,
		BypassShield: true,
	}

	if spec.Amount != 3 {
		t.Fatalf("expected amount %d, got %d", 3, spec.Amount)
	}
	if spec.Type != DamageTypeExplosive {
		t.Fatalf("expected kind %q, got %q", DamageTypeExplosive, spec.Type)
	}
	if spec.Cause != DamageCauseProjectile {
		t.Fatalf("expected cause %q, got %q", DamageCauseProjectile, spec.Cause)
	}
	if !spec.BypassShield {
		t.Fatal("expected bypass shield to be true")
	}
}

func TestDamageSourceConstruction(t *testing.T) {
	source := DamageSource{
		EntityID:   "bullet-1",
		EntityType: EntityTypeProjectile,
		Cause:      DamageCauseProjectile,
	}

	if source.EntityID != "bullet-1" {
		t.Fatalf("expected entity id %q, got %q", "bullet-1", source.EntityID)
	}
	if source.EntityType != EntityTypeProjectile {
		t.Fatalf("expected entity type %q, got %q", EntityTypeProjectile, source.EntityType)
	}
	if source.Cause != DamageCauseProjectile {
		t.Fatalf("expected cause %q, got %q", DamageCauseProjectile, source.Cause)
	}
}

func TestDamageTargetConstruction(t *testing.T) {
	target := DamageTarget{
		EntityID:   "player-1",
		EntityType: EntityTypePlayer,
		Health:     3,
		Shield:     1,
	}

	if target.EntityID != "player-1" {
		t.Fatalf("expected entity id %q, got %q", "player-1", target.EntityID)
	}
	if target.EntityType != EntityTypePlayer {
		t.Fatalf("expected entity type %q, got %q", EntityTypePlayer, target.EntityType)
	}
	if target.Health != 3 {
		t.Fatalf("expected health %d, got %d", 3, target.Health)
	}
	if target.Shield != 1 {
		t.Fatalf("expected shield %d, got %d", 1, target.Shield)
	}
}

func TestDamageCauseStringValues(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{name: "collision", got: string(DamageCauseCollision), want: "collision"},
		{name: "projectile", got: string(DamageCauseProjectile), want: "projectile"},
		{name: "debug", got: string(DamageCauseDebug), want: "debug"},
		{name: "area", got: string(DamageCauseArea), want: "area"},
		{name: "dot", got: string(DamageCauseDot), want: "dot"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, tc.got)
			}
		})
	}
}

func TestDamageTypeStringValues(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{name: "kinetic", got: string(DamageTypeKinetic), want: "kinetic"},
		{name: "explosive", got: string(DamageTypeExplosive), want: "explosive"},
		{name: "energy", got: string(DamageTypeEnergy), want: "energy"},
		{name: "thermal", got: string(DamageTypeThermal), want: "thermal"},
		{name: "radioactive", got: string(DamageTypeRadioactive), want: "radioactive"},
		{name: "true_damage", got: string(DamageTypeTrueDamage), want: "true_damage"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, tc.got)
			}
		})
	}
}



