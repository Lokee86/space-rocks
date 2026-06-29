package realtime

import (
	"reflect"
	"testing"

	game "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
)

func TestProjectSessionLaneUsesSharedFactsAndDeterministicOrder(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "snapshot-1",
		PlayerSessions: map[string]game.PlayerSessionState{
			"player-b": {
				ID:                  "player-b",
				ShipType:            "v_wing",
				Score:               8,
				Lives:               2,
				RespawnCooldown:     1.25,
				PrimaryWeaponID:     "laser",
				PrimaryAmmoPolicy:   "limited",
				SecondaryWeaponID:   "mine",
				SecondaryAmmoPolicy: "infinite",
				SpawnX:              30,
				SpawnY:              40,
			},
			"player-a": {
				ID:                  "player-a",
				ShipType:            "v_wing",
				Score:               5,
				Lives:               3,
				RespawnCooldown:     0.75,
				PrimaryWeaponID:     "pulse",
				PrimaryAmmoPolicy:   "limited",
				SecondaryWeaponID:   "drone",
				SecondaryAmmoPolicy: "limited",
				SpawnX:              10,
				SpawnY:              20,
			},
		},
		PlayerLifecycle: map[string]string{
			"player-b": "active",
			"player-a": "active",
		},
		TotalAsteroids: 42,
		Players: map[string]runtime.ShipState{
			"player-a": {PrimaryCooldownRemaining: 99, PrimaryAmmoRemaining: 77, SecondaryCooldownRemaining: 55, SecondaryAmmoRemaining: 33},
			"player-b": {PrimaryCooldownRemaining: 88, PrimaryAmmoRemaining: 66, SecondaryCooldownRemaining: 44, SecondaryAmmoRemaining: 22},
		},
	}

	first := ProjectSessionLane(snapshot)
	second := ProjectSessionLane(snapshot)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("expected deterministic session projection, got %#v then %#v", first, second)
	}
	if first.TotalAsteroids != 42 {
		t.Fatalf("expected total asteroids to be preserved, got %d", first.TotalAsteroids)
	}
	if len(first.Players) != 2 || first.Players[0].ID != "player-a" || first.Players[1].ID != "player-b" {
		t.Fatalf("expected players sorted by ID, got %#v", first.Players)
	}

	playerA := first.Players[0]
	if playerA.Score != 5 || playerA.Lives != 3 || playerA.RespawnCooldown != 0.75 || playerA.ShipType != "v_wing" || playerA.PrimaryWeaponID != "pulse" || playerA.PrimaryAmmoPolicy != "limited" || playerA.SecondaryWeaponID != "drone" || playerA.SecondaryAmmoPolicy != "limited" || playerA.SpawnX != 10 || playerA.SpawnY != 20 {
		t.Fatalf("expected player-a shared session facts to be preserved, got %#v", playerA)
	}

	playerB := first.Players[1]
	if playerB.Score != 8 || playerB.Lives != 2 || playerB.RespawnCooldown != 1.25 || playerB.ShipType != "v_wing" || playerB.PrimaryWeaponID != "laser" || playerB.PrimaryAmmoPolicy != "limited" || playerB.SecondaryWeaponID != "mine" || playerB.SecondaryAmmoPolicy != "infinite" || playerB.SpawnX != 30 || playerB.SpawnY != 40 {
		t.Fatalf("expected player-b shared session facts to be preserved, got %#v", playerB)
	}

	if first.PlayerLifecycle[0].PlayerID != "player-a" || first.PlayerLifecycle[1].PlayerID != "player-b" {
		t.Fatalf("expected lifecycle sorted by player ID, got %#v", first.PlayerLifecycle)
	}
	if first.PlayerLifecycle[0].Status != "active" || first.PlayerLifecycle[1].Status != "active" {
		t.Fatalf("expected lifecycle status to be preserved, got %#v", first.PlayerLifecycle)
	}

	packet := BuildSessionFullPacket(snapshot, 5)
	if packet.Type != PacketFamilySessionFull {
		t.Fatalf("expected session full packet type, got %q", packet.Type)
	}
	if packet.Metadata.Lane != LaneSession || packet.Metadata.Sequence != 5 || packet.Metadata.BaselineID != "snapshot-1" || packet.Metadata.SnapshotID != "snapshot-1" || packet.Metadata.SnapshotKind != SnapshotKind("full") || packet.Metadata.ChunkIndex != 0 || packet.Metadata.ChunkCount != 1 || !packet.Metadata.IsFinalChunk {
		t.Fatalf("expected session metadata to be populated, got %#v", packet.Metadata)
	}
	if len(packet.Players) != 2 || packet.Players[0].ID != "player-a" || packet.Players[1].ID != "player-b" {
		t.Fatalf("expected packet players sorted by ID, got %#v", packet.Players)
	}
	if len(packet.PlayerLifecycle) != 2 || packet.PlayerLifecycle[0].PlayerID != "player-a" || packet.PlayerLifecycle[1].PlayerID != "player-b" {
		t.Fatalf("expected packet lifecycle sorted by ID, got %#v", packet.PlayerLifecycle)
	}

	if _, ok := snapshot.Players["player-a"]; !ok {
		t.Fatal("expected snapshot to retain player-a runtime state")
	}
	if snapshot.Players["player-a"].PrimaryCooldownRemaining != 99 || snapshot.Players["player-a"].PrimaryAmmoRemaining != 77 || snapshot.Players["player-a"].SecondaryCooldownRemaining != 55 || snapshot.Players["player-a"].SecondaryAmmoRemaining != 33 {
		t.Fatalf("expected player-a runtime state to remain untouched, got %#v", snapshot.Players["player-a"])
	}
	if snapshot.Players["player-b"].PrimaryCooldownRemaining != 88 || snapshot.Players["player-b"].PrimaryAmmoRemaining != 66 || snapshot.Players["player-b"].SecondaryCooldownRemaining != 44 || snapshot.Players["player-b"].SecondaryAmmoRemaining != 22 {
		t.Fatalf("expected player-b runtime state to remain untouched, got %#v", snapshot.Players["player-b"])
	}
}
