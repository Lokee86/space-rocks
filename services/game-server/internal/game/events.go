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
	case events.EventPickupCollected:
		return EventState{
			Type:       "pickup_collected",
			PlayerID:   event.PlayerID,
			PickupID:   event.PickupID,
			PickupType: event.PickupType,
			X:          event.X,
			Y:          event.Y,
		}
	case events.EventPickupEffectApplied:
		return EventState{
			Type:       "pickup_effect_applied",
			PlayerID:   event.PlayerID,
			PickupID:   event.PickupID,
			PickupType: event.PickupType,
			EffectType: event.EffectType,
			Amount:     event.Amount,
			LivesAfter: event.LivesAfter,
		}
	case events.EventPickupDropped:
		return EventState{
			Type:       "pickup_dropped",
			PickupID:   event.PickupID,
			PickupType: event.PickupType,
			SourceType: event.SourceType,
			SourceID:   event.SourceID,
			TableID:    event.TableID,
			X:          event.X,
			Y:          event.Y,
		}
	default:
		return EventState{}
	}
}

func (game *Game) broadcastEvent(event EventState) {
	for playerID := range game.entities.Players {
		game.pendingPresentationEvents[playerID] = append(game.pendingPresentationEvents[playerID], event)
	}
}
