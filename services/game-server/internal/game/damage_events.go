package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/damage"
	"github.com/Lokee86/space-rocks/server/internal/game/events"
)

func damageOverTimeStartedEvent(effect damage.ActiveDamageOverTime) events.Event {
	return events.Event{
		Type:         events.EventDamageOverTimeStarted,
		SourceID:     effect.Source.EntityID,
		SourceType:   string(effect.Source.EntityType),
		TargetID:     effect.Target.EntityID,
		TargetType:   string(effect.Target.EntityType),
		DamageType:   string(effect.Type),
		DamageCause:  string(effect.Source.Cause),
		Amount:       effect.AmountPerTick,
	}
}

func damageAppliedEventForResult(result damage.DamageResult, x float64, y float64) (events.Event, bool) {
	if result.Ignored {
		return events.Event{}, false
	}
	if result.AppliedToHealth == 0 && result.AbsorbedByShield == 0 {
		return events.Event{}, false
	}

	return events.Event{
		Type:             events.EventDamageApplied,
		SourceID:         result.SourceEntityID,
		SourceType:       string(result.SourceEntityType),
		TargetID:         result.TargetEntityID,
		TargetType:       string(result.TargetEntityType),
		DamageType:       string(result.Type),
		DamageCause:      string(result.Cause),
		BaseAmount:       result.BaseAmount,
		ModifiedAmount:   result.ModifiedAmount,
		AppliedToHealth:  result.AppliedToHealth,
		AbsorbedByShield: result.AbsorbedByShield,
		RemainingHealth:  result.RemainingHealth,
		RemainingShield:  result.RemainingShield,
		X:                x,
		Y:                y,
	}, true
}

func damageOverTimeTickEvent(result damage.DamageResult, x float64, y float64) (events.Event, bool) {
	if result.Ignored {
		return events.Event{}, false
	}
	if result.AppliedToHealth == 0 && result.AbsorbedByShield == 0 {
		return events.Event{}, false
	}

	return events.Event{
		Type:             events.EventDamageOverTimeTick,
		SourceID:         result.SourceEntityID,
		SourceType:       string(result.SourceEntityType),
		TargetID:         result.TargetEntityID,
		TargetType:       string(result.TargetEntityType),
		DamageType:       string(result.Type),
		DamageCause:      string(result.Cause),
		BaseAmount:       result.BaseAmount,
		ModifiedAmount:   result.ModifiedAmount,
		AppliedToHealth:  result.AppliedToHealth,
		AbsorbedByShield: result.AbsorbedByShield,
		RemainingHealth:  result.RemainingHealth,
		RemainingShield:  result.RemainingShield,
		X:                x,
		Y:                y,
	}, true
}
