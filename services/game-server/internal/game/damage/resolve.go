package damage

func ResolveSingle(req DamageResolutionRequest) DamageResult {
	modifiers := make([]DamageModifier, 0, len(req.Modifiers)+len(req.Target.Modifiers))
	modifiers = append(modifiers, req.Modifiers...)
	modifiers = append(modifiers, req.Target.Modifiers...)
	modified := ResolveModifiedAmount(req.Spec.Amount, modifiers, req.Spec.Type)
	result := DamageResult{
		SourceEntityID:   req.Source.EntityID,
		SourceEntityType: req.Source.EntityType,
		TargetEntityID:   req.Target.EntityID,
		TargetEntityType: req.Target.EntityType,
		RemainingHealth:  req.Target.Health,
		RemainingShield:  req.Target.Shield,
		BaseAmount:       int(modified.BaseAmount),
		ModifiedAmount:   modified.ModifiedAmount,
		Type:             req.Spec.Type,
		Cause:            req.Spec.Cause,
		AppliedModifiers: modified.AppliedModifiers,
	}

	if req.Target.Health <= 0 {
		result.Kind = DamageResultKindDiscardedLethalTarget
		result.Ignored = true
		result.Discarded = true
		result.Reason = "lethal_target"
		return result
	}

	if modified.ModifiedAmount == 0 {
		result.Kind = DamageResultKindIneffective
		result.Ignored = true
		result.Reason = "ineffective"
		return result
	}

	if modified.ModifiedAmount > 0 {
		resolveDamage(&result, req)
	} else {
		resolveRestoration(&result, req, -modified.ModifiedAmount)
	}

	if result.Kind == DamageResultKindIneffective {
		return result
	}

	if req.Spec.DoT.Enabled {
		dotSource := req.Source
		dotSource.Cause = DamageCauseDot
		result.CreatedDamageOverTime = []ActiveDamageOverTime{
			{
				Source: dotSource,
				Target: DamageTargetRef{
					EntityID:   req.Target.EntityID,
					EntityType: req.Target.EntityType,
				},
				AmountPerTick:   req.Spec.DoT.AmountPerTick,
				TickSeconds:     req.Spec.DoT.TickSeconds,
				DurationSeconds: req.Spec.DoT.DurationSeconds,
				Type:            req.Spec.DoT.Type,
				Modifiers:       req.Spec.DoT.Modifiers,
				StackKey:        req.Spec.DoT.StackKey,
				StackingPolicy:  req.Spec.DoT.StackingPolicy,
				MaxStacks:       req.Spec.DoT.MaxStacks,
			},
		}
	}

	return result
}

func resolveDamage(result *DamageResult, req DamageResolutionRequest) {
	damageToApply := result.ModifiedAmount
	if !req.Spec.BypassShield && req.Target.Shield > 0 {
		absorbed := min(req.Target.Shield, damageToApply)
		result.AbsorbedByShield = absorbed
		result.RemainingShield = req.Target.Shield - absorbed
		damageToApply -= absorbed
		if damageToApply > 0 && req.Spec.OverflowPolicy == ShieldOverflowDiscard {
			damageToApply = 0
		}
	}

	if damageToApply > 0 {
		remaining := max(req.Target.Health-damageToApply, 0)
		result.AppliedToHealth = req.Target.Health - remaining
		result.RemainingHealth = remaining
	}

	result.Kind = DamageResultKindDamage
	result.Destroyed = result.RemainingHealth <= 0
	result.Fatal = result.Destroyed && req.Target.EntityType == EntityTypePlayer
}

func resolveRestoration(result *DamageResult, req DamageResolutionRequest, amount int) {
	remaining := amount
	if req.Spec.RestorationDestination == "" || req.Spec.RestorationDestination == RestorationDestinationHealth || req.Spec.RestorationDestination == RestorationDestinationBoth {
		restored := min(remaining, max(req.Target.MaxHealth-req.Target.Health, 0))
		result.RestoredToHealth = restored
		result.RemainingHealth += restored
		remaining -= restored
	}
	if req.Spec.RestorationDestination == RestorationDestinationShield || req.Spec.RestorationDestination == RestorationDestinationBoth {
		restored := min(remaining, max(req.Target.MaxShield-req.Target.Shield, 0))
		result.RestoredToShield = restored
		result.RemainingShield += restored
	}

	if result.RestoredToHealth == 0 && result.RestoredToShield == 0 {
		result.Kind = DamageResultKindIneffective
		result.Ignored = true
		result.Reason = "ineffective_restoration"
		return
	}

	if result.RestoredToHealth > 0 {
		result.Kind = DamageResultKindHealing
	} else {
		result.Kind = DamageResultKindRepair
	}
}
