package realtime

import (
	"fmt"
	"reflect"
	"testing"

	game "github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func TestMovingShipDoesNotForceSuppressedAsteroidOrBulletLanes(t *testing.T) {
	previous, current := syncedMovingAsteroidSnapshots(300)
	previousBullets, currentBullets := syncedMovingBulletSnapshots(240)
	previous.Bullets = previousBullets.Bullets
	current.Bullets = currentBullets.Bullets
	previous.Players = map[string]runtime.ShipState{
		"player-1": {ID: "player-1", ShipType: "v_wing", X: 10, Y: 20, Health: 5},
	}
	current.Players = map[string]runtime.ShipState{
		"player-1": {ID: "player-1", ShipType: "v_wing", X: 11, Y: 20, Health: 5},
	}

	state := syncedWorldState(t, previous)
	state.HotLaneTick = 1
	prepared := state
	plan, err := assembleRealtimeLaneCandidates(current, state, &prepared)
	if err != nil {
		t.Fatalf("assemble realtime lane candidates: %v", err)
	}

	if _, ok := findCandidateByLane(plan.Candidates, LaneShips); !ok {
		t.Fatal("expected moving ship on the ship hot lane")
	}
	for _, lane := range []Lane{LaneWorld, LaneAsteroids, LaneBullets} {
		if _, ok := findCandidateByLane(plan.Candidates, lane); ok {
			t.Fatalf("ship movement forced unrelated lane %q", lane)
		}
	}
	if prepared.HotLaneCohorts.AsteroidMode != HotLaneModeFullOwned15Hz {
		t.Fatalf("asteroid mode = %q, want 15hz under four-or-more-chunk pressure", prepared.HotLaneCohorts.AsteroidMode)
	}
	if prepared.HotLaneCohorts.BulletMode != HotLaneModeFullOwned15Hz {
		t.Fatalf("bullet mode = %q, want 15hz under four-or-more-chunk pressure", prepared.HotLaneCohorts.BulletMode)
	}
}

func TestReliableWorldCommitDoesNotConsumeDeferredAsteroidMovement(t *testing.T) {
	previous, current := syncedMovingAsteroidSnapshots(300)
	previous.Players = map[string]runtime.ShipState{
		"player-1": {ID: "player-1", ShipType: "v_wing", X: 10, Y: 20, Health: 5},
	}
	current.Players = map[string]runtime.ShipState{
		"player-1": {ID: "player-1", ShipType: "v_wing", X: 10, Y: 20, Health: 4},
	}

	state := syncedWorldState(t, previous)
	previousWorld := mustWorldWireFull(t, previous, 1)
	seedHotLaneProjections(&state, previousWorld)
	state.HotLaneTick = 1

	prepared := state
	blocked, err := assembleRealtimeLaneCandidates(current, state, &prepared)
	if err != nil {
		t.Fatalf("assemble blocked cadence tick: %v", err)
	}
	if _, ok := findCandidateByLane(blocked.Candidates, LaneWorld); !ok {
		t.Fatal("expected reliable world projection commit for ship health change")
	}
	if _, ok := findCandidateByLane(blocked.Candidates, LaneShipsLifecycle); !ok {
		t.Fatal("expected reliable ship lifecycle update")
	}
	if _, ok := findCandidateByLane(blocked.Candidates, LaneAsteroids); ok {
		t.Fatal("asteroid hot lane escaped blocked 15hz cadence")
	}

	committed := prepared
	for _, candidate := range blocked.Candidates {
		if candidate.Lane() == LaneWorld || candidate.Lane() == LaneShipsLifecycle {
			CommitSuccessfulCandidate(&committed, candidate)
		}
	}
	committed.HotLaneTick = 4

	allowed, err := assembleRealtimeLaneCandidates(current, committed, nil)
	if err != nil {
		t.Fatalf("assemble allowed cadence tick: %v", err)
	}
	asteroidCandidate, ok := findCandidateByLane(allowed.Candidates, LaneAsteroids)
	if !ok {
		t.Fatal("reliable world commit consumed deferred asteroid movement")
	}
	packet, ok := asteroidCandidate.Payload.(AsteroidWireDeltaPacket)
	if !ok {
		t.Fatalf("asteroid candidate payload = %T, want AsteroidWireDeltaPacket", asteroidCandidate.Payload)
	}
	if len(packet.AsteroidUpdates) != len(current.Asteroids) {
		t.Fatalf("deferred asteroid updates = %d, want %d", len(packet.AsteroidUpdates), len(current.Asteroids))
	}
}

func TestChunkedHotProjectionCommitsOnlyAfterFinalChunk(t *testing.T) {
	previous := BulletHotLaneProjection{Records: []WorldBulletWireRecord{{ID: "bullet-1", X: 1, Y: 2, Rotation: 3}}}
	current := BulletHotLaneProjection{Records: []WorldBulletWireRecord{{ID: "bullet-1", X: 11, Y: 12, Rotation: 13}}}
	updates := make([]map[string]any, 0, 240)
	for index := 0; index < 240; index++ {
		updates = append(updates, map[string]any{
			"id":       fmt.Sprintf("bullet-%06d", index),
			"x":        index + 10,
			"y":        index + 20,
			"rotation": index + 30,
		})
	}

	candidate := mustRealtimeLaneCandidate(BulletWireDeltaPacket{
		Type: PacketFamilyBulletDelta,
		Metadata: Metadata{
			Lane:         LaneBullets,
			Sequence:     2,
			BaselineID:   "world-baseline",
			SnapshotID:   "bullets-snapshot-2",
			SnapshotKind: SnapshotKind("delta"),
		},
		BulletUpdates: updates,
	}, current)
	chunks := mustExpandRealtimeCandidateChunks([]RealtimeLaneCandidate{candidate})
	if len(chunks) < 2 {
		t.Fatalf("expected chunked bullet candidate, got %d chunk(s)", len(chunks))
	}

	state := NewRealtimeSessionState("player-1", "match-1")
	state.StoreBaselineProjection(LaneBullets, previous)
	CommitSuccessfulCandidate(&state, chunks[0])
	projection, _ := state.BaselineProjection(LaneBullets)
	if !reflect.DeepEqual(projection, previous) {
		t.Fatalf("non-final chunk committed projection: %#v", projection)
	}

	CommitSuccessfulCandidate(&state, chunks[len(chunks)-1])
	projection, _ = state.BaselineProjection(LaneBullets)
	if !reflect.DeepEqual(projection, current) {
		t.Fatalf("final chunk projection = %#v, want %#v", projection, current)
	}
}

func TestWorldFullCommitSeedsIndependentHotProjections(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		Players: map[string]runtime.ShipState{
			"player-1": {ID: "player-1", X: 1, Y: 2},
		},
		Asteroids: map[string]runtime.AsteroidState{
			"asteroid-1": {ID: "asteroid-1", X: 3, Y: 4},
		},
		Bullets: map[string]runtime.BulletState{
			"bullet-1": {ID: "bullet-1", X: 5, Y: 6, Rotation: 7},
		},
	}
	world := mustWorldWireFull(t, snapshot, 1)
	candidate := mustRealtimeLaneCandidate(world, world)
	state := NewRealtimeSessionState("player-1", "match-1")
	CommitSuccessfulCandidate(&state, candidate)

	for _, lane := range []Lane{LaneShips, LaneAsteroids, LaneBullets} {
		if _, ok := state.BaselineProjection(lane); !ok {
			t.Fatalf("world full commit did not seed lane %q projection", lane)
		}
	}
}
