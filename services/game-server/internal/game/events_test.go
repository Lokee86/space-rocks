package game

import "testing"

func TestEventStateForDomainEventConvertsBulletBlast(t *testing.T) {
	event := eventStateForDomainEvent(gameEvent{
		Type: gameEventBulletBlast,
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
	event := eventStateForDomainEvent(gameEvent{
		Type:         gameEventShipDeath,
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

func TestRecordDomainEventQueuesBulletBlastPacketEvent(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()

	game.recordDomainEvent(gameEvent{
		Type: gameEventBulletBlast,
		X:    12.5,
		Y:    34.75,
	})

	events := game.pendingEvents[playerID]
	if len(events) != 1 {
		t.Fatalf("expected 1 queued event, got %d", len(events))
	}
	event := events[0]
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

	game.recordDomainEvent(gameEvent{
		Type:         gameEventShipDeath,
		PlayerID:     "player-1",
		Lives:        2,
		RespawnDelay: 1.25,
		X:            45.5,
		Y:            67.75,
	})

	events := game.pendingEvents[playerID]
	if len(events) != 1 {
		t.Fatalf("expected 1 queued event, got %d", len(events))
	}
	event := events[0]
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
	game.recordDomainEvent(gameEvent{
		Type: gameEventBulletBlast,
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
