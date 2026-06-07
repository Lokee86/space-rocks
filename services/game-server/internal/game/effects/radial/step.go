package radial

import (
	"github.com/Lokee86/space-rocks/server/internal/game/space"
)

func Step(effect *Effect, delta float64, candidates []Candidate) StepResult {
	result := StepResult{}
	if effect.AgeSeconds >= effect.Spec.TotalSeconds {
		result.Expired = true
		return result
	}

	if effect.Spec.CoverageMode == CoverageExpandingFill {
		radius := effectFillRadius(effect)
		if radius <= 0 {
			effect.AgeSeconds += delta
			return result
		}

		for _, candidate := range candidates {
			if !effect.Spec.TargetFilter.Allows(candidate.Kind) {
				continue
			}
			if !fillOverlapsCandidate(radius, space.Delta(effect.Origin, candidate.Position).Length(), candidate.Radius) {
				continue
			}

			result.Hits = append(result.Hits, Hit{
				EffectID:       effect.ID,
				SourceID:       effect.SourceID,
				SourcePlayerID: effect.SourcePlayerID,
				TargetID:       candidate.ID,
				TargetKind:     candidate.Kind,
				TargetPosition: candidate.Position,
				Damage:         effect.Spec.Damage,
			})
		}

		effect.AgeSeconds += delta
		return result
	}

	for i := range effect.Zones {
		zone := &effect.Zones[i]
		if effect.AgeSeconds < zone.StartsAt {
			continue
		}
		if effect.AgeSeconds >= zone.ExpiresAt {
			continue
		}
		if effect.AgeSeconds < zone.NextTickAt {
			continue
		}

		for _, candidate := range candidates {
			if !effect.Spec.TargetFilter.Allows(candidate.Kind) {
				continue
			}

			if !zoneOverlapsCandidate(*zone, space.Delta(effect.Origin, candidate.Position).Length(), candidate.Radius) {
				continue
			}

			result.Hits = append(result.Hits, Hit{
				EffectID:       effect.ID,
				SourceID:       effect.SourceID,
				SourcePlayerID: effect.SourcePlayerID,
				ZoneIndex:      zone.Index,
				TargetID:       candidate.ID,
				TargetKind:     candidate.Kind,
				TargetPosition: candidate.Position,
				Damage:         effect.Spec.Damage,
			})
		}

		zone.NextTickAt += effect.Spec.TickSeconds
	}

	effect.AgeSeconds += delta

	return result
}

func effectFillRadius(effect *Effect) float64 {
	radius := 0.0
	for i := range effect.Zones {
		zone := effect.Zones[i]
		if effect.AgeSeconds < zone.StartsAt {
			continue
		}
		if zone.OuterRadius > radius {
			radius = zone.OuterRadius
		}
	}
	return radius
}
