package damage

func ResolveSingle(req DamageResolutionRequest) DamageResult {
	modifiers := make([]DamageModifier, 0, len(req.Modifiers)+len(req.Target.Modifiers))
	modifiers = append(modifiers, req.Modifiers...)
	modifiers = append(modifiers, req.Target.Modifiers...)
	modified := ResolveModifiedAmount(req.Spec.Amount, modifiers, req.Spec.Type)
	result := DamageResult{
		TargetEntityID:   req.Target.EntityID,
		TargetEntityType: req.Target.EntityType,
		BaseAmount:       int(modified.BaseAmount),
		ModifiedAmount:   modified.ModifiedAmount,
		Type:             req.Spec.Type,
		Cause:            req.Spec.Cause,
		AppliedModifiers: modified.AppliedModifiers,
	}

	if req.Spec.DoT.Enabled {
		result.CreatedDamageOverTime = []ActiveDamageOverTime{
			{
				Source: DamageSource{
					EntityID:   req.Source.EntityID,
					EntityType: req.Source.EntityType,
					Cause:      DamageCauseDot,
				},
				Target: DamageTargetRef{
					EntityID:   req.Target.EntityID,
					EntityType: req.Target.EntityType,
				},
				AmountPerTick:   req.Spec.DoT.AmountPerTick,
				TickSeconds:     req.Spec.DoT.TickSeconds,
				DurationSeconds: req.Spec.DoT.DurationSeconds,
				Type:            req.Spec.DoT.Type,
				Modifiers:       req.Spec.DoT.Modifiers,
			},
		}
	}

	if modified.ModifiedAmount <= 0 {
		result.Ignored = true
		result.CreatedDamageOverTime = nil
		return result
	}

	if req.Target.Health <= 0 {
		result.Ignored = true
		result.CreatedDamageOverTime = nil
		return result
	}

	damageToApply := modified.ModifiedAmount
	if !req.Spec.BypassShield && req.Target.Shield > 0 {
		absorbed := min(req.Target.Shield, damageToApply)
		result.AbsorbedByShield = absorbed
		result.RemainingShield = req.Target.Shield - absorbed
		damageToApply -= absorbed
	} else {
		result.RemainingShield = req.Target.Shield
	}

	if damageToApply > 0 {
		remaining := max(req.Target.Health-damageToApply, 0)
		result.AppliedToHealth = req.Target.Health - remaining
		result.RemainingHealth = remaining
	} else {
		result.RemainingHealth = req.Target.Health
	}

	result.Destroyed = result.RemainingHealth <= 0
	result.Fatal = result.Destroyed && req.Target.EntityType == EntityTypePlayer

	return result
}
