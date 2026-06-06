package damage

type DamageOverTimeSpec struct {
	Enabled         bool
	AmountPerTick   int
	TickSeconds     float64
	DurationSeconds float64
	Kind            DamageKind
	Modifiers       []DamageModifier
}

type DamageTargetRef struct {
	EntityID   string
	EntityType EntityType
}

type ActiveDamageOverTime struct {
	Source         DamageSource
	Target         DamageTargetRef
	AmountPerTick  int
	TickSeconds    float64
	DurationSeconds float64
	Kind           DamageKind
	Modifiers      []DamageModifier
}

type DamageOverTimeTickResult struct {
	Source           DamageSource
	Target           DamageTargetRef
	AmountPerTick    int
	TickSeconds      float64
	TickRemaining    float64
	DurationSeconds float64
	DurationRemaining float64
	Kind             DamageKind
	Modifiers        []DamageModifier
	Results          []DamageResult
	Expired          bool
}

func TickDamageOverTime(effect ActiveDamageOverTime, target DamageTarget, delta float64) DamageOverTimeTickResult {
	result := DamageOverTimeTickResult{
		Source:           effect.Source,
		Target:           effect.Target,
		AmountPerTick:    effect.AmountPerTick,
		TickSeconds:      effect.TickSeconds,
		DurationSeconds:  effect.DurationSeconds,
		DurationRemaining: max(effect.DurationSeconds-delta, 0),
		Kind:             effect.Kind,
		Modifiers:        effect.Modifiers,
	}

	if delta <= 0 {
		result.TickRemaining = effect.TickSeconds
		return result
	}

	tickRemaining := effect.TickSeconds - delta
	if tickRemaining > 0 {
		result.TickRemaining = tickRemaining
		return result
	}

	remainingDelta := delta
	remainingDuration := effect.DurationSeconds
	tickRemaining = effect.TickSeconds
	for remainingDelta >= tickRemaining && remainingDuration > 0 {
		single := ResolveSingle(DamageResolutionRequest{
			Source: effect.Source,
			Target: target,
			Spec: DamageSpec{
				Amount: effect.AmountPerTick,
				Kind:   effect.Kind,
				Cause:  DamageCauseDot,
			},
			Modifiers: effect.Modifiers,
		})
		result.Results = append(result.Results, single)
		remainingDelta -= tickRemaining
		remainingDuration -= tickRemaining
		tickRemaining = effect.TickSeconds
		if tickRemaining <= 0 {
			break
		}
	}

	result.TickRemaining = tickRemaining - remainingDelta
	if result.TickRemaining < 0 {
		result.TickRemaining = 0
	}
	result.DurationRemaining = max(effect.DurationSeconds-delta, 0)
	result.Expired = result.DurationRemaining <= 0
	return result
}
