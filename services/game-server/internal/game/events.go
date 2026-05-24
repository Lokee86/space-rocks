package game

type gameEventType string

const (
	gameEventBulletBlast gameEventType = "bullet_blast"
	gameEventShipDeath   gameEventType = "ship_death"
)

type gameEvent struct {
	Type         gameEventType
	PlayerID     string
	Lives        int
	RespawnDelay float64
	X            float64
	Y            float64
}

func (game *Game) recordDomainEvent(event gameEvent) {
	game.broadcastEvent(eventStateForDomainEvent(event))
}

func eventStateForDomainEvent(event gameEvent) EventState {
	switch event.Type {
	case gameEventBulletBlast:
		return EventState{
			Type: PacketTypeBulletBlast,
			X:    event.X,
			Y:    event.Y,
		}
	case gameEventShipDeath:
		return EventState{
			Type:         PacketTypeShipDeath,
			PlayerID:     event.PlayerID,
			Lives:        event.Lives,
			RespawnDelay: event.RespawnDelay,
			X:            event.X,
			Y:            event.Y,
		}
	default:
		return EventState{}
	}
}
