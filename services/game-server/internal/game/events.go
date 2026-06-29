package game

import (
	"fmt"

	"github.com/Lokee86/space-rocks/server/internal/game/events"
)

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
	case events.EventRadialEffectStarted:
		return EventState{
			Type:       "radial_effect_started",
			SourceID:   event.SourceID,
			EffectType: event.EffectType,
			X:          event.X,
			Y:          event.Y,
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
	case events.EventPickupExpired:
		return EventState{
			Type:       "pickup_expired",
			PickupID:   event.PickupID,
			PickupType: event.PickupType,
			X:          event.X,
			Y:          event.Y,
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
	case events.EventDamageApplied:
		return EventState{
			Type:         "damage_applied",
			SourceType:   event.SourceType,
			SourceID:     event.SourceID,
			EffectType:   event.DamageType,
			Amount:       event.ModifiedAmount,
			X:            event.X,
			Y:            event.Y,
		}
	case events.EventDamageOverTimeStarted:
		return EventState{
			Type:       "damage_over_time_started",
			SourceType: event.SourceType,
			SourceID:   event.SourceID,
			EffectType: event.DamageType,
			Amount:     event.Amount,
		}
	case events.EventDamageOverTimeTick:
		return EventState{
			Type:         "damage_over_time_tick",
			SourceType:   event.SourceType,
			SourceID:     event.SourceID,
			EffectType:   event.DamageType,
			Amount:       event.ModifiedAmount,
			X:            event.X,
			Y:            event.Y,
		}
	default:
		return EventState{}
	}
}

func (game *Game) broadcastEvent(event EventState) {
	for playerID := range game.playerSessions {
		game.nextPresentationEventID++
		game.pendingPresentationEvents[playerID] = append(game.pendingPresentationEvents[playerID], PendingPresentationEvent{
			EventID: formatPresentationEventID(game.nextPresentationEventID),
			Event:   event,
		})
	}
}

func formatPresentationEventID(id int) string {
	return fmt.Sprintf("presentation-event-%d", id)
}
