package realtime

import "testing"

func TestCompareLaneRecordsEmitsCreateForMissingFromPrevious(t *testing.T) {
	delta := CompareLaneRecords(nil, []WorldShipRecord{{ID: "ship-a", ShipType: "v_wing"}}, func(record WorldShipRecord) string { return record.ID }, func(left, right WorldShipRecord) bool { return left == right })

	if len(delta.Creates) != 1 || delta.Creates[0].ID != "ship-a" {
		t.Fatalf("expected create for ship-a, got %#v", delta.Creates)
	}
	if len(delta.Updates) != 0 || len(delta.Deletes) != 0 {
		t.Fatalf("expected only a create, got %#v", delta)
	}
}

func TestCompareLaneRecordsEmitsUpdateForChangedRecord(t *testing.T) {
	previous := []WorldShipRecord{{ID: "ship-a", ShipType: "v_wing", X: 10}}
	current := []WorldShipRecord{{ID: "ship-a", ShipType: "v_wing", X: 11}}

	delta := CompareLaneRecords(previous, current, func(record WorldShipRecord) string { return record.ID }, func(left, right WorldShipRecord) bool { return left == right })

	if len(delta.Creates) != 0 {
		t.Fatalf("expected no creates, got %#v", delta.Creates)
	}
	if len(delta.Updates) != 1 || delta.Updates[0].ID != "ship-a" || delta.Updates[0].X != 11 {
		t.Fatalf("expected update for ship-a, got %#v", delta.Updates)
	}
	if len(delta.Deletes) != 0 {
		t.Fatalf("expected no deletes, got %#v", delta.Deletes)
	}
}

func TestCompareLaneRecordsEmitsNothingForUnchangedRecord(t *testing.T) {
	previous := []WorldShipRecord{{ID: "ship-a", ShipType: "v_wing", X: 10}}
	current := []WorldShipRecord{{ID: "ship-a", ShipType: "v_wing", X: 10}}

	delta := CompareLaneRecords(previous, current, func(record WorldShipRecord) string { return record.ID }, func(left, right WorldShipRecord) bool { return left == right })

	if len(delta.Creates) != 0 || len(delta.Updates) != 0 || len(delta.Deletes) != 0 {
		t.Fatalf("expected no delta for unchanged record, got %#v", delta)
	}
}

func TestCompareLaneRecordsEmitsDeleteForMissingFromCurrent(t *testing.T) {
	previous := []SessionPlayerRecord{{ID: "player-a", ShipType: "v_wing", Score: 5}, {ID: "player-b", ShipType: "v_wing", Score: 8}}
	current := []SessionPlayerRecord{{ID: "player-a", ShipType: "v_wing", Score: 5}}

	delta := CompareLaneRecords(previous, current, func(record SessionPlayerRecord) string { return record.ID }, func(left, right SessionPlayerRecord) bool { return left == right })

	if len(delta.Creates) != 0 || len(delta.Updates) != 0 {
		t.Fatalf("expected only a delete, got %#v", delta)
	}
	if len(delta.Deletes) != 1 || delta.Deletes[0] != "player-b" {
		t.Fatalf("expected delete for player-b, got %#v", delta.Deletes)
	}
}

func TestCompareLaneRecordsTreatsMissingFromDeltaAsUnchanged(t *testing.T) {
	previous := []SessionPlayerRecord{{ID: "player-a", ShipType: "v_wing", Score: 5}, {ID: "player-b", ShipType: "v_wing", Score: 8}}
	current := []SessionPlayerRecord{{ID: "player-a", ShipType: "v_wing", Score: 5}}

	delta := CompareLaneRecords(previous, current, func(record SessionPlayerRecord) string { return record.ID }, func(left, right SessionPlayerRecord) bool { return left == right })

	if len(delta.Creates) != 0 || len(delta.Updates) != 0 || len(delta.Deletes) != 1 || delta.Deletes[0] != "player-b" {
		t.Fatalf("expected missing-from-current to be the only delete, got %#v", delta)
	}
	if delta.Deletes[0] == "player-c" {
		t.Fatal("expected missing delta entity to remain unchanged, not deleted")
	}
}

func TestCompareLaneRecordsOrdersDeterministically(t *testing.T) {
	previous := []WorldShipRecord{{ID: "ship-c", ShipType: "v_wing"}}
	current := []WorldShipRecord{{ID: "ship-b", ShipType: "v_wing"}, {ID: "ship-a", ShipType: "v_wing"}}

	delta := CompareLaneRecords(previous, current, func(record WorldShipRecord) string { return record.ID }, func(left, right WorldShipRecord) bool { return left == right })

	if len(delta.Creates) != 2 || delta.Creates[0].ID != "ship-a" || delta.Creates[1].ID != "ship-b" {
		t.Fatalf("expected creates sorted by ID, got %#v", delta.Creates)
	}
	if len(delta.Deletes) != 1 || delta.Deletes[0] != "ship-c" {
		t.Fatalf("expected delete sorted deterministically, got %#v", delta.Deletes)
	}
}

func TestProjectionChangedReturnsFalseForEqualValues(t *testing.T) {
	previous := map[string]any{"type": "world_full", "count": 3}
	current := map[string]any{"type": "world_full", "count": 3}

	if ProjectionChanged(previous, current) {
		t.Fatal("expected equal values to report no change")
	}
}

func TestProjectionChangedReturnsTrueForDifferentValues(t *testing.T) {
	previous := map[string]any{"type": "world_full", "count": 3}
	current := map[string]any{"type": "world_full", "count": 4}

	if !ProjectionChanged(previous, current) {
		t.Fatal("expected different values to report change")
	}
}

func TestProjectionChangedReturnsTrueForNilPrevious(t *testing.T) {
	if !ProjectionChanged(nil, map[string]any{"type": "world_full"}) {
		t.Fatal("expected nil previous to report change")
	}
}

func TestProjectionChangedReturnsTrueForNilCurrent(t *testing.T) {
	if !ProjectionChanged(map[string]any{"type": "world_full"}, nil) {
		t.Fatal("expected nil current to report change")
	}
}

func TestBuildWorldDeltaPacketEmitsCreate(t *testing.T) {
	previous := WorldFullPacket{Type: PacketTypeWorldFull}
	current := WorldFullPacket{Type: PacketTypeWorldFull, Ships: []WorldShipRecord{{ID: "ship-a", ShipType: "v_wing"}}}

	delta := BuildWorldDeltaPacket(previous, current)

	if !WorldDeltaHasChanges(delta) {
		t.Fatal("expected world delta to report changes for create")
	}
	if len(delta.Ships.Creates) != 1 || delta.Ships.Creates[0].ID != "ship-a" {
		t.Fatalf("expected ship create, got %#v", delta.Ships.Creates)
	}
	if len(delta.Ships.Updates) != 0 || len(delta.Ships.Deletes) != 0 {
		t.Fatalf("expected only ship create, got %#v", delta.Ships)
	}
}

func TestBuildWorldDeltaPacketEmitsUpdate(t *testing.T) {
	previous := WorldFullPacket{Type: PacketTypeWorldFull, Bullets: []WorldBulletRecord{{ID: "bullet-a", X: 1, Y: 2}}}
	current := WorldFullPacket{Type: PacketTypeWorldFull, Bullets: []WorldBulletRecord{{ID: "bullet-a", X: 3, Y: 4}}}

	delta := BuildWorldDeltaPacket(previous, current)

	if !WorldDeltaHasChanges(delta) {
		t.Fatal("expected world delta to report changes for update")
	}
	if len(delta.Bullets.Updates) != 1 || delta.Bullets.Updates[0].ID != "bullet-a" || delta.Bullets.Updates[0].X != 3 || delta.Bullets.Updates[0].Y != 4 {
		t.Fatalf("expected bullet update, got %#v", delta.Bullets.Updates)
	}
	if len(delta.Bullets.Creates) != 0 || len(delta.Bullets.Deletes) != 0 {
		t.Fatalf("expected only bullet update, got %#v", delta.Bullets)
	}
}

func TestBuildWorldDeltaPacketEmitsDelete(t *testing.T) {
	previous := WorldFullPacket{Type: PacketTypeWorldFull, Asteroids: []WorldAsteroidRecord{{ID: "asteroid-a", Size: 2}}}
	current := WorldFullPacket{Type: PacketTypeWorldFull}

	delta := BuildWorldDeltaPacket(previous, current)

	if !WorldDeltaHasChanges(delta) {
		t.Fatal("expected world delta to report changes for delete")
	}
	if len(delta.Asteroids.Deletes) != 1 || delta.Asteroids.Deletes[0] != "asteroid-a" {
		t.Fatalf("expected asteroid delete, got %#v", delta.Asteroids.Deletes)
	}
	if len(delta.Asteroids.Creates) != 0 || len(delta.Asteroids.Updates) != 0 {
		t.Fatalf("expected only asteroid delete, got %#v", delta.Asteroids)
	}
}

func TestBuildWorldDeltaPacketReturnsNoChangesForIdenticalProjection(t *testing.T) {
	previous := WorldFullPacket{
		Type: PacketTypeWorldFull,
		Ships: []WorldShipRecord{{ID: "ship-a", ShipType: "v_wing", X: 1}},
		Bullets: []WorldBulletRecord{{ID: "bullet-a", OwnerID: "ship-a", X: 2}},
		Asteroids: []WorldAsteroidRecord{{ID: "asteroid-a", X: 3, Size: 1}},
		Pickups: []WorldPickupRecord{{ID: "pickup-a", Type: "shield", X: 4}},
	}
	current := previous

	delta := BuildWorldDeltaPacket(previous, current)

	if WorldDeltaHasChanges(delta) {
		t.Fatalf("expected no changes for identical projection, got %#v", delta)
	}
	if len(delta.Ships.Creates) != 0 || len(delta.Ships.Updates) != 0 || len(delta.Ships.Deletes) != 0 {
		t.Fatalf("expected no ship delta, got %#v", delta.Ships)
	}
	if len(delta.Bullets.Creates) != 0 || len(delta.Bullets.Updates) != 0 || len(delta.Bullets.Deletes) != 0 {
		t.Fatalf("expected no bullet delta, got %#v", delta.Bullets)
	}
	if len(delta.Asteroids.Creates) != 0 || len(delta.Asteroids.Updates) != 0 || len(delta.Asteroids.Deletes) != 0 {
		t.Fatalf("expected no asteroid delta, got %#v", delta.Asteroids)
	}
	if len(delta.Pickups.Creates) != 0 || len(delta.Pickups.Updates) != 0 || len(delta.Pickups.Deletes) != 0 {
		t.Fatalf("expected no pickup delta, got %#v", delta.Pickups)
	}
}

func TestBuildWorldDeltaPacketSetsDeltaSnapshotKind(t *testing.T) {
	previous := WorldFullPacket{Type: PacketTypeWorldFull}
	current := WorldFullPacket{Type: PacketTypeWorldFull, Metadata: Metadata{Lane: LaneWorld, Sequence: 9, BaselineID: "baseline-1", SnapshotID: "snapshot-1", SnapshotKind: SnapshotKind("full"), ServerSentMsec: 123}, Ships: []WorldShipRecord{{ID: "ship-a", ShipType: "v_wing"}}}

	delta := BuildWorldDeltaPacket(previous, current)

	if got, want := delta.Metadata.SnapshotKind, SnapshotKind("delta"); got != want {
		t.Fatalf("world delta snapshot kind = %q, want %q", got, want)
	}
	if got, want := delta.Metadata.Sequence, 9; got != want {
		t.Fatalf("world delta sequence = %d, want %d", got, want)
	}
}

func TestBuildOverlayDeltaPacketSetsDeltaSnapshotKind(t *testing.T) {
	previous := OverlayFullPacket{Type: PacketFamilyOverlayFull, Metadata: Metadata{Lane: LaneOverlay, Sequence: 4, BaselineID: "baseline-1", SnapshotID: "snapshot-1", SnapshotKind: SnapshotKind("full")}, Receiver: OverlayReceiverRecord{SelfID: "player-1"}}
	current := OverlayFullPacket{Type: PacketFamilyOverlayFull, Metadata: Metadata{Lane: LaneOverlay, Sequence: 5, BaselineID: "baseline-1", SnapshotID: "snapshot-2", SnapshotKind: SnapshotKind("full")}, Receiver: OverlayReceiverRecord{SelfID: "player-1", Lives: 3}}

	delta := BuildOverlayDeltaPacket(previous, current)

	if got, want := delta.Metadata.SnapshotKind, SnapshotKind("delta"); got != want {
		t.Fatalf("overlay delta snapshot kind = %q, want %q", got, want)
	}
	if got, want := delta.Metadata.Sequence, 5; got != want {
		t.Fatalf("overlay delta sequence = %d, want %d", got, want)
	}
}

func TestBuildSessionDeltaPacketSetsDeltaSnapshotKind(t *testing.T) {
	previous := SessionFullPacket{Type: PacketFamilySessionFull, Metadata: Metadata{Lane: LaneSession, Sequence: 7, BaselineID: "baseline-1", SnapshotID: "snapshot-1", SnapshotKind: SnapshotKind("full")}, Players: []SessionPlayerRecord{{ID: "player-a", ShipType: "v_wing"}}}
	current := SessionFullPacket{Type: PacketFamilySessionFull, Metadata: Metadata{Lane: LaneSession, Sequence: 8, BaselineID: "baseline-1", SnapshotID: "snapshot-2", SnapshotKind: SnapshotKind("full")}, Players: []SessionPlayerRecord{{ID: "player-a", ShipType: "v_wing", Score: 4}}}

	delta := BuildSessionDeltaPacket(previous, current)

	if got, want := delta.Metadata.SnapshotKind, SnapshotKind("delta"); got != want {
		t.Fatalf("session delta snapshot kind = %q, want %q", got, want)
	}
	if got, want := delta.Metadata.Sequence, 8; got != want {
		t.Fatalf("session delta sequence = %d, want %d", got, want)
	}
}

func TestBuildOverlayDeltaPacketEmitsChangedOverlayData(t *testing.T) {
	previous := OverlayFullPacket{Type: PacketFamilyOverlayFull, Receiver: OverlayReceiverRecord{SelfID: "player-1", Lives: 2, Score: 5, RespawnCooldown: 1.25}}
	current := OverlayFullPacket{Type: PacketFamilyOverlayFull, Receiver: OverlayReceiverRecord{SelfID: "player-1", Lives: 3, Score: 9, RespawnCooldown: 0.5}}

	delta := BuildOverlayDeltaPacket(previous, current)

	if !OverlayDeltaHasChanges(delta) {
		t.Fatal("expected overlay delta to report changes")
	}
	if len(delta.Receiver.Updates) != 1 {
		t.Fatalf("expected one overlay update, got %#v", delta.Receiver.Updates)
	}
	if got := delta.Receiver.Updates[0]; got.SelfID != "player-1" || got.Lives != 3 || got.Score != 9 || got.RespawnCooldown != 0.5 {
		t.Fatalf("expected changed overlay receiver to be returned, got %#v", got)
	}
}

func TestBuildOverlayDeltaPacketReturnsNoChangesForIdenticalOverlayData(t *testing.T) {
	current := OverlayFullPacket{Type: PacketFamilyOverlayFull, Receiver: OverlayReceiverRecord{SelfID: "player-1", Lives: 2, Score: 5, RespawnCooldown: 1.25}}

	delta := BuildOverlayDeltaPacket(current, current)

	if OverlayDeltaHasChanges(delta) {
		t.Fatalf("expected no overlay delta for identical projections, got %#v", delta)
	}
	if len(delta.Receiver.Creates) != 0 || len(delta.Receiver.Updates) != 0 || len(delta.Receiver.Deletes) != 0 {
		t.Fatalf("expected no overlay delta records, got %#v", delta.Receiver)
	}
}

func TestBuildSessionDeltaPacketEmitsChangedSessionData(t *testing.T) {
	previous := SessionFullPacket{
		Type: PacketFamilySessionFull,
		Metadata: Metadata{Lane: LaneSession, SnapshotID: "session-1"},
		Players: []SessionPlayerRecord{{ID: "player-a", ShipType: "v_wing", Score: 5, Lives: 3}},
		PlayerLifecycle: []SessionLifecycleRecord{{PlayerID: "player-a", Status: "active"}},
		TotalAsteroids: 4,
	}
	current := SessionFullPacket{
		Type: PacketFamilySessionFull,
		Metadata: Metadata{Lane: LaneSession, SnapshotID: "session-1"},
		Players: []SessionPlayerRecord{{ID: "player-a", ShipType: "v_wing", Score: 9, Lives: 2}},
		PlayerLifecycle: []SessionLifecycleRecord{{PlayerID: "player-a", Status: "respawning"}},
		TotalAsteroids: 7,
	}

	delta := BuildSessionDeltaPacket(previous, current)

	if !SessionDeltaHasChanges(delta) {
		t.Fatal("expected session delta to report changes")
	}
	if len(delta.Players.Updates) != 1 || delta.Players.Updates[0].Score != 9 || delta.Players.Updates[0].Lives != 2 {
		t.Fatalf("expected player update, got %#v", delta.Players)
	}
	if len(delta.PlayerLifecycle.Updates) != 1 || delta.PlayerLifecycle.Updates[0].Status != "respawning" {
		t.Fatalf("expected lifecycle update, got %#v", delta.PlayerLifecycle)
	}
	if len(delta.TotalAsteroids.Updates) != 1 || delta.TotalAsteroids.Updates[0].Count != 7 {
		t.Fatalf("expected total asteroid update, got %#v", delta.TotalAsteroids)
	}
}

func TestBuildSessionDeltaPacketReturnsNoChangesForIdenticalSessionData(t *testing.T) {
	current := SessionFullPacket{
		Type: PacketFamilySessionFull,
		Metadata: Metadata{Lane: LaneSession, SnapshotID: "session-1"},
		Players: []SessionPlayerRecord{{ID: "player-a", ShipType: "v_wing", Score: 5, Lives: 3}},
		PlayerLifecycle: []SessionLifecycleRecord{{PlayerID: "player-a", Status: "active"}},
		TotalAsteroids: 4,
	}

	delta := BuildSessionDeltaPacket(current, current)

	if SessionDeltaHasChanges(delta) {
		t.Fatalf("expected no session delta for identical projections, got %#v", delta)
	}
	if len(delta.Players.Creates) != 0 || len(delta.Players.Updates) != 0 || len(delta.Players.Deletes) != 0 {
		t.Fatalf("expected no player delta, got %#v", delta.Players)
	}
	if len(delta.PlayerLifecycle.Creates) != 0 || len(delta.PlayerLifecycle.Updates) != 0 || len(delta.PlayerLifecycle.Deletes) != 0 {
		t.Fatalf("expected no lifecycle delta, got %#v", delta.PlayerLifecycle)
	}
	if len(delta.TotalAsteroids.Creates) != 0 || len(delta.TotalAsteroids.Updates) != 0 || len(delta.TotalAsteroids.Deletes) != 0 {
		t.Fatalf("expected no total asteroid delta, got %#v", delta.TotalAsteroids)
	}
}

func TestBuildSessionDeltaPacketEmitsDeleteForMissingSessionRecord(t *testing.T) {
	previous := SessionFullPacket{
		Type: PacketFamilySessionFull,
		Metadata: Metadata{Lane: LaneSession, SnapshotID: "session-1"},
		Players: []SessionPlayerRecord{{ID: "player-a", ShipType: "v_wing", Score: 5, Lives: 3}, {ID: "player-b", ShipType: "v_wing", Score: 8, Lives: 2}},
		PlayerLifecycle: []SessionLifecycleRecord{{PlayerID: "player-a", Status: "active"}, {PlayerID: "player-b", Status: "active"}},
		TotalAsteroids: 4,
	}
	current := SessionFullPacket{
		Type: PacketFamilySessionFull,
		Metadata: Metadata{Lane: LaneSession, SnapshotID: "session-1"},
		Players: []SessionPlayerRecord{{ID: "player-a", ShipType: "v_wing", Score: 5, Lives: 3}},
		PlayerLifecycle: []SessionLifecycleRecord{{PlayerID: "player-a", Status: "active"}},
		TotalAsteroids: 4,
	}

	delta := BuildSessionDeltaPacket(previous, current)

	if !SessionDeltaHasChanges(delta) {
		t.Fatal("expected session delta to report delete changes")
	}
	if len(delta.Players.Deletes) != 1 || delta.Players.Deletes[0] != "player-b" {
		t.Fatalf("expected player delete, got %#v", delta.Players.Deletes)
	}
	if len(delta.PlayerLifecycle.Deletes) != 1 || delta.PlayerLifecycle.Deletes[0] != "player-b" {
		t.Fatalf("expected lifecycle delete, got %#v", delta.PlayerLifecycle.Deletes)
	}
}
