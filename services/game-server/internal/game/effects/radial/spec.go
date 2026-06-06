package radial

import "github.com/Lokee86/space-rocks/server/internal/game/damage"

type Spec struct {
	CoverageMode       CoverageMode
	ExpirationMode     ExpirationMode
	TargetFilter       TargetFilter
	ZoneCount          int
	ZoneWidth          float64
	ZoneSpawnSeconds   float64
	TickSeconds        float64
	TotalSeconds       float64
	ZoneLifetimeSeconds float64
	Damage             damage.DamageSpec
}
