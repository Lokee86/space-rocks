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
