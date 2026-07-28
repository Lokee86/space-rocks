package realtime

import (
	"testing"

	game "github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func TestReceiverWorldWireProjectionFiltersSharedGeneration(t *testing.T) {
	shared := WorldWireFullPacket{
		Type: PacketFamilyWorldFull,
		Metadata: Metadata{
			Lane:           LaneWorld,
			ServerSentMsec: 10,
			SnapshotKind:   SnapshotKind("full"),
		},
		Ships:     []WorldShipWireRecord{{ID: "ship-a"}, {ID: "ship-b"}},
		Bullets:   []WorldBulletWireRecord{{ID: "bullet-a"}, {ID: "bullet-b"}},
		Asteroids: []WorldAsteroidWireRecord{{ID: "asteroid-a"}, {ID: "asteroid-b"}},
		Pickups:   []WorldPickupWireRecord{{ID: "pickup-a"}, {ID: "pickup-b"}},
	}
	snapshot := game.GameplayPresentationSnapshot{
		Players:        map[string]runtime.ShipState{"ship-b": {}},
		Bullets:        map[string]runtime.BulletState{"bullet-a": {}},
		Asteroids:      map[string]runtime.AsteroidState{"asteroid-b": {}},
		Pickups:        map[string]runtime.PickupState{"pickup-a": {}},
		ServerSentMsec: 99,
	}

	projection := receiverWorldWireProjection(shared, snapshot, 7)
	if len(projection.Ships) != 1 || projection.Ships[0].ID != "ship-b" {
		t.Fatalf("unexpected ships: %#v", projection.Ships)
	}
	if len(projection.Bullets) != 1 || projection.Bullets[0].ID != "bullet-a" {
		t.Fatalf("unexpected bullets: %#v", projection.Bullets)
	}
	if len(projection.Asteroids) != 1 || projection.Asteroids[0].ID != "asteroid-b" {
		t.Fatalf("unexpected asteroids: %#v", projection.Asteroids)
	}
	if len(projection.Pickups) != 1 || projection.Pickups[0].ID != "pickup-a" {
		t.Fatalf("unexpected pickups: %#v", projection.Pickups)
	}
	if projection.Metadata.Sequence != 7 || projection.Metadata.BaselineID != FullBaselineID(LaneWorld, 7) || projection.Metadata.SnapshotID != FullBaselineID(LaneWorld, 7) || projection.Metadata.ServerSentMsec != 99 {
		t.Fatalf("unexpected receiver metadata: %#v", projection.Metadata)
	}
}
