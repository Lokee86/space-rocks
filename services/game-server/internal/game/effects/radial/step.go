package radial

import (
	"math"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/damage"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
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
			distance := space.Delta(effect.Origin, candidate.Position).Length()
			if !fillOverlapsCandidate(radius, distance, candidate.Radius) {
				continue
			}

			result.Hits = append(result.Hits, Hit{
				EffectID:       effect.ID,
				SourceID:       effect.SourceID,
				SourcePlayerID: effect.SourcePlayerID,
				TargetID:       candidate.ID,
				TargetKind:     candidate.Kind,
				TargetPosition: candidate.Position,
				Damage:         damageAtDistance(effect.Spec, distance),
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

			distance := space.Delta(effect.Origin, candidate.Position).Length()
			if !zoneOverlapsCandidate(*zone, distance, candidate.Radius) {
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
				Damage:         damageAtDistance(effect.Spec, distance),
			})
		}

		zone.NextTickAt += effect.Spec.TickSeconds
	}

	effect.AgeSeconds += delta

	return result
}

func damageAtDistance(spec Spec, distance float64) damage.DamageSpec {
	resolved := spec.Damage
	if spec.FalloffMode != FalloffLinear || resolved.Amount == 0 {
		return resolved
	}
	maximumRadius := float64(spec.ZoneCount) * spec.ZoneWidth
	if maximumRadius <= 0 {
		return resolved
	}
	minimum := math.Max(0, math.Min(spec.MinimumMultiplier, 1))
	ratio := math.Max(0, math.Min(distance/maximumRadius, 1))
	multiplier := 1 - ratio*(1-minimum)
	resolved.Amount = int(math.Round(float64(resolved.Amount) * multiplier))
	return resolved
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
