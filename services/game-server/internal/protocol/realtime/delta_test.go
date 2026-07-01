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

func TestCompareLaneRecordFieldsHandlesWorldBulletCreatesUpdatesDeletes(t *testing.T) {
	previous := []WorldBulletRecord{{ID: "bullet-a", X: 1, Y: 2, WeaponID: "pulse"}}
	current := []WorldBulletRecord{{ID: "bullet-a", X: 1, Y: 2, WeaponID: "pulse"}, {ID: "bullet-b", X: 4, Y: 5, WeaponID: "laser"}}

	delta := CompareLaneRecordFields(previous, current, func(record WorldBulletRecord) string { return record.ID }, "id")

	if len(delta.Creates) != 1 || delta.Creates[0].ID != "bullet-b" {
		t.Fatalf("expected bullet-b create, got %#v", delta.Creates)
	}
	if len(delta.Updates) != 0 {
		t.Fatalf("expected no updates for unchanged bullets, got %#v", delta.Updates)
	}
	if len(delta.Deletes) != 0 {
		t.Fatalf("expected no deletes, got %#v", delta.Deletes)
	}
}

func TestCompareLaneRecordFieldsEmitsDeleteForMissingWorldBullet(t *testing.T) {
	previous := []WorldBulletRecord{{ID: "bullet-a", X: 1, Y: 2, WeaponID: "pulse"}}
	current := []WorldBulletRecord{}

	delta := CompareLaneRecordFields(previous, current, func(record WorldBulletRecord) string { return record.ID }, "id")

	if len(delta.Creates) != 0 || len(delta.Updates) != 0 {
		t.Fatalf("expected only delete, got %#v", delta)
	}
	if len(delta.Deletes) != 1 || delta.Deletes[0] != "bullet-a" {
		t.Fatalf("expected bullet-a delete, got %#v", delta.Deletes)
	}
}

func TestCompareLaneRecordFieldsEmitsUpdateForChangedWorldBulletX(t *testing.T) {
	previous := []WorldBulletRecord{{ID: "bullet-a", X: 1, Y: 2, WeaponID: "pulse"}}
	current := []WorldBulletRecord{{ID: "bullet-a", X: 3, Y: 2, WeaponID: "pulse"}}

	delta := CompareLaneRecordFields(previous, current, func(record WorldBulletRecord) string { return record.ID }, "id")

	if len(delta.Creates) != 0 || len(delta.Deletes) != 0 {
		t.Fatalf("expected only update, got %#v", delta)
	}
	if len(delta.Updates) != 1 {
		t.Fatalf("expected one update, got %#v", delta.Updates)
	}
	if got := delta.Updates[0]; got["id"] != "bullet-a" || len(got) != 2 || got["x"] != float64(3) {
		t.Fatalf("expected id and x only, got %#v", got)
	}
}

func TestCompareLaneRecordFieldsIncludesZeroValueBulletRotationChange(t *testing.T) {
	previous := []WorldBulletRecord{{ID: "bullet-a", X: 1, Y: 2, Rotation: 9, WeaponID: "pulse"}}
	current := []WorldBulletRecord{{ID: "bullet-a", X: 1, Y: 2, Rotation: 0, WeaponID: "pulse"}}

	delta := CompareLaneRecordFields(previous, current, func(record WorldBulletRecord) string { return record.ID }, "id")

	if len(delta.Updates) != 1 {
		t.Fatalf("expected one update, got %#v", delta.Updates)
	}
	if got := delta.Updates[0]; got["id"] != "bullet-a" || len(got) != 2 || got["rotation"] != float64(0) {
		t.Fatalf("expected id and zero rotation only, got %#v", got)
	}
}

func TestCompareLaneRecordFieldsEmitsUpdateForChangedWorldBulletWeaponID(t *testing.T) {
	previous := []WorldBulletRecord{{ID: "bullet-a", X: 1, Y: 2, WeaponID: "pulse"}}
	current := []WorldBulletRecord{{ID: "bullet-a", X: 1, Y: 2, WeaponID: "laser"}}

	delta := CompareLaneRecordFields(previous, current, func(record WorldBulletRecord) string { return record.ID }, "id")

	if len(delta.Creates) != 0 || len(delta.Deletes) != 0 {
		t.Fatalf("expected only update, got %#v", delta)
	}
	if len(delta.Updates) != 1 {
		t.Fatalf("expected one update, got %#v", delta.Updates)
	}
	if got := delta.Updates[0]; got["id"] != "bullet-a" || len(got) != 2 || got["weapon_id"] != "laser" {
		t.Fatalf("expected id and weapon_id only, got %#v", got)
	}
}

func TestCompareLaneRecordFieldsOrdersUpdatesDeterministically(t *testing.T) {
	previous := []WorldBulletRecord{
		{ID: "bullet-c", X: 1, Y: 2, WeaponID: "pulse"},
		{ID: "bullet-b", X: 4, Y: 2, WeaponID: "pulse"},
		{ID: "bullet-a", X: 3, Y: 2, WeaponID: "laser"},
	}
	current := []WorldBulletRecord{
		{ID: "bullet-b", X: 5, Y: 2, WeaponID: "pulse"},
		{ID: "bullet-a", X: 6, Y: 2, WeaponID: "laser"},
	}

	delta := CompareLaneRecordFields(previous, current, func(record WorldBulletRecord) string { return record.ID }, "id")

	if len(delta.Creates) != 0 {
		t.Fatalf("expected no creates, got %#v", delta.Creates)
	}
	if len(delta.Updates) != 2 {
		t.Fatalf("expected two updates, got %#v", delta.Updates)
	}
	if delta.Updates[0]["id"] != "bullet-a" || delta.Updates[1]["id"] != "bullet-b" {
		t.Fatalf("expected updates sorted by id, got %#v", delta.Updates)
	}
	if len(delta.Deletes) != 1 || delta.Deletes[0] != "bullet-c" {
		t.Fatalf("expected delete for bullet-c, got %#v", delta.Deletes)
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
	previous := WorldFullPacket{Type: PacketTypeWorldFull, Bullets: []WorldBulletRecord{{ID: "bullet-a", X: 1, Y: 2, Rotation: 3, OwnerID: "ship-a", WeaponID: "pulse", ProjectileType: "laser"}}}
	current := WorldFullPacket{Type: PacketTypeWorldFull, Bullets: []WorldBulletRecord{{ID: "bullet-a", X: 3, Y: 4, Rotation: 5, OwnerID: "ship-a", WeaponID: "pulse", ProjectileType: "laser"}}}

	delta := BuildWorldDeltaPacket(previous, current)

	if !WorldDeltaHasChanges(delta) {
		t.Fatal("expected world delta to report changes for update")
	}
	if len(delta.Bullets.Updates) != 1 {
		t.Fatalf("expected one bullet update, got %#v", delta.Bullets.Updates)
	}
	if got := delta.Bullets.Updates[0]; got["id"] != "bullet-a" || len(got) != 4 || got["x"] != float64(3) || got["y"] != float64(4) || got["rotation"] != float64(5) {
		t.Fatalf("expected partial bullet update, got %#v", got)
	}
	if _, ok := delta.Bullets.Updates[0]["owner_id"]; ok {
		t.Fatalf("expected owner_id to be omitted, got %#v", delta.Bullets.Updates[0])
	}
	if _, ok := delta.Bullets.Updates[0]["weapon_id"]; ok {
		t.Fatalf("expected weapon_id to be omitted, got %#v", delta.Bullets.Updates[0])
	}
	if _, ok := delta.Bullets.Updates[0]["projectile_type"]; ok {
		t.Fatalf("expected projectile_type to be omitted, got %#v", delta.Bullets.Updates[0])
	}
	if len(delta.Bullets.Creates) != 0 || len(delta.Bullets.Deletes) != 0 {
		t.Fatalf("expected only bullet update, got %#v", delta.Bullets)
	}
}

func TestBuildWorldDeltaPacketEmitsBulletCreateAsTypedRecord(t *testing.T) {
	previous := WorldFullPacket{Type: PacketTypeWorldFull}
	current := WorldFullPacket{Type: PacketTypeWorldFull, Bullets: []WorldBulletRecord{{ID: "bullet-a", X: 1, Y: 2, Rotation: 3, OwnerID: "ship-a", WeaponID: "pulse", ProjectileType: "laser"}}}

	delta := BuildWorldDeltaPacket(previous, current)

	if len(delta.Bullets.Creates) != 1 || delta.Bullets.Creates[0].ID != "bullet-a" {
		t.Fatalf("expected bullet create, got %#v", delta.Bullets.Creates)
	}
	if len(delta.Bullets.Updates) != 0 || len(delta.Bullets.Deletes) != 0 {
		t.Fatalf("expected only bullet create, got %#v", delta.Bullets)
	}
}

func TestBuildWorldDeltaPacketEmitsBulletDeleteAsID(t *testing.T) {
	previous := WorldFullPacket{Type: PacketTypeWorldFull, Bullets: []WorldBulletRecord{{ID: "bullet-a", X: 1, Y: 2, Rotation: 3, OwnerID: "ship-a", WeaponID: "pulse", ProjectileType: "laser"}}}
	current := WorldFullPacket{Type: PacketTypeWorldFull}

	delta := BuildWorldDeltaPacket(previous, current)

	if len(delta.Bullets.Deletes) != 1 || delta.Bullets.Deletes[0] != "bullet-a" {
		t.Fatalf("expected bullet delete, got %#v", delta.Bullets.Deletes)
	}
	if len(delta.Bullets.Creates) != 0 || len(delta.Bullets.Updates) != 0 {
		t.Fatalf("expected only bullet delete, got %#v", delta.Bullets)
	}
}

func TestBuildWorldDeltaPacketEmitsAsteroidCreatesAsTypedRecords(t *testing.T) {
	previous := WorldFullPacket{Type: PacketTypeWorldFull}
	current := WorldFullPacket{Type: PacketTypeWorldFull, Asteroids: []WorldAsteroidRecord{{ID: "asteroid-a", X: 1, Y: 2, Size: 3, Health: 4, Scale: 5, Variant: 1}}}

	delta := BuildWorldDeltaPacket(previous, current)

	if len(delta.Asteroids.Creates) != 1 || delta.Asteroids.Creates[0].ID != "asteroid-a" || delta.Asteroids.Creates[0].Size != 3 {
		t.Fatalf("expected typed asteroid create, got %#v", delta.Asteroids.Creates)
	}
	if len(delta.Asteroids.Updates) != 0 || len(delta.Asteroids.Deletes) != 0 {
		t.Fatalf("expected only asteroid create, got %#v", delta.Asteroids)
	}
}

func TestBuildWorldDeltaPacketEmitsAsteroidUpdatesAsFieldDeltas(t *testing.T) {
	previous := WorldFullPacket{Type: PacketTypeWorldFull, Asteroids: []WorldAsteroidRecord{{ID: "asteroid-a", X: 1, Y: 2, Size: 3, Health: 4, Scale: 5, Variant: 1}}}
	current := WorldFullPacket{Type: PacketTypeWorldFull, Asteroids: []WorldAsteroidRecord{{ID: "asteroid-a", X: 8, Y: 9, Size: 3, Health: 4, Scale: 5, Variant: 1}}}

	delta := BuildWorldDeltaPacket(previous, current)

	if len(delta.Asteroids.Updates) != 1 {
		t.Fatalf("expected one asteroid update, got %#v", delta.Asteroids.Updates)
	}
	got := delta.Asteroids.Updates[0]
	if got["id"] != "asteroid-a" || len(got) != 3 || got["x"] != float64(8) || got["y"] != float64(9) {
		t.Fatalf("expected id, x, and y only, got %#v", got)
	}
	if _, ok := got["size"]; ok {
		t.Fatalf("expected size to be omitted, got %#v", got)
	}
	if _, ok := got["health"]; ok {
		t.Fatalf("expected health to be omitted, got %#v", got)
	}
	if _, ok := got["scale"]; ok {
		t.Fatalf("expected scale to be omitted, got %#v", got)
	}
	if _, ok := got["variant"]; ok {
		t.Fatalf("expected variant to be omitted, got %#v", got)
	}
}

func TestBuildWorldDeltaPacketEmitsAsteroidDeletesAsIDs(t *testing.T) {
	previous := WorldFullPacket{Type: PacketTypeWorldFull, Asteroids: []WorldAsteroidRecord{{ID: "asteroid-a", X: 1, Y: 2, Size: 3, Health: 4, Scale: 5, Variant: 1}}}
	current := WorldFullPacket{Type: PacketTypeWorldFull}

	delta := BuildWorldDeltaPacket(previous, current)

	if len(delta.Asteroids.Deletes) != 1 || delta.Asteroids.Deletes[0] != "asteroid-a" {
		t.Fatalf("expected asteroid delete by ID, got %#v", delta.Asteroids.Deletes)
	}
	if len(delta.Asteroids.Creates) != 0 || len(delta.Asteroids.Updates) != 0 {
		t.Fatalf("expected only asteroid delete, got %#v", delta.Asteroids)
	}
}

func TestBuildWorldDeltaPacketEmitsPickupCreatesAsTypedRecords(t *testing.T) {
	previous := WorldFullPacket{Type: PacketTypeWorldFull}
	current := WorldFullPacket{Type: PacketTypeWorldFull, Pickups: []WorldPickupRecord{{ID: "pickup-a", Type: "shield", PickupClass: "powerup", X: 1, Y: 2, Health: 3, AgeSeconds: 4, LifespanSeconds: 5}}}

	delta := BuildWorldDeltaPacket(previous, current)

	if len(delta.Pickups.Creates) != 1 || delta.Pickups.Creates[0].ID != "pickup-a" || delta.Pickups.Creates[0].Type != "shield" {
		t.Fatalf("expected typed pickup create, got %#v", delta.Pickups.Creates)
	}
	if len(delta.Pickups.Updates) != 0 || len(delta.Pickups.Deletes) != 0 {
		t.Fatalf("expected only pickup create, got %#v", delta.Pickups)
	}
}

func TestBuildWorldDeltaPacketEmitsPickupUpdatesAsFieldDeltas(t *testing.T) {
	previous := WorldFullPacket{Type: PacketTypeWorldFull, Pickups: []WorldPickupRecord{{ID: "pickup-a", Type: "shield", PickupClass: "powerup", X: 1, Y: 2, Health: 3, AgeSeconds: 4, LifespanSeconds: 5}}}
	current := WorldFullPacket{Type: PacketTypeWorldFull, Pickups: []WorldPickupRecord{{ID: "pickup-a", Type: "shield", PickupClass: "powerup", X: 8, Y: 9, Health: 3, AgeSeconds: 7, LifespanSeconds: 5}}}

	delta := BuildWorldDeltaPacket(previous, current)

	if len(delta.Pickups.Updates) != 1 {
		t.Fatalf("expected one pickup update, got %#v", delta.Pickups.Updates)
	}
	got := delta.Pickups.Updates[0]
	if got["id"] != "pickup-a" || len(got) != 4 || got["x"] != float64(8) || got["y"] != float64(9) || got["age_seconds"] != float64(7) {
		t.Fatalf("expected id, x, y, and age_seconds only, got %#v", got)
	}
	if _, ok := got["type"]; ok {
		t.Fatalf("expected type to be omitted, got %#v", got)
	}
	if _, ok := got["pickup_class"]; ok {
		t.Fatalf("expected pickup_class to be omitted, got %#v", got)
	}
	if _, ok := got["health"]; ok {
		t.Fatalf("expected health to be omitted, got %#v", got)
	}
	if _, ok := got["lifespan_seconds"]; ok {
		t.Fatalf("expected lifespan_seconds to be omitted, got %#v", got)
	}
}

func TestBuildWorldDeltaPacketEmitsPickupDeletesAsIDs(t *testing.T) {
	previous := WorldFullPacket{Type: PacketTypeWorldFull, Pickups: []WorldPickupRecord{{ID: "pickup-a", Type: "shield", PickupClass: "powerup", X: 1, Y: 2, Health: 3, AgeSeconds: 4, LifespanSeconds: 5}}}
	current := WorldFullPacket{Type: PacketTypeWorldFull}

	delta := BuildWorldDeltaPacket(previous, current)

	if len(delta.Pickups.Deletes) != 1 || delta.Pickups.Deletes[0] != "pickup-a" {
		t.Fatalf("expected pickup delete by ID, got %#v", delta.Pickups.Deletes)
	}
	if len(delta.Pickups.Creates) != 0 || len(delta.Pickups.Updates) != 0 {
		t.Fatalf("expected only pickup delete, got %#v", delta.Pickups)
	}
}

func TestBuildWorldDeltaPacketEmitsShipCreatesAsTypedRecords(t *testing.T) {
	previous := WorldFullPacket{Type: PacketTypeWorldFull}
	current := WorldFullPacket{Type: PacketTypeWorldFull, Ships: []WorldShipRecord{{ID: "ship-a", ShipType: "v_wing", X: 1, Y: 2, Rotation: 3, Health: 4, Shields: 5, Thrusting: true, TargetKind: "player", TargetID: "player-1"}}}

	delta := BuildWorldDeltaPacket(previous, current)

	if len(delta.Ships.Creates) != 1 || delta.Ships.Creates[0].ID != "ship-a" || delta.Ships.Creates[0].ShipType != "v_wing" {
		t.Fatalf("expected typed ship create, got %#v", delta.Ships.Creates)
	}
	if len(delta.Ships.Updates) != 0 || len(delta.Ships.Deletes) != 0 {
		t.Fatalf("expected only ship create, got %#v", delta.Ships)
	}
}

func TestBuildWorldDeltaPacketEmitsShipUpdatesAsFieldDeltas(t *testing.T) {
	previous := WorldFullPacket{Type: PacketTypeWorldFull, Ships: []WorldShipRecord{{ID: "ship-a", ShipType: "v_wing", X: 1, Y: 2, Rotation: 3, Health: 4, Shields: 5, Thrusting: false, TargetKind: "player", TargetID: "player-1"}}}
	current := WorldFullPacket{Type: PacketTypeWorldFull, Ships: []WorldShipRecord{{ID: "ship-a", ShipType: "v_wing", X: 8, Y: 9, Rotation: 10, Health: 4, Shields: 5, Thrusting: true, TargetKind: "player", TargetID: "player-1"}}}

	delta := BuildWorldDeltaPacket(previous, current)

	if len(delta.Ships.Updates) != 1 {
		t.Fatalf("expected one ship update, got %#v", delta.Ships.Updates)
	}
	got := delta.Ships.Updates[0]
	if got["id"] != "ship-a" || len(got) != 5 || got["x"] != float64(8) || got["y"] != float64(9) || got["rotation"] != float64(10) || got["thrusting"] != true {
		t.Fatalf("expected id, x, y, rotation, and thrusting only, got %#v", got)
	}
	if _, ok := got["ship_type"]; ok {
		t.Fatalf("expected ship_type to be omitted, got %#v", got)
	}
	if _, ok := got["health"]; ok {
		t.Fatalf("expected health to be omitted, got %#v", got)
	}
	if _, ok := got["shields"]; ok {
		t.Fatalf("expected shields to be omitted, got %#v", got)
	}
	if _, ok := got["target_kind"]; ok {
		t.Fatalf("expected target_kind to be omitted, got %#v", got)
	}
	if _, ok := got["target_id"]; ok {
		t.Fatalf("expected target_id to be omitted, got %#v", got)
	}
}

func TestBuildWorldDeltaPacketEmitsShipDeletesAsIDs(t *testing.T) {
	previous := WorldFullPacket{Type: PacketTypeWorldFull, Ships: []WorldShipRecord{{ID: "ship-a", ShipType: "v_wing", X: 1}}}
	current := WorldFullPacket{Type: PacketTypeWorldFull}

	delta := BuildWorldDeltaPacket(previous, current)

	if len(delta.Ships.Deletes) != 1 || delta.Ships.Deletes[0] != "ship-a" {
		t.Fatalf("expected ship delete by ID, got %#v", delta.Ships.Deletes)
	}
	if len(delta.Ships.Creates) != 0 || len(delta.Ships.Updates) != 0 {
		t.Fatalf("expected only ship delete, got %#v", delta.Ships)
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
	previous := OverlayFullPacket{Type: PacketFamilyOverlayFull, Receiver: OverlayReceiverRecord{SelfID: "player-1", Lives: 2, Score: 5, RespawnCooldown: 1.25, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "limited", PrimaryCooldownRemaining: 7, PrimaryAmmoRemaining: 3, SecondaryWeaponID: "bomb", SecondaryAmmoPolicy: "infinite", SecondaryCooldownRemaining: 11, SecondaryAmmoRemaining: 4}}
	current := OverlayFullPacket{Type: PacketFamilyOverlayFull, Receiver: OverlayReceiverRecord{SelfID: "player-1", Lives: 3, Score: 9, RespawnCooldown: 0.5, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "limited", PrimaryCooldownRemaining: 7, PrimaryAmmoRemaining: 3, SecondaryWeaponID: "bomb", SecondaryAmmoPolicy: "infinite", SecondaryCooldownRemaining: 11, SecondaryAmmoRemaining: 4}}

	delta := BuildOverlayDeltaPacket(previous, current)

	if !OverlayDeltaHasChanges(delta) {
		t.Fatal("expected overlay delta to report changes")
	}
	if len(delta.Receiver.Updates) != 1 {
		t.Fatalf("expected one overlay update, got %#v", delta.Receiver.Updates)
	}
	if got := delta.Receiver.Updates[0]; len(got) != 4 || got["self_id"] != "player-1" || got["lives"] != 3 || got["score"] != 9 || got["respawn_cooldown"] != 0.5 {
		t.Fatalf("expected changed overlay receiver patch, got %#v", got)
	}
	if _, ok := delta.Receiver.Updates[0]["primary_weapon_id"]; ok {
		t.Fatalf("expected primary_weapon_id to be omitted, got %#v", delta.Receiver.Updates[0])
	}
	if _, ok := delta.Receiver.Updates[0]["primary_ammo_policy"]; ok {
		t.Fatalf("expected primary_ammo_policy to be omitted, got %#v", delta.Receiver.Updates[0])
	}
	if _, ok := delta.Receiver.Updates[0]["primary_cooldown_remaining"]; ok {
		t.Fatalf("expected primary_cooldown_remaining to be omitted, got %#v", delta.Receiver.Updates[0])
	}
	if _, ok := delta.Receiver.Updates[0]["primary_ammo_remaining"]; ok {
		t.Fatalf("expected primary_ammo_remaining to be omitted, got %#v", delta.Receiver.Updates[0])
	}
	if _, ok := delta.Receiver.Updates[0]["secondary_weapon_id"]; ok {
		t.Fatalf("expected secondary_weapon_id to be omitted, got %#v", delta.Receiver.Updates[0])
	}
	if _, ok := delta.Receiver.Updates[0]["secondary_ammo_policy"]; ok {
		t.Fatalf("expected secondary_ammo_policy to be omitted, got %#v", delta.Receiver.Updates[0])
	}
	if _, ok := delta.Receiver.Updates[0]["secondary_cooldown_remaining"]; ok {
		t.Fatalf("expected secondary_cooldown_remaining to be omitted, got %#v", delta.Receiver.Updates[0])
	}
	if _, ok := delta.Receiver.Updates[0]["secondary_ammo_remaining"]; ok {
		t.Fatalf("expected secondary_ammo_remaining to be omitted, got %#v", delta.Receiver.Updates[0])
	}
}

func TestBuildOverlayDeltaPacketEmitsScorePatchOnly(t *testing.T) {
	previous := OverlayFullPacket{Type: PacketFamilyOverlayFull, Receiver: OverlayReceiverRecord{SelfID: "player-1", Score: 5, Lives: 2, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "limited", PrimaryCooldownRemaining: 4, PrimaryAmmoRemaining: 8, SecondaryWeaponID: "bomb", SecondaryAmmoPolicy: "infinite", SecondaryCooldownRemaining: 6, SecondaryAmmoRemaining: 9}}
	current := OverlayFullPacket{Type: PacketFamilyOverlayFull, Receiver: OverlayReceiverRecord{SelfID: "player-1", Score: 12, Lives: 2, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "limited", PrimaryCooldownRemaining: 4, PrimaryAmmoRemaining: 8, SecondaryWeaponID: "bomb", SecondaryAmmoPolicy: "infinite", SecondaryCooldownRemaining: 6, SecondaryAmmoRemaining: 9}}

	delta := BuildOverlayDeltaPacket(previous, current)

	if len(delta.Receiver.Updates) != 1 {
		t.Fatalf("expected one overlay update, got %#v", delta.Receiver.Updates)
	}
	if got := delta.Receiver.Updates[0]; len(got) != 2 || got["self_id"] != "player-1" || got["score"] != 12 {
		t.Fatalf("expected self_id and score only, got %#v", got)
	}
}

func TestBuildOverlayDeltaPacketEmitsLivesPatchOnly(t *testing.T) {
	previous := OverlayFullPacket{Type: PacketFamilyOverlayFull, Receiver: OverlayReceiverRecord{SelfID: "player-1", Score: 5, Lives: 2, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "limited", PrimaryCooldownRemaining: 4, PrimaryAmmoRemaining: 8, SecondaryWeaponID: "bomb", SecondaryAmmoPolicy: "infinite", SecondaryCooldownRemaining: 6, SecondaryAmmoRemaining: 9}}
	current := OverlayFullPacket{Type: PacketFamilyOverlayFull, Receiver: OverlayReceiverRecord{SelfID: "player-1", Score: 5, Lives: 4, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "limited", PrimaryCooldownRemaining: 4, PrimaryAmmoRemaining: 8, SecondaryWeaponID: "bomb", SecondaryAmmoPolicy: "infinite", SecondaryCooldownRemaining: 6, SecondaryAmmoRemaining: 9}}

	delta := BuildOverlayDeltaPacket(previous, current)

	if len(delta.Receiver.Updates) != 1 {
		t.Fatalf("expected one overlay update, got %#v", delta.Receiver.Updates)
	}
	if got := delta.Receiver.Updates[0]; len(got) != 2 || got["self_id"] != "player-1" || got["lives"] != 4 {
		t.Fatalf("expected self_id and lives only, got %#v", got)
	}
}

func TestBuildOverlayDeltaPacketEmitsPrimaryCooldownPatchOnly(t *testing.T) {
	previous := OverlayFullPacket{Type: PacketFamilyOverlayFull, Receiver: OverlayReceiverRecord{SelfID: "player-1", Score: 5, Lives: 2, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "limited", PrimaryCooldownRemaining: 1.5, PrimaryAmmoRemaining: 8, SecondaryWeaponID: "bomb", SecondaryAmmoPolicy: "infinite", SecondaryCooldownRemaining: 6, SecondaryAmmoRemaining: 9}}
	current := OverlayFullPacket{Type: PacketFamilyOverlayFull, Receiver: OverlayReceiverRecord{SelfID: "player-1", Score: 5, Lives: 2, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "limited", PrimaryCooldownRemaining: 0, PrimaryAmmoRemaining: 8, SecondaryWeaponID: "bomb", SecondaryAmmoPolicy: "infinite", SecondaryCooldownRemaining: 6, SecondaryAmmoRemaining: 9}}

	delta := BuildOverlayDeltaPacket(previous, current)

	if len(delta.Receiver.Updates) != 1 {
		t.Fatalf("expected one overlay update, got %#v", delta.Receiver.Updates)
	}
	if got := delta.Receiver.Updates[0]; len(got) != 2 || got["self_id"] != "player-1" || got["primary_cooldown_remaining"] != float64(0) {
		t.Fatalf("expected self_id and primary_cooldown_remaining only, got %#v", got)
	}
}

func TestBuildOverlayDeltaPacketReturnsNoChangesForIdenticalOverlayData(t *testing.T) {
	current := OverlayFullPacket{Type: PacketFamilyOverlayFull, Receiver: OverlayReceiverRecord{SelfID: "player-1", Lives: 2, Score: 5, RespawnCooldown: 1.25, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "limited", PrimaryCooldownRemaining: 7, PrimaryAmmoRemaining: 3, SecondaryWeaponID: "bomb", SecondaryAmmoPolicy: "infinite", SecondaryCooldownRemaining: 11, SecondaryAmmoRemaining: 4}}

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
	if len(delta.Players.Updates) != 1 || delta.Players.Updates[0]["id"] != "player-a" || delta.Players.Updates[0]["score"] != 9 || delta.Players.Updates[0]["lives"] != 2 {
		t.Fatalf("expected player update, got %#v", delta.Players)
	}
	if len(delta.PlayerLifecycle.Updates) != 1 || delta.PlayerLifecycle.Updates[0]["player_id"] != "player-a" || delta.PlayerLifecycle.Updates[0]["status"] != "respawning" {
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



func TestBuildWorldWireDeltaPacketEmitsWireChanges(t *testing.T) {
	previous := WorldWireFullPacket{Type: PacketTypeWorldFull, Metadata: Metadata{Lane: LaneWorld, SnapshotID: "world-1", SnapshotKind: SnapshotKind("full")}, Ships: []WorldShipWireRecord{{ID: "ship-a", ShipType: "v_wing", X: 10, Y: 20, Rotation: 30, Health: 4, Shields: 5, Thrusting: false, TargetKind: "player", TargetID: "player-1"}}}
	current := WorldWireFullPacket{Type: PacketTypeWorldFull, Metadata: Metadata{Lane: LaneWorld, SnapshotID: "world-1", SnapshotKind: SnapshotKind("full")}, Ships: []WorldShipWireRecord{{ID: "ship-a", ShipType: "v_wing", X: 11, Y: 20, Rotation: 30, Health: 4, Shields: 5, Thrusting: false, TargetKind: "player", TargetID: "player-1"}}}

	delta := BuildWorldWireDeltaPacket(previous, current)

	if !WorldWireDeltaHasChanges(delta) {
		t.Fatal("expected world wire delta to report changes")
	}
	if got, want := delta.Metadata.SnapshotKind, SnapshotKind("delta"); got != want {
		t.Fatalf("world wire delta snapshot kind = %q, want %q", got, want)
	}
	if len(delta.Ships.Updates) != 1 {
		t.Fatalf("expected one ship update, got %#v", delta.Ships.Updates)
	}
	if got := delta.Ships.Updates[0]; got["id"] != "ship-a" || len(got) != 2 || got["x"] != int64(11) {
		t.Fatalf("expected id and quantized x only, got %#v", got)
	}
}
