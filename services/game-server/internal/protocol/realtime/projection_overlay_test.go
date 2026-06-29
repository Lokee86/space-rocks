package realtime

import (
	"testing"

	game "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
)

func TestProjectOverlayLaneUsesReceiverLocalFields(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		Lives:  2,
		Players: map[string]runtime.ShipState{
			"player-1": {
				ID:                         "player-1",
				ShipType:                   "v_wing",
				PrimaryWeaponID:            "laser",
				PrimaryAmmoPolicy:          "limited",
				PrimaryCooldownRemaining:   12,
				PrimaryAmmoRemaining:       3,
				SecondaryWeaponID:          "torpedo",
				SecondaryAmmoPolicy:        "limited",
				SecondaryCooldownRemaining: 7,
				SecondaryAmmoRemaining:     1,
			},
			"player-2": {
				ID:                         "player-2",
				ShipType:                   "v_wing",
				PrimaryCooldownRemaining:   99,
				PrimaryAmmoRemaining:       99,
				SecondaryCooldownRemaining: 99,
				SecondaryAmmoRemaining:     99,
			},
		},
		PlayerSessions: map[string]game.PlayerSessionState{
			"player-1": {
				ID:                  "player-1",
				ShipType:            "v_wing",
				Score:               42,
				Lives:               2,
				RespawnCooldown:     1.5,
				PrimaryWeaponID:     "laser",
				PrimaryAmmoPolicy:   "limited",
				SecondaryWeaponID:   "mine",
				SecondaryAmmoPolicy: "infinite",
			},
		},
	}

	first := ProjectOverlayLane(snapshot, "player-1")
	second := ProjectOverlayLane(snapshot, "player-1")

	if first != second {
		t.Fatalf("expected deterministic overlay projection, got %#v then %#v", first, second)
	}

	receiver := first.Receiver
	if receiver.SelfID != "player-1" || receiver.Lives != 2 {
		t.Fatalf("expected receiver identity/lives to be preserved, got %#v", receiver)
	}
	if receiver.Score != 42 || receiver.RespawnCooldown != 1.5 {
		t.Fatalf("expected receiver score/respawn to be preserved, got %#v", receiver)
	}
	if receiver.PrimaryWeaponID != "laser" || receiver.PrimaryAmmoPolicy != "limited" {
		t.Fatalf("expected primary weapon policy fields, got %#v", receiver)
	}
	if receiver.PrimaryCooldownRemaining != 12 || receiver.PrimaryAmmoRemaining != 3 {
		t.Fatalf("expected local primary cooldown/ammo from receiver ship, got %#v", receiver)
	}
	if receiver.SecondaryWeaponID != "torpedo" || receiver.SecondaryAmmoPolicy != "limited" {
		t.Fatalf("expected active secondary weapon policy fields from receiver ship, got %#v", receiver)
	}
	if receiver.SecondaryCooldownRemaining != 7 || receiver.SecondaryAmmoRemaining != 1 {
		t.Fatalf("expected local secondary cooldown/ammo from receiver ship, got %#v", receiver)
	}

	packet := BuildOverlayFullPacket(snapshot, "player-1", 4)
	if packet.Type != PacketFamilyOverlayFull {
		t.Fatalf("expected overlay full packet type, got %q", packet.Type)
	}
	if packet.Metadata.Lane != LaneOverlay || packet.Metadata.Sequence != 4 || packet.Metadata.BaselineID != "player-1" || packet.Metadata.SnapshotID != "player-1" || packet.Metadata.SnapshotKind != SnapshotKind("full") || packet.Metadata.ChunkIndex != 0 || packet.Metadata.ChunkCount != 1 || !packet.Metadata.IsFinalChunk {
		t.Fatalf("expected overlay metadata to be populated, got %#v", packet.Metadata)
	}
	if packet.Receiver != receiver {
		t.Fatalf("expected overlay packet receiver to match projection, got %#v want %#v", packet.Receiver, receiver)
	}

	other := snapshot.Players["player-2"]
	if other.PrimaryCooldownRemaining != 99 || other.PrimaryAmmoRemaining != 99 || other.SecondaryCooldownRemaining != 99 || other.SecondaryAmmoRemaining != 99 {
		t.Fatalf("expected other player snapshot to remain untouched, got %#v", other)
	}
	if receiver.PrimaryCooldownRemaining == other.PrimaryCooldownRemaining && receiver.PrimaryAmmoRemaining == other.PrimaryAmmoRemaining && receiver.SecondaryCooldownRemaining == other.SecondaryCooldownRemaining && receiver.SecondaryAmmoRemaining == other.SecondaryAmmoRemaining {
		t.Fatal("expected overlay projection to use receiver-local weapon state, not other player values")
	}
}
