package radial

import (
	"github.com/Lokee86/space-rocks/server/internal/game/damage"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

type Candidate struct {
	ID       string
	Kind     TargetKind
	Position physics.Vector2
	Radius   float64
}

type Hit struct {
	EffectID       string
	SourceID       string
	SourcePlayerID string
	ZoneIndex      int
	TargetID       string
	TargetKind     TargetKind
	TargetPosition physics.Vector2
	Damage         damage.DamageSpec
}

type StepResult struct {
	Hits    []Hit
	Expired bool
}
