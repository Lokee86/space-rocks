package game

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/damage"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/events"
)

func damageOverTimeStartedEvent(effect damage.ActiveDamageOverTime) events.Event {
	return events.Event{
		Type:        events.EventDamageOverTimeStarted,
		SourceID:    effect.Source.EntityID,
		SourceType:  string(effect.Source.EntityType),
		TargetID:    effect.Target.EntityID,
		TargetType:  string(effect.Target.EntityType),
		DamageType:  string(effect.Type),
		DamageCause: string(effect.Source.Cause),
		Amount:      effect.AmountPerTick,
	}
}

func damageResultEvent(result damage.DamageResult, x float64, y float64) (events.Event, bool) {
	eventType := events.Type("")
	switch result.Kind {
	case "", damage.DamageResultKindDamage:
		if result.Ignored || result.AppliedToHealth == 0 && result.AbsorbedByShield == 0 {
			return events.Event{}, false
		}
		eventType = events.EventDamageApplied
	case damage.DamageResultKindBlocked:
		eventType = events.EventDamageBlocked
	case damage.DamageResultKindHealing:
		if result.RestoredToHealth == 0 {
			return events.Event{}, false
		}
		eventType = events.EventHealingApplied
	case damage.DamageResultKindRepair:
		if result.RestoredToShield == 0 {
			return events.Event{}, false
		}
		eventType = events.EventRepairApplied
	case damage.DamageResultKindIneffective:
		eventType = events.EventDamageIneffective
	case damage.DamageResultKindDiscardedLethalTarget:
		eventType = events.EventDamageDiscarded
	default:
		return events.Event{}, false
	}

	return events.Event{
		Type:             eventType,
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
		RestoredToHealth: result.RestoredToHealth,
		RestoredToShield: result.RestoredToShield,
		ResultReason:     result.Reason,
		RemainingHealth:  result.RemainingHealth,
		RemainingShield:  result.RemainingShield,
		X:                x,
		Y:                y,
	}, true
}

func damageAppliedEventForResult(result damage.DamageResult, x float64, y float64) (events.Event, bool) {
	event, ok := damageResultEvent(result, x, y)
	if !ok || event.Type != events.EventDamageApplied {
		return events.Event{}, false
	}
	return event, true
}

func damageOverTimeTickEvent(result damage.DamageResult, x float64, y float64) (events.Event, bool) {
	if result.Ignored || result.AppliedToHealth == 0 && result.AbsorbedByShield == 0 {
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
