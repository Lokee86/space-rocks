package realtime

import (
	"fmt"
	"testing"

	game "github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func TestAssembleRealtimeLaneCandidatesUsesNextWorldSequenceForUnsyncedFull(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{SelfID: "player-1"}

	plan := mustAssembleRealtimeLaneCandidates(t, snapshot, NewRealtimeSessionState("player-1", "match-1"))
	candidate, ok := findCandidateByLane(plan.Candidates, LaneWorld)
	if !ok {
		t.Fatalf("expected world candidate")
	}
	full, ok := candidate.Payload.(WorldWireFullPacket)
	if !ok {
		t.Fatalf("world candidate full type = %T, want WorldWireFullPacket", candidate.Payload)
	}
	if got, want := full.Metadata.Sequence, 1; got != want {
		t.Fatalf("world full sequence = %d, want %d", got, want)
	}
}

func TestAssembleRealtimeLaneCandidatesEmitsWorldFullWhenNoBaseline(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID:  "player-1",
		Players: map[string]runtime.ShipState{"player-1": {ID: "player-1", ShipType: "v_wing"}},
	}

	state := NewRealtimeSessionState("player-1", "match-1")
	state.UpdateLane(LaneWorld, Metadata{Sequence: 1, IsFinalChunk: true})

	plan := mustAssembleRealtimeLaneCandidates(t, snapshot, state)
	world, ok := findCandidateByLane(plan.Candidates, LaneWorld)
	if !ok {
		t.Fatal("expected world candidate when no usable baseline exists")
	}
	if world.Kind() != RealtimeLaneCandidateKindFull {
		t.Fatalf("expected world full candidate, got kind=%q", world.Kind())
	}
	if _, ok := world.Payload.(WorldWireFullPacket); !ok {
		t.Fatalf("expected world full packet, got %T", world.Payload)
	}
}

func TestAssembleRealtimeLaneCandidatesOmitsWorldWhenStoredBaselineMatches(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID:  "player-1",
		Players: map[string]runtime.ShipState{"player-1": {ID: "player-1", ShipType: "v_wing"}},
	}

	state := NewRealtimeSessionState("player-1", "match-1")
	state.UpdateLane(LaneWorld, Metadata{Sequence: 1, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.StoreBaselineProjection(LaneWorld, mustWorldWireFull(t, snapshot, 1))

	plan := mustAssembleRealtimeLaneCandidates(t, snapshot, state)
	if _, ok := findCandidateByLane(plan.Candidates, LaneWorld); ok {
		t.Fatalf("expected no world candidate when stored baseline matches, got %#v", plan.Candidates)
	}
}

func TestAssembleRealtimeLaneCandidatesRoutesMovementOnlyShipChangesToHotLane(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID:  "player-1",
		Players: map[string]runtime.ShipState{"player-1": {ID: "player-1", ShipType: "v_wing", X: 2}},
	}

	state := NewRealtimeSessionState("player-1", "match-1")
	state.UpdateLane(LaneWorld, Metadata{Sequence: 2, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.StoreBaselineProjection(LaneWorld, mustWorldWireFull(t, game.GameplayPresentationSnapshot{SelfID: "player-1", Players: map[string]runtime.ShipState{"player-1": {ID: "player-1", ShipType: "v_wing", X: 1, Y: 0, Rotation: 0}}}, 1))

	plan := mustAssembleRealtimeLaneCandidates(t, snapshot, state)
	if _, ok := findCandidateByLane(plan.Candidates, LaneWorld); ok {
		t.Fatal("movement-only ship changes must not emit world delta")
	}
	ship, ok := findCandidateByLane(plan.Candidates, LaneShips)
	if !ok {
		t.Fatal("expected ship hot delta candidate")
	}
	if ship.Kind() != RealtimeLaneCandidateKindDelta {
		t.Fatalf("expected ship delta candidate, got kind=%q", ship.Kind())
	}
	if _, ok := ship.Payload.(ShipWireDeltaPacket); !ok {
		t.Fatalf("expected ship wire delta packet, got %T", ship.Payload)
	}
	if _, ok := ship.Projection.(ShipHotLaneProjection); !ok {
		t.Fatalf("expected ship hot projection, got %T", ship.Projection)
	}
}

func TestAssembleRealtimeLaneCandidatesCarriesIndependentAsteroidProjection(t *testing.T) {
	policy := DefaultHotLaneOffloadPolicy()
	count := policy.AsteroidHotLaneEntityBudget*2 + 1
	snapshot := game.GameplayPresentationSnapshot{SelfID: "player-1", Asteroids: map[string]runtime.AsteroidState{}}
	previous := game.GameplayPresentationSnapshot{SelfID: "player-1", Asteroids: map[string]runtime.AsteroidState{}}
	for i := 1; i <= count; i++ {
		id := fmt.Sprintf("asteroid-%d", i)
		snapshot.Asteroids[id] = runtime.AsteroidState{ID: id, X: float64(i + 100), Y: float64(i + 110)}
		previous.Asteroids[id] = runtime.AsteroidState{ID: id, X: float64(i), Y: float64(i + 10)}
	}

	state := NewRealtimeSessionState("player-1", "match-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 2, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.StoreBaselineProjection(LaneWorld, mustWorldWireFull(t, previous, 1))
	state.HotLaneTick = 12

	plan := mustAssembleRealtimeLaneCandidates(t, snapshot, state)
	if _, ok := findCandidateByLane(plan.Candidates, LaneWorld); ok {
		t.Fatal("movement-only asteroid changes must not emit world delta")
	}
	asteroid, ok := findCandidateByLane(plan.Candidates, LaneAsteroids)
	if !ok {
		t.Fatal("expected asteroid hot lane candidate")
	}
	projection, ok := asteroid.Projection.(AsteroidHotLaneProjection)
	if !ok {
		t.Fatalf("expected asteroid hot projection, got %T", asteroid.Projection)
	}
	if len(projection.Records) != count {
		t.Fatalf("asteroid projection records = %d, want %d", len(projection.Records), count)
	}
}

func TestAssembleRealtimeLaneCandidatesKeepsAsteroidLifecycleInWorldDeltaUnderPressure(t *testing.T) {
	policy := DefaultHotLaneOffloadPolicy()
	count := policy.AsteroidHotLaneEntityBudget*2 + 1
	snapshot := game.GameplayPresentationSnapshot{SelfID: "player-1", Asteroids: map[string]runtime.AsteroidState{}}
	previous := game.GameplayPresentationSnapshot{SelfID: "player-1", Asteroids: map[string]runtime.AsteroidState{}}
	for i := 1; i <= count; i++ {
		id := fmt.Sprintf("asteroid-%d", i)
		previous.Asteroids[id] = runtime.AsteroidState{ID: id, X: float64(i), Y: float64(i + 10)}
		snapshot.Asteroids[id] = runtime.AsteroidState{ID: id, X: float64(i + 100), Y: float64(i + 110)}
	}
	delete(snapshot.Asteroids, fmt.Sprintf("asteroid-%d", count))
	delete(previous.Asteroids, "asteroid-1")

	state := NewRealtimeSessionState("player-1", "match-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 2, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.StoreBaselineProjection(LaneWorld, mustWorldWireFull(t, previous, 1))
	state.HotLaneTick = 2

	plan := mustAssembleRealtimeLaneCandidates(t, snapshot, state)
	world, ok := findCandidateByLane(plan.Candidates, LaneWorld)
	if !ok {
		t.Fatal("expected world candidate under asteroid pressure")
	}
	worldDelta, ok := world.Payload.(WorldWireDeltaPacket)
	if !ok {
		t.Fatalf("expected world delta packet, got %T", world.Payload)
	}
	if len(worldDelta.Asteroids.Creates) != 0 || len(worldDelta.Asteroids.Deletes) != 0 {
		t.Fatalf("expected asteroid creates and deletes to move out of world delta, got %#v", worldDelta)
	}
	if len(worldDelta.Asteroids.Updates) != 0 {
		t.Fatalf("expected asteroid movement updates removed from world delta, got %d", len(worldDelta.Asteroids.Updates))
	}
	lifecycle, ok := findCandidateByLane(plan.Candidates, LaneAsteroidsLifecycle)
	if !ok {
		t.Fatal("expected asteroid lifecycle candidate under pressure")
	}
	lifecycleDelta, ok := lifecycle.Payload.(AsteroidWireDeltaPacket)
	if !ok {
		t.Fatalf("expected asteroid lifecycle packet, got %T", lifecycle.Payload)
	}
	if len(lifecycleDelta.AsteroidCreates) == 0 || len(lifecycleDelta.AsteroidDeletes) == 0 {
		t.Fatalf("expected asteroid lifecycle creates and deletes, got %#v", lifecycleDelta)
	}
	if _, ok := findCandidateByLane(plan.Candidates, LaneAsteroids); !ok {
		t.Fatal("expected asteroid hot delta candidate under pressure")
	}
}

func TestAssembleRealtimeLaneCandidatesMovesBulletLifecycleOutOfWorldDeltaUnderPressure(t *testing.T) {
	policy := DefaultHotLaneOffloadPolicy()
	count := policy.BulletHotLaneEntityBudget*2 + 1
	snapshot := game.GameplayPresentationSnapshot{SelfID: "player-1", Bullets: map[string]runtime.BulletState{}}
	previous := game.GameplayPresentationSnapshot{SelfID: "player-1", Bullets: map[string]runtime.BulletState{}}
	for i := 1; i <= count; i++ {
		id := fmt.Sprintf("bullet-%d", i)
		previous.Bullets[id] = runtime.BulletState{ID: id, OwnerID: "player-1", X: float64(i), Y: float64(i + 20), Rotation: float64(i + 30), WeaponID: "laser", ProjectileType: "bolt"}
		snapshot.Bullets[id] = runtime.BulletState{ID: id, OwnerID: "player-1", X: float64(i + 100), Y: float64(i + 120), Rotation: float64(i + 130), WeaponID: "laser", ProjectileType: "bolt"}
	}
	delete(snapshot.Bullets, fmt.Sprintf("bullet-%d", count))
	delete(previous.Bullets, "bullet-1")

	state := NewRealtimeSessionState("player-1", "match-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 2, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.StoreBaselineProjection(LaneWorld, mustWorldWireFull(t, previous, 1))
	state.HotLaneTick = 12

	plan := mustAssembleRealtimeLaneCandidates(t, snapshot, state)
	world, ok := findCandidateByLane(plan.Candidates, LaneWorld)
	if !ok {
		t.Fatal("expected world candidate under bullet pressure")
	}
	worldDelta, ok := world.Payload.(WorldWireDeltaPacket)
	if !ok {
		t.Fatalf("expected world delta packet, got %T", world.Payload)
	}
	if len(worldDelta.Bullets.Creates) != 0 || len(worldDelta.Bullets.Deletes) != 0 {
		t.Fatalf("expected bullet creates and deletes to move out of world delta, got %#v", worldDelta)
	}
	if len(worldDelta.Bullets.Updates) != 0 {
		t.Fatalf("expected bullet movement updates removed from world delta, got %d", len(worldDelta.Bullets.Updates))
	}
	lifecycle, ok := findCandidateByLane(plan.Candidates, LaneBulletsLifecycle)
	if !ok {
		t.Fatal("expected bullet lifecycle candidate under pressure")
	}
	lifecycleDelta, ok := lifecycle.Payload.(BulletWireDeltaPacket)
	if !ok {
		t.Fatalf("expected bullet lifecycle packet, got %T", lifecycle.Payload)
	}
	if len(lifecycleDelta.BulletCreates) == 0 || len(lifecycleDelta.BulletDeletes) == 0 {
		t.Fatalf("expected bullet lifecycle creates and deletes, got %#v", lifecycleDelta)
	}
	if _, ok := findCandidateByLane(plan.Candidates, LaneBullets); !ok {
		t.Fatal("expected bullet hot delta candidate under pressure")
	}
}

func TestAssembleRealtimeLaneCandidatesDoesNotEmitEmptyHotCandidate(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{SelfID: "player-1"}
	state := NewRealtimeSessionState("player-1", "match-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 1, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.StoreBaselineProjection(LaneWorld, mustWorldWireFull(t, snapshot, 1))

	plan := mustAssembleRealtimeLaneCandidates(t, snapshot, state)
	if _, ok := findCandidateByLane(plan.Candidates, LaneAsteroids); ok {
		t.Fatal("unexpected asteroid hot candidate with no offloaded updates")
	}
	if _, ok := findCandidateByLane(plan.Candidates, LaneBullets); ok {
		t.Fatal("unexpected bullet hot candidate with no offloaded updates")
	}
}

func TestAssembleRealtimeLaneCandidatesEmitsAsteroidLifecycleCandidateWhenAsteroidsAreCreatedOrDeleted(t *testing.T) {
	previous := game.GameplayPresentationSnapshot{SelfID: "player-1"}
	current := game.GameplayPresentationSnapshot{SelfID: "player-1", Asteroids: map[string]runtime.AsteroidState{"asteroid-a": {ID: "asteroid-a", X: 10, Y: 20, Size: 3, Health: 4, Scale: 5, Variant: 6}}}

	state := NewRealtimeSessionState("player-1", "match-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 1, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.StoreBaselineProjection(LaneWorld, mustWorldWireFull(t, previous, 1))

	plan := mustAssembleRealtimeLaneCandidates(t, current, state)
	candidate, ok := findCandidateByLane(plan.Candidates, LaneAsteroidsLifecycle)
	if !ok {
		t.Fatal("expected asteroid lifecycle candidate")
	}
	if candidate.Kind() != RealtimeLaneCandidateKindDelta {
		t.Fatalf("expected asteroid lifecycle delta candidate kind, got %q", candidate.Kind())
	}
	delta, ok := candidate.Payload.(AsteroidWireDeltaPacket)
	if !ok {
		t.Fatalf("expected asteroid lifecycle packet, got %T", candidate.Payload)
	}
	if len(delta.AsteroidCreates) != 1 || delta.AsteroidCreates[0].ID != "asteroid-a" {
		t.Fatalf("expected asteroid create to move to lifecycle lane, got %#v", delta.AsteroidCreates)
	}
	if len(delta.AsteroidDeletes) != 0 {
		t.Fatalf("expected no asteroid deletes, got %#v", delta.AsteroidDeletes)
	}
	if len(delta.AsteroidUpdates) != 0 {
		t.Fatalf("expected lifecycle lane to omit asteroid hot updates, got %#v", delta.AsteroidUpdates)
	}
	if got, want := delta.Type, PacketFamilyAsteroidsLifecycle; got != want {
		t.Fatalf("expected asteroid lifecycle packet type %q, got %q", want, got)
	}
}

func TestAssembleRealtimeLaneCandidatesEmitsUnchunkedAsteroidsAt60HzWithoutWorldDelta(t *testing.T) {
	previous, current := syncedMovingAsteroidSnapshots(1)
	for _, tick := range []int{1, 2} {
		state := syncedWorldState(t, previous)
		state.HotLaneTick = tick
		plan := mustAssembleRealtimeLaneCandidates(t, current, state)
		if _, ok := findCandidateByLane(plan.Candidates, LaneAsteroids); !ok {
			t.Fatalf("tick %d omitted 60hz asteroid hot candidate", tick)
		}
		if _, ok := findCandidateByLane(plan.Candidates, LaneWorld); ok {
			t.Fatalf("tick %d emitted world delta for movement-only asteroid changes", tick)
		}
	}
}

func TestAssembleRealtimeLaneCandidatesFloorsHeavyAsteroidsAt15HzAndBurstsAllChunks(t *testing.T) {
	count := 300
	previous, current := syncedMovingAsteroidSnapshots(count)
	for _, tick := range []int{1, 2, 3, 4} {
		state := syncedWorldState(t, previous)
		state.HotLaneTick = tick
		plan := mustAssembleRealtimeLaneCandidates(t, current, state)
		candidate, hasHot := findCandidateByLane(plan.Candidates, LaneAsteroids)
		if hasWorld := func() bool { _, ok := findCandidateByLane(plan.Candidates, LaneWorld); return ok }(); hasWorld {
			t.Fatalf("tick %d emitted world delta for movement-only asteroid pressure", tick)
		}
		if tick < 4 && hasHot {
			t.Fatalf("tick %d emitted 15hz asteroid hot candidate early", tick)
		}
		if tick == 4 {
			if !hasHot {
				t.Fatal("fourth tick omitted 15hz asteroid hot candidate")
			}
			chunks := mustExpandRealtimeCandidateChunks([]RealtimeLaneCandidate{candidate})
			if len(chunks) < 4 {
				t.Fatalf("heavy asteroid candidate expanded to %d chunks, want at least 4", len(chunks))
			}
			for index, chunk := range chunks {
				metadata, _ := chunk.Metadata()
				if metadata.Sequence != candidate.Payload.(AsteroidWireDeltaPacket).Metadata.Sequence || metadata.ChunkIndex != index || metadata.ChunkCount != len(chunks) {
					t.Fatalf("chunk %d metadata = %#v, want same-sequence parallel burst", index, metadata)
				}
			}
		}
	}
}

func TestAssembleRealtimeLaneCandidatesKeepsAsteroidLifecycleIndependentFromHotCadence(t *testing.T) {
	previous, current := syncedMovingAsteroidSnapshots(300)
	delete(current.Asteroids, "asteroid-300")
	current.Asteroids["asteroid-301"] = runtime.AsteroidState{ID: "asteroid-301", X: 320, Y: 330, Size: 2, Health: 3, Scale: 1, Variant: 1}

	blockedState := syncedWorldState(t, previous)
	blockedState.HotLaneTick = 1
	blocked := mustAssembleRealtimeLaneCandidates(t, current, blockedState)
	for _, lane := range []Lane{LaneAsteroidsLifecycle, LaneWorld} {
		if _, ok := findCandidateByLane(blocked.Candidates, lane); !ok {
			t.Fatalf("expected reliable lane %q on blocked hot tick", lane)
		}
	}
	if _, ok := findCandidateByLane(blocked.Candidates, LaneAsteroids); ok {
		t.Fatal("asteroid hot lane escaped blocked cadence tick")
	}

	allowedState := syncedWorldState(t, previous)
	allowedState.HotLaneTick = 4
	allowed := mustAssembleRealtimeLaneCandidates(t, current, allowedState)
	for _, lane := range []Lane{LaneAsteroidsLifecycle, LaneWorld, LaneAsteroids} {
		if _, ok := findCandidateByLane(allowed.Candidates, lane); !ok {
			t.Fatalf("expected lane %q on permitted hot tick", lane)
		}
	}
}

func TestAssembleRealtimeLaneCandidatesEmitsLifecycleWithoutAsteroidHotMovement(t *testing.T) {
	previous := game.GameplayPresentationSnapshot{SelfID: "player-1", Asteroids: map[string]runtime.AsteroidState{}}
	current := previous
	current.Asteroids = map[string]runtime.AsteroidState{"asteroid-1": {ID: "asteroid-1", X: 10, Y: 20, Size: 2, Health: 3, Scale: 1, Variant: 1}}
	state := syncedWorldState(t, previous)
	state.HotLaneTick = 1
	plan := mustAssembleRealtimeLaneCandidates(t, current, state)
	if _, ok := findCandidateByLane(plan.Candidates, LaneAsteroidsLifecycle); !ok {
		t.Fatal("expected asteroid lifecycle candidate")
	}
	if _, ok := findCandidateByLane(plan.Candidates, LaneAsteroids); ok {
		t.Fatal("did not expect asteroid hot candidate")
	}
	world, ok := findCandidateByLane(plan.Candidates, LaneWorld)
	if !ok || world.Projection == nil {
		t.Fatalf("expected world candidate with projection, got %#v", world)
	}
}

func TestAssembleRealtimeLaneCandidatesAdjustsBulletCadenceByChunkPressure(t *testing.T) {
	for _, tc := range []struct {
		name       string
		target     int
		count      int
		period     int
		wantChunks int
		minimum    bool
	}{
		{name: "one chunk 60hz", target: 1, period: 1, wantChunks: 1},
		{name: "two chunks 30hz", target: 2, period: 2, wantChunks: 2},
		{name: "three chunks 20hz", target: 3, period: 3, wantChunks: 3},
		{name: "four or more chunks 15hz", count: 240, period: 4, wantChunks: 4, minimum: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			count := tc.count
			if count == 0 {
				count = bulletUpdateCountForChunkTarget(t, tc.target)
			}
			previous, current := syncedMovingBulletSnapshots(count)
			for tick := 1; tick <= 12; tick++ {
				state := syncedWorldState(t, previous)
				state.HotLaneTick = tick
				preparedState := state
				plan, err := assembleRealtimeLaneCandidates(current, state, &preparedState)
				if err != nil {
					t.Fatalf("assemble realtime lane candidates: %v", err)
				}
				candidate, hasHot := findCandidateByLane(plan.Candidates, LaneBullets)
				if _, hasWorld := findCandidateByLane(plan.Candidates, LaneWorld); hasWorld {
					t.Fatalf("tick %d emitted world delta for movement-only bullets", tick)
				}
				want := tick%tc.period == 0
				if hasHot != want {
					t.Fatalf("tick %d hot=%t want=%t period=%d mode=%q count=%d", tick, hasHot, want, tc.period, preparedState.HotLaneCohorts.BulletMode, count)
				}
				if want {
					got := len(mustExpandRealtimeCandidateChunks([]RealtimeLaneCandidate{candidate}))
					if tc.minimum {
						if got < tc.wantChunks {
							t.Fatalf("tick %d expanded chunks=%d, want at least %d", tick, got, tc.wantChunks)
						}
					} else if got != tc.wantChunks {
						t.Fatalf("tick %d expanded chunks=%d, want %d", tick, got, tc.wantChunks)
					}
				}
			}
		})
	}
}

func syncedWorldState(t *testing.T, snapshot game.GameplayPresentationSnapshot) RealtimeSessionState {
	t.Helper()
	state := NewRealtimeSessionState("player-1", "match-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 1, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.StoreBaselineProjection(LaneWorld, mustWorldWireFull(t, snapshot, 1))
	return state
}

func syncedMovingAsteroidSnapshots(count int) (game.GameplayPresentationSnapshot, game.GameplayPresentationSnapshot) {
	previous := game.GameplayPresentationSnapshot{SelfID: "player-1", Asteroids: map[string]runtime.AsteroidState{}}
	current := game.GameplayPresentationSnapshot{SelfID: "player-1", Asteroids: map[string]runtime.AsteroidState{}}
	for i := 1; i <= count; i++ {
		id := fmt.Sprintf("asteroid-%d", i)
		previous.Asteroids[id] = runtime.AsteroidState{ID: id, X: float64(i), Y: float64(i + 10), Size: 2, Health: 3, Scale: 1, Variant: 1}
		current.Asteroids[id] = runtime.AsteroidState{ID: id, X: float64(i + 1), Y: float64(i + 11), Size: 2, Health: 3, Scale: 1, Variant: 1}
	}
	return previous, current
}

func syncedMovingBulletSnapshots(count int) (game.GameplayPresentationSnapshot, game.GameplayPresentationSnapshot) {
	previous := game.GameplayPresentationSnapshot{SelfID: "player-1", Bullets: map[string]runtime.BulletState{}}
	current := game.GameplayPresentationSnapshot{SelfID: "player-1", Bullets: map[string]runtime.BulletState{}}
	for i := 1; i <= count; i++ {
		id := fmt.Sprintf("bullet-%d", i)
		previous.Bullets[id] = runtime.BulletState{ID: id, OwnerID: "player-1", X: float64(i), Y: float64(i + 10), Rotation: float64(i + 20), WeaponID: "laser", ProjectileType: "bolt"}
		current.Bullets[id] = runtime.BulletState{ID: id, OwnerID: "player-1", X: float64(i + 1), Y: float64(i + 11), Rotation: float64(i + 21), WeaponID: "laser", ProjectileType: "bolt"}
	}
	return previous, current
}

func bulletUpdateCountForChunkTarget(t *testing.T, target int) int {
	t.Helper()
	for count := 1; count <= 600; count++ {
		previous, current := syncedMovingBulletSnapshots(count)
		delta := BuildWorldWireDeltaPacket(mustWorldWireFull(t, previous, 1), mustWorldWireFull(t, current, 2))
		split := SplitWorldHotUpdates(delta, NewHotLaneCohortState(), DefaultHotLaneOffloadPolicy())
		if split.BulletDelta == nil {
			continue
		}
		metadata := split.BulletDelta.Metadata
		metadata.Lane = LaneBullets
		metadata.Sequence = 1
		metadata.SnapshotKind = SnapshotKind("delta")
		metadata = metadata.WithChunk(0, 1)
		split.BulletDelta.Metadata = metadata
		if bulletWireDeltaChunkCount(*split.BulletDelta) == target {
			return count
		}
	}
	t.Fatalf("no bullet update count produced exactly %d chunks", target)
	return 0
}
