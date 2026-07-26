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

func TestAssembleRealtimeLaneCandidatesEmitsWorldDeltaWhenStoredBaselineDiffers(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID:  "player-1",
		Players: map[string]runtime.ShipState{"player-1": {ID: "player-1", ShipType: "v_wing", X: 2}},
	}

	state := NewRealtimeSessionState("player-1", "match-1")
	state.UpdateLane(LaneWorld, Metadata{Sequence: 2, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.StoreBaselineProjection(LaneWorld, mustWorldWireFull(t, game.GameplayPresentationSnapshot{SelfID: "player-1", Players: map[string]runtime.ShipState{"player-1": {ID: "player-1", ShipType: "v_wing", X: 1, Y: 0, Rotation: 0}}}, 1))

	plan := mustAssembleRealtimeLaneCandidates(t, snapshot, state)
	world, ok := findCandidateByLane(plan.Candidates, LaneWorld)
	if !ok {
		t.Fatal("expected world delta candidate when stored baseline differs")
	}
	if world.Kind() != RealtimeLaneCandidateKindDelta {
		t.Fatalf("expected world delta candidate, got kind=%q", world.Kind())
	}
	if _, ok := world.Payload.(WorldWireDeltaPacket); !ok {
		t.Fatalf("expected world delta packet, got %T", world.Payload)
	}
	if _, ok := world.Projection.(WorldWireFullPacket); !ok {
		t.Fatalf("expected current world full projection to be carried, got %T", world.Projection)
	}
}

func TestAssembleRealtimeLaneCandidatesUsesFullWorldProjectionAfterHotSplit(t *testing.T) {
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
	state.HotLaneTick = 2

	plan := mustAssembleRealtimeLaneCandidates(t, snapshot, state)
	world, ok := findCandidateByLane(plan.Candidates, LaneWorld)
	if !ok {
		t.Fatal("expected world candidate")
	}
	delta, ok := world.Payload.(WorldWireDeltaPacket)
	if !ok {
		t.Fatalf("expected world delta packet, got %T", world.Payload)
	}
	if len(delta.Asteroids.Updates) != 0 {
		t.Fatalf("expected world asteroid updates removed, got %d", len(delta.Asteroids.Updates))
	}
	projection, ok := world.Projection.(WorldWireFullPacket)
	if !ok {
		t.Fatalf("expected world projection packet, got %T", world.Projection)
	}
	if len(projection.Asteroids) != count {
		t.Fatalf("expected full projection to remain at %d asteroids, got %d", count, len(projection.Asteroids))
	}
	if _, ok := findCandidateByLane(plan.Candidates, LaneAsteroids); !ok {
		t.Fatal("expected asteroid hot lane candidate")
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

func TestAssembleRealtimeLaneCandidatesHonorsAsteroidHotCadence(t *testing.T) {
	previous, current := syncedMovingAsteroidSnapshots(1)
	for _, tick := range []int{1, 2} {
		state := syncedWorldState(t, previous)
		state.HotLaneTick = tick
		plan := mustAssembleRealtimeLaneCandidates(t, current, state)
		_, hasHot := findCandidateByLane(plan.Candidates, LaneAsteroids)
		_, hasWorld := findCandidateByLane(plan.Candidates, LaneWorld)
		if !hasHot || !hasWorld {
			t.Fatalf("tick %d omitted hot/world candidates", tick)
		}
	}
}

func TestAssembleRealtimeLaneCandidatesThrottlesChunkedAsteroidsTo30Hz(t *testing.T) {
	count := 300
	previous, current := syncedMovingAsteroidSnapshots(count)
	for _, tick := range []int{1, 2, 3} {
		state := syncedWorldState(t, previous)
		state.HotLaneTick = tick
		plan := mustAssembleRealtimeLaneCandidates(t, current, state)
		_, hasHot := findCandidateByLane(plan.Candidates, LaneAsteroids)
		_, hasWorld := findCandidateByLane(plan.Candidates, LaneWorld)
		if (tick == 1 || tick == 3) && (hasHot || hasWorld) {
			t.Fatalf("tick %d unexpectedly emitted hot/world candidates", tick)
		}
		if tick == 2 && (!hasHot || !hasWorld) {
			t.Fatalf("tick %d omitted hot/world candidates", tick)
		}
		if tick == 2 {
			candidate, _ := findCandidateByLane(plan.Candidates, LaneAsteroids)
			if len(mustExpandRealtimeCandidateChunks([]RealtimeLaneCandidate{candidate})) <= 1 {
				t.Fatal("expected asteroid candidate to expand into multiple chunks")
			}
		}
	}
}

func TestAssembleRealtimeLaneCandidatesKeepsChunkedAsteroidLifecycleOnCadence(t *testing.T) {
	previous, current := syncedMovingAsteroidSnapshots(300)
	delete(current.Asteroids, "asteroid-300")
	current.Asteroids["asteroid-301"] = runtime.AsteroidState{ID: "asteroid-301", X: 320, Y: 330, Size: 2, Health: 3, Scale: 1, Variant: 1}

	blockedState := syncedWorldState(t, previous)
	blockedState.HotLaneTick = 1
	blocked := mustAssembleRealtimeLaneCandidates(t, current, blockedState)
	for _, lane := range []Lane{LaneAsteroidsLifecycle, LaneAsteroids, LaneWorld} {
		if _, ok := findCandidateByLane(blocked.Candidates, lane); ok {
			t.Fatalf("lane %q escaped the blocked asteroid cadence tick", lane)
		}
	}

	allowedState := syncedWorldState(t, previous)
	allowedState.HotLaneTick = 2
	allowed := mustAssembleRealtimeLaneCandidates(t, current, allowedState)
	for _, lane := range []Lane{LaneAsteroidsLifecycle, LaneAsteroids, LaneWorld} {
		if _, ok := findCandidateByLane(allowed.Candidates, lane); !ok {
			t.Fatalf("expected lane %q on the permitted asteroid cadence tick", lane)
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

func TestAssembleRealtimeLaneCandidatesHonorsBulletChunkCadence(t *testing.T) {
	for _, tc := range []struct {
		name       string
		target     int
		wantChunks int
	}{
		{name: "one chunk", target: 1, wantChunks: 1},
		{name: "two chunks", target: 2, wantChunks: 2},
		{name: "three chunks", target: 3, wantChunks: 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			count := bulletUpdateCountForChunkTarget(t, tc.target)
			previous, current := syncedMovingBulletSnapshots(count)
			for _, tick := range []int{1, 2, 3} {
				state := syncedWorldState(t, previous)
				state.HotLaneTick = tick
				plan := mustAssembleRealtimeLaneCandidates(t, current, state)
				_, hasHot := findCandidateByLane(plan.Candidates, LaneBullets)
				_, hasWorld := findCandidateByLane(plan.Candidates, LaneWorld)
				want := tc.target == 1 || tick == tc.target
				if hasHot != want || hasWorld != want {
					t.Fatalf("tick %d hot=%t world=%t want=%t", tick, hasHot, hasWorld, want)
				}
				if want {
					candidate, _ := findCandidateByLane(plan.Candidates, LaneBullets)
					if got := len(mustExpandRealtimeCandidateChunks([]RealtimeLaneCandidate{candidate})); got != tc.wantChunks {
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
