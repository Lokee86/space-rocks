package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/damage"
	"github.com/Lokee86/space-rocks/server/internal/game/events"
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

func TestDamageAppliedEventForResultCreatesEvent(t *testing.T) {
	event, ok := damageAppliedEventForResult(damage.DamageResult{
		SourceEntityID:   "bullet-1",
		SourceEntityType: "projectile",
		TargetEntityID:   "asteroid-1",
		TargetEntityType: "asteroid",
		Type:             damage.DamageTypeExplosive,
		Cause:            damage.DamageCauseProjectile,
		BaseAmount:       3,
		ModifiedAmount:   5,
		AppliedToHealth:  4,
		AbsorbedByShield: 1,
		RemainingHealth:  6,
		RemainingShield:  2,
	}, 12.5, 34.75)

	if !ok {
		t.Fatal("expected damage applied event to be created")
	}
	if event.Type != events.EventDamageApplied {
		t.Fatalf("expected event type %q, got %q", events.EventDamageApplied, event.Type)
	}
	if event.SourceID != "bullet-1" || event.SourceType != "projectile" {
		t.Fatalf("expected source to be preserved, got %q/%q", event.SourceID, event.SourceType)
	}
	if event.TargetID != "asteroid-1" || event.TargetType != "asteroid" {
		t.Fatalf("expected target to be preserved, got %q/%q", event.TargetID, event.TargetType)
	}
	if event.DamageType != string(damage.DamageTypeExplosive) {
		t.Fatalf("expected damage type %q, got %q", damage.DamageTypeExplosive, event.DamageType)
	}
	if event.DamageCause != string(damage.DamageCauseProjectile) {
		t.Fatalf("expected damage cause %q, got %q", damage.DamageCauseProjectile, event.DamageCause)
	}
	if event.BaseAmount != 3 || event.ModifiedAmount != 5 {
		t.Fatalf("expected base/modified amounts 3/5, got %d/%d", event.BaseAmount, event.ModifiedAmount)
	}
	if event.AppliedToHealth != 4 || event.AbsorbedByShield != 1 {
		t.Fatalf("expected applied/absorbed amounts 4/1, got %d/%d", event.AppliedToHealth, event.AbsorbedByShield)
	}
	if event.RemainingHealth != 6 || event.RemainingShield != 2 {
		t.Fatalf("expected remaining health/shield 6/2, got %d/%d", event.RemainingHealth, event.RemainingShield)
	}
	if event.X != 12.5 || event.Y != 34.75 {
		t.Fatalf("expected coordinates (12.5, 34.75), got (%v, %v)", event.X, event.Y)
	}
}

func TestDamageAppliedEventForResultSuppressesIgnoredAndNoOpResults(t *testing.T) {
	if event, ok := damageAppliedEventForResult(damage.DamageResult{Ignored: true}, 1, 2); ok || event != (events.Event{}) {
		t.Fatal("expected ignored damage result to produce no event")
	}

	if event, ok := damageAppliedEventForResult(damage.DamageResult{AppliedToHealth: 0, AbsorbedByShield: 0}, 1, 2); ok || event != (events.Event{}) {
		t.Fatal("expected no-op damage result to produce no event")
	}
}

func TestEventStateForDomainEventConvertsDamageApplied(t *testing.T) {
	event := eventStateForDomainEvent(events.Event{
		Type:           events.EventDamageApplied,
		SourceType:     "projectile",
		SourceID:       "bullet-1",
		DamageType:     "explosive",
		ModifiedAmount: 5,
		X:              12.5,
		Y:              34.75,
	})

	if event.Type != "damage_applied" {
		t.Fatalf("expected event type %q, got %q", "damage_applied", event.Type)
	}
	if event.SourceType != "projectile" {
		t.Fatalf("expected source type %q, got %q", "projectile", event.SourceType)
	}
	if event.SourceID != "bullet-1" {
		t.Fatalf("expected source id %q, got %q", "bullet-1", event.SourceID)
	}
	if event.EffectType != "explosive" {
		t.Fatalf("expected effect type %q, got %q", "explosive", event.EffectType)
	}
	if event.Amount != 5 {
		t.Fatalf("expected amount %d, got %d", 5, event.Amount)
	}
	if event.X != 12.5 || event.Y != 34.75 {
		t.Fatalf("expected damage applied coordinates (12.5, 34.75), got (%v, %v)", event.X, event.Y)
	}
}

func TestEventStateForDomainEventConvertsDamageOverTimeStarted(t *testing.T) {
	event := eventStateForDomainEvent(events.Event{
		Type:       events.EventDamageOverTimeStarted,
		SourceType: "asteroid",
		SourceID:   "hazard-1",
		DamageType: "radioactive",
		Amount:     2,
	})

	if event.Type != "damage_over_time_started" {
		t.Fatalf("expected event type %q, got %q", "damage_over_time_started", event.Type)
	}
	if event.SourceType != "asteroid" || event.SourceID != "hazard-1" {
		t.Fatalf("expected source to be preserved, got %q/%q", event.SourceID, event.SourceType)
	}
	if event.EffectType != "radioactive" {
		t.Fatalf("expected effect type %q, got %q", "radioactive", event.EffectType)
	}
	if event.Amount != 2 {
		t.Fatalf("expected amount %d, got %d", 2, event.Amount)
	}
}

func TestEventStateForDomainEventConvertsDamageOverTimeTick(t *testing.T) {
	event := eventStateForDomainEvent(events.Event{
		Type:           events.EventDamageOverTimeTick,
		SourceType:     "asteroid",
		SourceID:       "hazard-1",
		DamageType:     "radioactive",
		ModifiedAmount: 3,
		X:              11.5,
		Y:              22.25,
	})

	if event.Type != "damage_over_time_tick" {
		t.Fatalf("expected event type %q, got %q", "damage_over_time_tick", event.Type)
	}
	if event.SourceType != "asteroid" || event.SourceID != "hazard-1" {
		t.Fatalf("expected source to be preserved, got %q/%q", event.SourceID, event.SourceType)
	}
	if event.EffectType != "radioactive" {
		t.Fatalf("expected effect type %q, got %q", "radioactive", event.EffectType)
	}
	if event.Amount != 3 {
		t.Fatalf("expected amount %d, got %d", 3, event.Amount)
	}
	if event.X != 11.5 || event.Y != 22.25 {
		t.Fatalf("expected damage over time tick coordinates (11.5, 22.25), got (%v, %v)", event.X, event.Y)
	}
}

func TestDamageOverTimeStartedEventForEffect(t *testing.T) {
	event := damageOverTimeStartedEvent(damage.ActiveDamageOverTime{
		Source: damage.DamageSource{
			EntityID:   "hazard-1",
			EntityType: damage.EntityTypeAsteroid,
			Cause:      damage.DamageCauseDot,
		},
		Target: damage.DamageTargetRef{
			EntityID:   "player-1",
			EntityType: damage.EntityTypePlayer,
		},
		AmountPerTick:  2,
		TickSeconds:    0.5,
		DurationSeconds: 3,
		Type:           damage.DamageTypeRadioactive,
	})

	if event.Type != events.EventDamageOverTimeStarted {
		t.Fatalf("expected event type %q, got %q", events.EventDamageOverTimeStarted, event.Type)
	}
	if event.SourceID != "hazard-1" || event.SourceType != string(damage.EntityTypeAsteroid) {
		t.Fatalf("expected source to be preserved, got %q/%q", event.SourceID, event.SourceType)
	}
	if event.TargetID != "player-1" || event.TargetType != string(damage.EntityTypePlayer) {
		t.Fatalf("expected target to be preserved, got %q/%q", event.TargetID, event.TargetType)
	}
	if event.DamageType != string(damage.DamageTypeRadioactive) {
		t.Fatalf("expected damage type %q, got %q", damage.DamageTypeRadioactive, event.DamageType)
	}
	if event.DamageCause != string(damage.DamageCauseDot) {
		t.Fatalf("expected damage cause %q, got %q", damage.DamageCauseDot, event.DamageCause)
	}
	if event.Amount != 2 {
		t.Fatalf("expected amount 2, got %d", event.Amount)
	}
}

func TestDamageOverTimeTickEventForResult(t *testing.T) {
	event, ok := damageOverTimeTickEvent(damage.DamageResult{
		SourceEntityID:   "hazard-1",
		SourceEntityType: damage.EntityTypeAsteroid,
		TargetEntityID:   "player-1",
		TargetEntityType: damage.EntityTypePlayer,
		Type:             damage.DamageTypeRadioactive,
		Cause:            damage.DamageCauseDot,
		BaseAmount:       2,
		ModifiedAmount:   3,
		AppliedToHealth:  3,
		AbsorbedByShield: 0,
		RemainingHealth:  7,
		RemainingShield:  1,
	}, 11.5, 22.25)

	if !ok {
		t.Fatal("expected damage over time tick event to be created")
	}
	if event.Type != events.EventDamageOverTimeTick {
		t.Fatalf("expected event type %q, got %q", events.EventDamageOverTimeTick, event.Type)
	}
	if event.SourceID != "hazard-1" || event.SourceType != "asteroid" {
		t.Fatalf("expected source to be preserved, got %q/%q", event.SourceID, event.SourceType)
	}
	if event.TargetID != "player-1" || event.TargetType != "player" {
		t.Fatalf("expected target to be preserved, got %q/%q", event.TargetID, event.TargetType)
	}
	if event.DamageType != string(damage.DamageTypeRadioactive) {
		t.Fatalf("expected damage type %q, got %q", damage.DamageTypeRadioactive, event.DamageType)
	}
	if event.DamageCause != string(damage.DamageCauseDot) {
		t.Fatalf("expected damage cause %q, got %q", damage.DamageCauseDot, event.DamageCause)
	}
	if event.ModifiedAmount != 3 || event.AppliedToHealth != 3 {
		t.Fatalf("expected modified/applied amounts 3/3, got %d/%d", event.ModifiedAmount, event.AppliedToHealth)
	}
	if event.X != 11.5 || event.Y != 22.25 {
		t.Fatalf("expected coordinates (11.5, 22.25), got (%v, %v)", event.X, event.Y)
	}
}
