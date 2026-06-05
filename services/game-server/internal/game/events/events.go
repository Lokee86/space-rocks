package events

type Type string

const (
	EventBulletBlast         Type = "bullet_blast"
	EventShipDeath           Type = "ship_death"
	EventPickupCollected     Type = "pickup_collected"
	EventPickupEffectApplied Type = "pickup_effect_applied"
	EventPickupDropped       Type = "pickup_dropped"
)

type Event struct {
	Type         Type
	PlayerID     string
	PickupID     string
	PickupType   string
	SourceType   string
	SourceID     string
	TableID      string
	EffectType   string
	Amount       int
	Lives        int
	LivesAfter   int
	RespawnDelay float64
	X            float64
	Y            float64
}
