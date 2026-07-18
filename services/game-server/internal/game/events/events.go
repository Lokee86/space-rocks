package events

type Type string

const (
	EventBulletBlast           Type = "bullet_blast"
	EventRadialEffectStarted   Type = "radial_effect_started"
	EventShipDeath             Type = "ship_death"
	EventPickupCollected       Type = "pickup_collected"
	EventPickupEffectApplied   Type = "pickup_effect_applied"
	EventPickupExpired         Type = "pickup_expired"
	EventPickupDropped         Type = "pickup_dropped"
	EventDamageApplied         Type = "damage_applied"
	EventDamageBlocked         Type = "damage_blocked"
	EventHealingApplied        Type = "healing_applied"
	EventRepairApplied         Type = "repair_applied"
	EventDamageIneffective     Type = "damage_ineffective"
	EventDamageDiscarded       Type = "damage_discarded"
	EventDamageOverTimeStarted Type = "damage_over_time_started"
	EventDamageOverTimeTick    Type = "damage_over_time_tick"
)

type Event struct {
	Type             Type
	PlayerID         string
	PickupID         string
	PickupType       string
	SourceType       string
	SourceID         string
	TableID          string
	EffectType       string
	TargetID         string
	TargetType       string
	DamageType       string
	DamageCause      string
	BaseAmount       int
	ModifiedAmount   int
	AppliedToHealth  int
	AbsorbedByShield int
	RestoredToHealth int
	RestoredToShield int
	ResultReason     string
	RemainingHealth  int
	RemainingShield  int
	Amount           int
	Lives            int
	LivesAfter       int
	RespawnDelay     float64
	X                float64
	Y                float64
}
