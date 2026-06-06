package damage

type AreaDamageRequest struct {
	Source     DamageSource
	OriginX    float64
	OriginY    float64
	Radius     float64
	Spec       DamageSpec
	Modifiers  []DamageModifier
	Candidates []DamageTarget
}

type AreaDamageResult struct {
	Results []DamageResult
}

func ResolveArea(req AreaDamageRequest) AreaDamageResult {
	if req.Radius <= 0 {
		return AreaDamageResult{}
	}

	result := AreaDamageResult{
		Results: make([]DamageResult, 0, len(req.Candidates)),
	}
	for _, candidate := range req.Candidates {
		modifiers := make([]DamageModifier, 0, len(req.Modifiers)+len(candidate.Modifiers))
		modifiers = append(modifiers, req.Modifiers...)
		modifiers = append(modifiers, candidate.Modifiers...)
		single := ResolveSingle(DamageResolutionRequest{
			Source: req.Source,
			Target: candidate,
			Spec:   req.Spec,
			Modifiers: modifiers,
		})
		result.Results = append(result.Results, single)
	}

	return result
}
