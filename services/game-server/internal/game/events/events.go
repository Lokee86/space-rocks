package events

type Type string

const (
	EventBulletBlast Type = "bullet_blast"
	EventShipDeath   Type = "ship_death"
)

type Event struct {
	Type         Type
	PlayerID     string
	Lives        int
	RespawnDelay float64
	X            float64
	Y            float64
}
