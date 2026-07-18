package damage

func BlockedResult(request DamageResolutionRequest, reason EligibilityBlockReason) DamageResult {
	return DamageResult{
		SourceEntityID:   request.Source.EntityID,
		SourceEntityType: request.Source.EntityType,
		TargetEntityID:   request.Target.EntityID,
		TargetEntityType: request.Target.EntityType,
		Kind:             DamageResultKindBlocked,
		BaseAmount:       request.Spec.Amount,
		ModifiedAmount:   request.Spec.Amount,
		Type:             request.Spec.Type,
		Cause:            request.Spec.Cause,
		RemainingHealth:  request.Target.Health,
		RemainingShield:  request.Target.Shield,
		Ignored:          true,
		Reason:           string(reason),
	}
}
