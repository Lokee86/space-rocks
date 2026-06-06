package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/events"
	"github.com/Lokee86/space-rocks/server/internal/game/damage"
)

func TestEventStateForDomainEventConvertsBulletBlast(t *testing.T) {
	event := eventStateForDomainEvent(events.Event{
		Type: events.EventBulletBlast,
		X:    12.5,
		Y:    34.75,
	})

	if event.Type != PacketTypeBulletBlast {
		t.Fatalf("expected event type %q, got %q", PacketTypeBulletBlast, event.Type)
	}
	if event.X != 12.5 || event.Y != 34.75 {
		t.Fatalf("expected bullet blast coordinates (12.5, 34.75), got (%v, %v)", event.X, event.Y)
	}
}

func TestEventStateForDomainEventConvertsShipDeath(t *testing.T) {
	event := eventStateForDomainEvent(events.Event{
		Type:         events.EventShipDeath,
		PlayerID:     "player-1",
		Lives:        2,
		RespawnDelay: 1.25,
		X:            45.5,
		Y:            67.75,
	})

	if event.Type != PacketTypeShipDeath {
		t.Fatalf("expected event type %q, got %q", PacketTypeShipDeath, event.Type)
	}
	if event.PlayerID != "player-1" {
		t.Fatalf("expected player ID %q, got %q", "player-1", event.PlayerID)
	}
	if event.Lives != 2 {
		t.Fatalf("expected lives 2, got %d", event.Lives)
	}
	if event.RespawnDelay != 1.25 {
		t.Fatalf("expected respawn delay 1.25, got %v", event.RespawnDelay)
	}
	if event.X != 45.5 || event.Y != 67.75 {
		t.Fatalf("expected ship death coordinates (45.5, 67.75), got (%v, %v)", event.X, event.Y)
	}
}

func TestEventStateForDomainEventConvertsPickupCollected(t *testing.T) {
	event := eventStateForDomainEvent(events.Event{
		Type:       events.EventPickupCollected,
		PlayerID:   "player-1",
		PickupID:   "pickup-1",
		PickupType: "1_up",
		X:          21.5,
		Y:          43.25,
	})

	if event.Type != "pickup_collected" {
		t.Fatalf("expected event type %q, got %q", "pickup_collected", event.Type)
	}
	if event.PlayerID != "player-1" {
		t.Fatalf("expected player ID %q, got %q", "player-1", event.PlayerID)
	}
	if event.PickupID != "pickup-1" {
		t.Fatalf("expected pickup ID %q, got %q", "pickup-1", event.PickupID)
	}
	if event.PickupType != "1_up" {
		t.Fatalf("expected pickup type %q, got %q", "1_up", event.PickupType)
	}
	if event.X != 21.5 || event.Y != 43.25 {
		t.Fatalf("expected pickup collected coordinates (21.5, 43.25), got (%v, %v)", event.X, event.Y)
	}
}

func TestEventStateForDomainEventConvertsPickupEffectApplied(t *testing.T) {
	event := eventStateForDomainEvent(events.Event{
		Type:       events.EventPickupEffectApplied,
		PlayerID:   "player-1",
		PickupID:   "pickup-1",
		PickupType: "1_up",
		EffectType: "add_lives",
		Amount:     1,
		LivesAfter: 6,
	})

	if event.Type != "pickup_effect_applied" {
		t.Fatalf("expected event type %q, got %q", "pickup_effect_applied", event.Type)
	}
	if event.PlayerID != "player-1" {
		t.Fatalf("expected player ID %q, got %q", "player-1", event.PlayerID)
	}
	if event.PickupID != "pickup-1" {
		t.Fatalf("expected pickup ID %q, got %q", "pickup-1", event.PickupID)
	}
	if event.PickupType != "1_up" {
		t.Fatalf("expected pickup type %q, got %q", "1_up", event.PickupType)
	}
	if event.EffectType != "add_lives" {
		t.Fatalf("expected effect type %q, got %q", "add_lives", event.EffectType)
	}
	if event.Amount != 1 {
		t.Fatalf("expected amount 1, got %d", event.Amount)
	}
	if event.LivesAfter != 6 {
		t.Fatalf("expected lives after 6, got %d", event.LivesAfter)
	}
}

func TestEventStateForDomainEventConvertsPickupExpired(t *testing.T) {
	event := eventStateForDomainEvent(events.Event{
		Type:       events.EventPickupExpired,
		PickupID:   "pickup-1",
		PickupType: "1_up",
		X:          21.5,
		Y:          43.25,
	})

	if event.Type != "pickup_expired" {
		t.Fatalf("expected event type %q, got %q", "pickup_expired", event.Type)
	}
	if event.PickupID != "pickup-1" {
		t.Fatalf("expected pickup ID %q, got %q", "pickup-1", event.PickupID)
	}
	if event.PickupType != "1_up" {
		t.Fatalf("expected pickup type %q, got %q", "1_up", event.PickupType)
	}
	if event.X != 21.5 || event.Y != 43.25 {
		t.Fatalf("expected pickup expired coordinates (21.5, 43.25), got (%v, %v)", event.X, event.Y)
	}
}

func TestRecordDomainEventQueuesBulletBlastPacketEvent(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()

	game.recordDomainEvent(events.Event{
		Type: events.EventBulletBlast,
		X:    12.5,
		Y:    34.75,
	})

	queuedEvents := game.pendingPresentationEvents[playerID]
	if len(queuedEvents) != 1 {
		t.Fatalf("expected 1 queued event, got %d", len(queuedEvents))
	}
	event := queuedEvents[0]
	if event.Type != PacketTypeBulletBlast {
		t.Fatalf("expected event type %q, got %q", PacketTypeBulletBlast, event.Type)
	}
	if event.X != 12.5 || event.Y != 34.75 {
		t.Fatalf("expected bullet blast coordinates (12.5, 34.75), got (%v, %v)", event.X, event.Y)
	}
}

func TestRecordDomainEventQueuesShipDeathPacketEvent(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()

	game.recordDomainEvent(events.Event{
		Type:         events.EventShipDeath,
		PlayerID:     "player-1",
		Lives:        2,
		RespawnDelay: 1.25,
		X:            45.5,
		Y:            67.75,
	})

	queuedEvents := game.pendingPresentationEvents[playerID]
	if len(queuedEvents) != 1 {
		t.Fatalf("expected 1 queued event, got %d", len(queuedEvents))
	}
	event := queuedEvents[0]
	if event.Type != PacketTypeShipDeath {
		t.Fatalf("expected event type %q, got %q", PacketTypeShipDeath, event.Type)
	}
	if event.PlayerID != "player-1" {
		t.Fatalf("expected player ID %q, got %q", "player-1", event.PlayerID)
	}
	if event.Lives != 2 {
		t.Fatalf("expected lives 2, got %d", event.Lives)
	}
	if event.RespawnDelay != 1.25 {
		t.Fatalf("expected respawn delay 1.25, got %v", event.RespawnDelay)
	}
	if event.X != 45.5 || event.Y != 67.75 {
		t.Fatalf("expected ship death coordinates (45.5, 67.75), got (%v, %v)", event.X, event.Y)
	}
}

func TestStateDrainsDomainEventPacketEvents(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	game.recordDomainEvent(events.Event{
		Type: events.EventBulletBlast,
		X:    12.5,
		Y:    34.75,
	})

	packet := game.StatePacket(playerID)
	if len(packet.Events) != 1 {
		t.Fatalf("expected first state packet to include 1 event, got %d", len(packet.Events))
	}
	if packet.Events[0].Type != PacketTypeBulletBlast {
		t.Fatalf("expected event type %q, got %q", PacketTypeBulletBlast, packet.Events[0].Type)
	}

	flushed := game.StatePacket(playerID)
	if len(flushed.Events) != 0 {
		t.Fatalf("expected later state packet to include 0 events, got %d", len(flushed.Events))
	}
}

func TestDamagePresentationEventForResultIsNoOp(t *testing.T) {
	event := damagePresentationEventForResult(damage.DamageResult{})

	if event != nil {
		t.Fatalf("expected no presentation event, got %v", event)
	}
}
