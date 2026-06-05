package events

type Type string

const (
	EventBulletBlast Type = "bullet_blast"
	EventShipDeath   Type = "ship_death"
	EventPickupCollected Type = "pickup_collected"
)

type Event struct {
	Type         Type
	PlayerID     string
	PickupID     string
	PickupType   string
	Lives        int
	LivesAfter   int
	RespawnDelay float64
	X            float64
	Y            float64
}
