package realtime

import (
	"fmt"
	"testing"
)

func TestSplitWorldHotUpdatesRoutesLowPressureUpdatesToHotLanes(t *testing.T) {
	worldDelta := WorldWireDeltaPacket{
		Type:     PacketTypeWorldDelta,
		Metadata: Metadata{Lane: LaneWorld, Sequence: 7},
		Asteroids: FieldRecordDelta[WorldAsteroidWireRecord]{Updates: []map[string]any{{"id": "asteroid-1"}, {"id": "asteroid-2"}}},
		Bullets:   FieldRecordDelta[WorldBulletWireRecord]{Updates: []map[string]any{{"id": "bullet-1"}}},
	}

	result := SplitWorldHotUpdates(worldDelta, NewHotLaneCohortState(), DefaultHotLaneOffloadPolicy())

	if got := len(result.WorldDelta.Asteroids.Updates); got != 0 {
		t.Fatalf("world asteroid updates = %d, want 0", got)
	}
	if got := len(result.WorldDelta.Bullets.Updates); got != 0 {
		t.Fatalf("world bullet updates = %d, want 0", got)
	}
	if result.AsteroidDelta == nil || len(result.AsteroidDelta.AsteroidUpdates) != 2 {
		t.Fatalf("expected asteroid delta with 2 updates, got %#v", result.AsteroidDelta)
	}
	if result.BulletDelta == nil || len(result.BulletDelta.BulletUpdates) != 1 {
		t.Fatalf("expected bullet delta with 1 update, got %#v", result.BulletDelta)
	}
	if got := result.CohortState.AsteroidRoutes["asteroid-1"]; got != HotUpdateRouteAsteroids {
		t.Fatalf("asteroid-1 route = %q, want asteroids", got)
	}
	if got := result.CohortState.BulletRoutes["bullet-1"]; got != HotUpdateRouteBullets {
		t.Fatalf("bullet-1 route = %q, want bullets", got)
	}
}

func TestSplitWorldHotUpdatesExistingAsteroidRouteRemainsStickyAcrossTicks(t *testing.T) {
	cohort := NewHotLaneCohortState()
	cohort.AsteroidRoutes["asteroid-1"] = HotUpdateRouteAsteroids
	worldDelta := WorldWireDeltaPacket{Type: PacketTypeWorldDelta, Metadata: Metadata{Lane: LaneWorld, Sequence: 30}, Asteroids: FieldRecordDelta[WorldAsteroidWireRecord]{Updates: []map[string]any{{"id": "asteroid-1"}}}}
	result := SplitWorldHotUpdates(worldDelta, cohort, DefaultHotLaneOffloadPolicy())
	if got := result.CohortState.AsteroidRoutes["asteroid-1"]; got != HotUpdateRouteAsteroids { t.Fatalf("asteroid-1 route = %q, want asteroids", got) }
}

func TestSplitWorldHotUpdatesExistingBulletRouteRemainsStickyAcrossTicks(t *testing.T) {
	cohort := NewHotLaneCohortState()
	cohort.BulletRoutes["bullet-1"] = HotUpdateRouteBullets
	worldDelta := WorldWireDeltaPacket{Type: PacketTypeWorldDelta, Metadata: Metadata{Lane: LaneWorld, Sequence: 31}, Bullets: FieldRecordDelta[WorldBulletWireRecord]{Updates: []map[string]any{{"id": "bullet-1"}}}}
	result := SplitWorldHotUpdates(worldDelta, cohort, DefaultHotLaneOffloadPolicy())
	if got := result.CohortState.BulletRoutes["bullet-1"]; got != HotUpdateRouteBullets { t.Fatalf("bullet-1 route = %q, want bullets", got) }
}

func TestSplitWorldHotUpdatesDeletedRouteIsRemoved(t *testing.T) {
	cohort := NewHotLaneCohortState()
	cohort.AsteroidRoutes["asteroid-1"] = HotUpdateRouteAsteroids
	worldDelta := WorldWireDeltaPacket{Type: PacketTypeWorldDelta, Metadata: Metadata{Lane: LaneWorld, Sequence: 32}}
	result := SplitWorldHotUpdates(worldDelta, cohort, DefaultHotLaneOffloadPolicy())
	if _, exists := result.CohortState.AsteroidRoutes["asteroid-1"]; exists { t.Fatalf("expected missing asteroid route to be removed") }
}

func TestSplitWorldHotUpdatesNewEntityAssignmentIsDeterministic(t *testing.T) {
	worldDelta := WorldWireDeltaPacket{Type: PacketTypeWorldDelta, Metadata: Metadata{Lane: LaneWorld, Sequence: 33}, Asteroids: FieldRecordDelta[WorldAsteroidWireRecord]{Updates: []map[string]any{{"id": "asteroid-9"}, {"id": "asteroid-1"}}}}
	result := SplitWorldHotUpdates(worldDelta, NewHotLaneCohortState(), DefaultHotLaneOffloadPolicy())
	if got := result.CohortState.AsteroidRoutes["asteroid-1"]; got != HotUpdateRouteAsteroids { t.Fatalf("asteroid-1 route = %q, want asteroids", got) }
	if got := result.CohortState.AsteroidRoutes["asteroid-9"]; got != HotUpdateRouteAsteroids { t.Fatalf("asteroid-9 route = %q, want asteroids", got) }
	if len(result.WorldDelta.Asteroids.Updates) != 0 { t.Fatalf("expected asteroid updates removed from world delta, got %d", len(result.WorldDelta.Asteroids.Updates)) }
	if result.AsteroidDelta == nil || len(result.AsteroidDelta.AsteroidUpdates) != 2 { t.Fatalf("expected asteroid delta with 2 updates, got %#v", result.AsteroidDelta) }
}

func TestSplitWorldHotUpdatesDoesNotReshuffleRoutesOnInputOrderChange(t *testing.T) {
	cohort := NewHotLaneCohortState()
	cohort.AsteroidRoutes["asteroid-1"] = HotUpdateRouteWorld
	cohort.AsteroidRoutes["asteroid-2"] = HotUpdateRouteWorld
	worldDelta := WorldWireDeltaPacket{Type: PacketTypeWorldDelta, Metadata: Metadata{Lane: LaneWorld, Sequence: 34}, Asteroids: FieldRecordDelta[WorldAsteroidWireRecord]{Updates: []map[string]any{{"id": "asteroid-2"}, {"id": "asteroid-1"}}}, Bullets: FieldRecordDelta[WorldBulletWireRecord]{Updates: []map[string]any{{"id": "bullet-2"}, {"id": "bullet-1"}}}}
	result := SplitWorldHotUpdates(worldDelta, cohort, DefaultHotLaneOffloadPolicy())
	if got := result.CohortState.AsteroidRoutes["asteroid-1"]; got != HotUpdateRouteAsteroids { t.Fatalf("asteroid-1 route = %q, want asteroids", got) }
	if got := result.CohortState.AsteroidRoutes["asteroid-2"]; got != HotUpdateRouteAsteroids { t.Fatalf("asteroid-2 route = %q, want asteroids", got) }
	if got := result.CohortState.BulletRoutes["bullet-1"]; got != HotUpdateRouteBullets { t.Fatalf("bullet-1 route = %q, want bullets", got) }
	if got := result.CohortState.BulletRoutes["bullet-2"]; got != HotUpdateRouteBullets { t.Fatalf("bullet-2 route = %q, want bullets", got) }
	if len(result.WorldDelta.Asteroids.Updates) != 0 { t.Fatalf("expected asteroid updates removed from world delta, got %d", len(result.WorldDelta.Asteroids.Updates)) }
	if len(result.WorldDelta.Bullets.Updates) != 0 { t.Fatalf("expected bullet updates removed from world delta, got %d", len(result.WorldDelta.Bullets.Updates)) }
}

func TestSplitWorldHotUpdatesAsteroidFullOwnedModeThresholds(t *testing.T) {
	policy := DefaultHotLaneOffloadPolicy()
	for _, tc := range []struct {
		name  string
		count int
		want  HotLaneMode
	}{
		{name: "single update", count: 1, want: HotLaneModeFullOwned30Hz},
		{name: "double budget", count: policy.AsteroidHotLaneEntityBudget * 2, want: HotLaneModeFullOwned30Hz},
		{name: "double budget plus one", count: policy.AsteroidHotLaneEntityBudget*2 + 1, want: HotLaneModeFullOwned20Hz},
		{name: "triple budget", count: policy.AsteroidHotLaneEntityBudget * 3, want: HotLaneModeFullOwned20Hz},
		{name: "triple budget plus one", count: policy.AsteroidHotLaneEntityBudget*3 + 1, want: HotLaneModeNeedsChunking},
	} {
		t.Run(tc.name, func(t *testing.T) {
			worldDelta := WorldWireDeltaPacket{Type: PacketTypeWorldDelta, Metadata: Metadata{Lane: LaneWorld, Sequence: 40 + tc.count}, Asteroids: FieldRecordDelta[WorldAsteroidWireRecord]{Updates: makeHotUpdates("asteroid", tc.count, false)}}
			result := SplitWorldHotUpdates(worldDelta, NewHotLaneCohortState(), policy)
			if result.AsteroidMode != tc.want { t.Fatalf("AsteroidMode = %q, want %q", result.AsteroidMode, tc.want) }
			if got := len(result.WorldDelta.Asteroids.Updates); got != 0 { t.Fatalf("world asteroid updates = %d, want 0", got) }
		})
	}
}

func TestSplitWorldHotUpdatesBulletFullOwnedModeThresholds(t *testing.T) {
	policy := DefaultHotLaneOffloadPolicy()
	for _, tc := range []struct {
		name  string
		count int
		want  HotLaneMode
	}{
		{name: "single update", count: 1, want: HotLaneModeFullOwned30Hz},
		{name: "double budget", count: policy.BulletHotLaneEntityBudget * 2, want: HotLaneModeFullOwned30Hz},
		{name: "double budget plus one", count: policy.BulletHotLaneEntityBudget*2 + 1, want: HotLaneModeFullOwned20Hz},
		{name: "triple budget", count: policy.BulletHotLaneEntityBudget * 3, want: HotLaneModeFullOwned20Hz},
		{name: "triple budget plus one", count: policy.BulletHotLaneEntityBudget*3 + 1, want: HotLaneModeNeedsChunking},
	} {
		t.Run(tc.name, func(t *testing.T) {
			worldDelta := WorldWireDeltaPacket{Type: PacketTypeWorldDelta, Metadata: Metadata{Lane: LaneWorld, Sequence: 80 + tc.count}, Bullets: FieldRecordDelta[WorldBulletWireRecord]{Updates: makeHotUpdates("bullet", tc.count, true)}}
			result := SplitWorldHotUpdates(worldDelta, NewHotLaneCohortState(), policy)
			if result.BulletMode != tc.want { t.Fatalf("BulletMode = %q, want %q", result.BulletMode, tc.want) }
			if got := len(result.WorldDelta.Bullets.Updates); got != 0 { t.Fatalf("world bullet updates = %d, want 0", got) }
		})
	}
}

func TestSplitWorldHotUpdatesOwnsAsteroidsAndBulletsIndependently(t *testing.T) {
	policy := DefaultHotLaneOffloadPolicy()
	worldDelta := WorldWireDeltaPacket{Type: PacketTypeWorldDelta, Metadata: Metadata{Lane: LaneWorld, Sequence: 23}, Asteroids: FieldRecordDelta[WorldAsteroidWireRecord]{Updates: makeHotUpdates("asteroid", 100, false)}, Bullets: FieldRecordDelta[WorldBulletWireRecord]{Updates: makeHotUpdates("bullet", 93, true)}}
	result := SplitWorldHotUpdates(worldDelta, NewHotLaneCohortState(), policy)
	if got := len(result.WorldDelta.Asteroids.Updates); got != 0 { t.Fatalf("world asteroid updates = %d, want 0", got) }
	if got := len(result.WorldDelta.Bullets.Updates); got != 0 { t.Fatalf("world bullet updates = %d, want 0", got) }
	if result.AsteroidDelta == nil || len(result.AsteroidDelta.AsteroidUpdates) != 100 { t.Fatalf("expected asteroid delta with 100 updates, got %#v", result.AsteroidDelta) }
	if result.BulletDelta == nil || len(result.BulletDelta.BulletUpdates) != 93 { t.Fatalf("expected bullet delta with 93 updates, got %#v", result.BulletDelta) }
	if result.AsteroidMode != hotLaneModeForFullOwnedCount(100, policy.AsteroidHotLaneEntityBudget) { t.Fatalf("AsteroidMode = %q, want derived full-owned mode", result.AsteroidMode) }
	if result.BulletMode != hotLaneModeForFullOwnedCount(93, policy.BulletHotLaneEntityBudget) { t.Fatalf("BulletMode = %q, want derived full-owned mode", result.BulletMode) }
}


func TestSplitWorldHotUpdatesMovesStressMovementUpdatesOutOfWorld(t *testing.T) {
	worldDelta := WorldWireDeltaPacket{
		Type:      PacketTypeWorldDelta,
		Metadata:  Metadata{Lane: LaneWorld, Sequence: 24},
		Asteroids: FieldRecordDelta[WorldAsteroidWireRecord]{Updates: makeHotUpdates("asteroid", 80, false)},
		Bullets:   FieldRecordDelta[WorldBulletWireRecord]{Updates: makeHotUpdates("bullet", 80, true)},
	}

	result := SplitWorldHotUpdates(worldDelta, NewHotLaneCohortState(), DefaultHotLaneOffloadPolicy())

	if got := len(result.WorldDelta.Asteroids.Updates); got != 0 { t.Fatalf("world asteroid updates = %d, want 0", got) }
	if got := len(result.WorldDelta.Bullets.Updates); got != 0 { t.Fatalf("world bullet updates = %d, want 0", got) }
	if result.AsteroidDelta == nil { t.Fatalf("expected asteroid delta to be non-nil") }
	if got := len(result.AsteroidDelta.AsteroidUpdates); got != 80 { t.Fatalf("asteroid delta updates = %d, want 80", got) }
	if result.BulletDelta == nil { t.Fatalf("expected bullet delta to be non-nil") }
	if got := len(result.BulletDelta.BulletUpdates); got != 80 { t.Fatalf("bullet delta updates = %d, want 80", got) }
	if result.AsteroidOffloaded != 80 { t.Fatalf("asteroid offloaded = %d, want 80", result.AsteroidOffloaded) }
	if result.BulletOffloaded != 80 { t.Fatalf("bullet offloaded = %d, want 80", result.BulletOffloaded) }
}

func TestSplitWorldHotUpdatesOverridesStaleWorldRoutesUnderStress(t *testing.T) {
	cohort := NewHotLaneCohortState()
	for _, update := range makeHotUpdates("asteroid", 80, false) {
		cohort.AsteroidRoutes[update["id"].(string)] = HotUpdateRouteWorld
	}
	for _, update := range makeHotUpdates("bullet", 80, true) {
		cohort.BulletRoutes[update["id"].(string)] = HotUpdateRouteWorld
	}

	worldDelta := WorldWireDeltaPacket{
		Type:      PacketTypeWorldDelta,
		Metadata:  Metadata{Lane: LaneWorld, Sequence: 25},
		Asteroids: FieldRecordDelta[WorldAsteroidWireRecord]{Updates: makeHotUpdates("asteroid", 80, false)},
		Bullets:   FieldRecordDelta[WorldBulletWireRecord]{Updates: makeHotUpdates("bullet", 80, true)},
	}

	result := SplitWorldHotUpdates(worldDelta, cohort, DefaultHotLaneOffloadPolicy())

	for _, update := range worldDelta.Asteroids.Updates {
		if got := result.CohortState.AsteroidRoutes[update["id"].(string)]; got != HotUpdateRouteAsteroids {
			t.Fatalf("asteroid route = %q, want asteroids", got)
		}
	}
	for _, update := range worldDelta.Bullets.Updates {
		if got := result.CohortState.BulletRoutes[update["id"].(string)]; got != HotUpdateRouteBullets {
			t.Fatalf("bullet route = %q, want bullets", got)
		}
	}
	if got := len(result.WorldDelta.Asteroids.Updates); got != 0 { t.Fatalf("world asteroid updates = %d, want 0", got) }
	if got := len(result.WorldDelta.Bullets.Updates); got != 0 { t.Fatalf("world bullet updates = %d, want 0", got) }
}

func TestSplitWorldHotUpdatesLifecycleSafetyKeepsCreatesAndDeletesInWorld(t *testing.T) {
	worldDelta := WorldWireDeltaPacket{
		Type:     PacketTypeWorldDelta,
		Metadata: Metadata{Lane: LaneWorld, Sequence: 35},
		Asteroids: FieldRecordDelta[WorldAsteroidWireRecord]{
			Creates: []WorldAsteroidWireRecord{{ID: "asteroid-create"}},
			Deletes: []string{"asteroid-delete"},
			Updates: []map[string]any{{"id": "asteroid-1"}},
		},
		Bullets: FieldRecordDelta[WorldBulletWireRecord]{
			Creates: []WorldBulletWireRecord{{ID: "bullet-create"}},
			Deletes: []string{"bullet-delete"},
			Updates: []map[string]any{{"id": "bullet-1"}},
		},
	}

	result := SplitWorldHotUpdates(worldDelta, NewHotLaneCohortState(), DefaultHotLaneOffloadPolicy())

	if len(result.WorldDelta.Asteroids.Creates) != 1 || len(result.WorldDelta.Asteroids.Deletes) != 1 {
		t.Fatalf("asteroid creates/deletes were not preserved in world delta: %#v", result.WorldDelta.Asteroids)
	}
	if len(result.WorldDelta.Bullets.Creates) != 1 || len(result.WorldDelta.Bullets.Deletes) != 1 {
		t.Fatalf("bullet creates/deletes were not preserved in world delta: %#v", result.WorldDelta.Bullets)
	}
	if len(result.WorldDelta.Asteroids.Updates) != 0 || len(result.WorldDelta.Bullets.Updates) != 0 {
		t.Fatalf("movement updates should be removed from world delta: %#v %#v", result.WorldDelta.Asteroids.Updates, result.WorldDelta.Bullets.Updates)
	}
}

func TestSplitWorldHotUpdatesKeepsLifecycleInWorld(t *testing.T) {
	worldDelta := WorldWireDeltaPacket{
		Type:     PacketTypeWorldDelta,
		Metadata: Metadata{Lane: LaneWorld, Sequence: 36},
		Asteroids: FieldRecordDelta[WorldAsteroidWireRecord]{
			Creates: []WorldAsteroidWireRecord{{ID: "asteroid-create-1"}},
			Updates: []map[string]any{{"id": "asteroid-update-1"}},
			Deletes: []string{"asteroid-delete-1"},
		},
		Bullets: FieldRecordDelta[WorldBulletWireRecord]{
			Creates: []WorldBulletWireRecord{{ID: "bullet-create-1"}},
			Updates: []map[string]any{{"id": "bullet-update-1"}},
			Deletes: []string{"bullet-delete-1"},
		},
	}

	result := SplitWorldHotUpdates(worldDelta, NewHotLaneCohortState(), DefaultHotLaneOffloadPolicy())

	if got := len(result.WorldDelta.Asteroids.Creates); got != 1 {
		t.Fatalf("asteroid creates = %d, want 1", got)
	}
	if got := len(result.WorldDelta.Asteroids.Deletes); got != 1 {
		t.Fatalf("asteroid deletes = %d, want 1", got)
	}
	if got := len(result.WorldDelta.Bullets.Creates); got != 1 {
		t.Fatalf("bullet creates = %d, want 1", got)
	}
	if got := len(result.WorldDelta.Bullets.Deletes); got != 1 {
		t.Fatalf("bullet deletes = %d, want 1", got)
	}
	if got := len(result.WorldDelta.Asteroids.Updates); got != 0 {
		t.Fatalf("asteroid updates = %d, want 0", got)
	}
	if got := len(result.WorldDelta.Bullets.Updates); got != 0 {
		t.Fatalf("bullet updates = %d, want 0", got)
	}
	if result.AsteroidDelta == nil || len(result.AsteroidDelta.AsteroidUpdates) != 1 {
		t.Fatalf("expected asteroid delta with 1 update, got %#v", result.AsteroidDelta)
	}
	if result.BulletDelta == nil || len(result.BulletDelta.BulletUpdates) != 1 {
		t.Fatalf("expected bullet delta with 1 update, got %#v", result.BulletDelta)
	}
}

func TestSplitWorldHotUpdatesNeverRetainsMovementUpdatesInWorld(t *testing.T) {
	worldDelta := WorldWireDeltaPacket{
		Type:     PacketTypeWorldDelta,
		Metadata: Metadata{Lane: LaneWorld, Sequence: 37},
		Asteroids: FieldRecordDelta[WorldAsteroidWireRecord]{Updates: []map[string]any{{"id": "asteroid-1"}, {"id": "asteroid-2"}, {"id": "asteroid-3"}}},
		Bullets:   FieldRecordDelta[WorldBulletWireRecord]{Updates: []map[string]any{{"id": "bullet-1"}, {"id": "bullet-2"}, {"id": "bullet-3"}}},
		Ships:     FieldRecordDelta[WorldShipWireRecord]{Updates: []map[string]any{{"id": "ship-1"}}},
		Pickups:   FieldRecordDelta[WorldPickupWireRecord]{Updates: []map[string]any{{"id": "pickup-1"}}},
	}

	result := SplitWorldHotUpdates(worldDelta, NewHotLaneCohortState(), DefaultHotLaneOffloadPolicy())

	if got := len(result.WorldDelta.Asteroids.Updates); got != 0 {
		t.Fatalf("asteroid updates = %d, want 0", got)
	}
	if got := len(result.WorldDelta.Bullets.Updates); got != 0 {
		t.Fatalf("bullet updates = %d, want 0", got)
	}
	if got := len(result.WorldDelta.Ships.Updates); got != 1 {
		t.Fatalf("ship updates = %d, want 1", got)
	}
	if got := len(result.WorldDelta.Pickups.Updates); got != 1 {
		t.Fatalf("pickup updates = %d, want 1", got)
	}
	if result.AsteroidDelta == nil || len(result.AsteroidDelta.AsteroidUpdates) != 3 {
		t.Fatalf("expected asteroid delta with 3 updates, got %#v", result.AsteroidDelta)
	}
	if result.BulletDelta == nil || len(result.BulletDelta.BulletUpdates) != 3 {
		t.Fatalf("expected bullet delta with 3 updates, got %#v", result.BulletDelta)
	}
}


func makeHotUpdates(kind string, count int, includeRotation bool) []map[string]any {
	updates := make([]map[string]any, 0, count)
	for i := count; i >= 1; i-- {
		update := map[string]any{"id": fmt.Sprintf("%s-%d", kind, i), "x": int64(i), "y": int64(i + 100)}
		if includeRotation { update["rotation"] = int64(i + 200) }
		updates = append(updates, update)
	}
	return updates
}
