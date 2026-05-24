package game

import "github.com/Lokee86/space-rocks/server/internal/game/events"

func (game *Game) recordDomainEvent(event events.Event) {
	game.broadcastEvent(eventStateForDomainEvent(event))
}

func eventStateForDomainEvent(event events.Event) EventState {
	switch event.Type {
	case events.EventBulletBlast:
		return EventState{
			Type: PacketTypeBulletBlast,
			X:    event.X,
			Y:    event.Y,
		}
	case events.EventShipDeath:
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

func (game *Game) broadcastEvent(event EventState) {
	for playerID := range game.state.Players {
		game.pendingPresentationEvents[playerID] = append(game.pendingPresentationEvents[playerID], event)
	}
}
