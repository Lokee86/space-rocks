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
		single := ResolveSingle(DamageResolutionRequest{
			Source:    req.Source,
			Target:    candidate,
			Spec:      req.Spec,
			Modifiers: req.Modifiers,
		})
		result.Results = append(result.Results, single)
	}

	return result
}
