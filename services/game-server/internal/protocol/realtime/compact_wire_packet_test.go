package realtime

import (
	"reflect"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/protocol/realtimewire"
)

func TestCompactWirePacketCompactsMetadataKeys(t *testing.T) {
	input := map[string]any{
		"type":             "world_full",
		"lane":             "world",
		"sequence":         int64(12),
		"snapshot_kind":    "full",
		"server_sent_msec": int64(34),
		"x":                1,
		"y":                2,
	}

	got := CompactWirePacket(input)

	if got["t"] != "wf" {
		t.Fatalf("type = %v, want wf", got["t"])
	}
	if got["l"] != "w" {
		t.Fatalf("lane = %v, want w", got["l"])
	}
	if got["q"] != int64(12) {
		t.Fatalf("sequence = %v, want 12", got["q"])
	}
	if got["k"] != "f" {
		t.Fatalf("snapshot_kind = %v, want f", got["k"])
	}
	if got["ms"] != int64(34) {
		t.Fatalf("server_sent_msec = %v, want 34", got["ms"])
	}
	if got["x"] != 1 || got["y"] != 2 {
		t.Fatalf("coordinate fields changed: %#v", got)
	}
	if _, ok := got["baseline_sequence"]; ok {
		t.Fatalf("did not expect readable baseline_sequence key in compact output: %#v", got)
	}
}

func TestCompactWirePacketCompactsLifecycleLaneValues(t *testing.T) {
	input := map[string]any{
		"type": "asteroids_lifecycle",
		"lane": "asteroids.lifecycle",
	}

	got := CompactWirePacket(input)

	if got["t"] != "al" {
		t.Fatalf("type = %v, want al", got["t"])
	}
	if got["l"] != "al" {
		t.Fatalf("asteroid lifecycle lane = %v, want al", got["l"])
	}

	bullet := CompactWirePacket(map[string]any{
		"type": "bullets_lifecycle",
		"lane": "bullets.lifecycle",
	})
	if bullet["t"] != "bl" {
		t.Fatalf("bullet type = %v, want bl", bullet["t"])
	}
	if bullet["l"] != "bl" {
		t.Fatalf("bullet lifecycle lane = %v, want bl", bullet["l"])
	}
}

func TestCompactWirePacketCompactsNestedMapRecordsRecursively(t *testing.T) {
	input := map[string]any{
		"type": "world_delta",
		"lane": "world",
		"pickup_updates": []any{
			map[string]any{
				"id":   "pickup-1",
				"x":    10,
				"y":    11,
				"nested": map[string]any{
					"owner_id": "player-9",
					"size":     3,
				},
			},
		},
	}

	got := CompactWirePacket(input)

	updates := got["pu"].([]any)
	first := updates[0].(map[string]any)
	if first["i"] != "pickup-1" {
		t.Fatalf("id = %v, want pickup-1", first["i"])
	}
	if first["x"] != 10 || first["y"] != 11 {
		t.Fatalf("position changed: %#v", first)
	}
	nested := first["nested"].(map[string]any)
	if nested["oi"] != "player-9" {
		t.Fatalf("nested owner_id = %v, want player-9", nested["oi"])
	}
	if nested["sz"] != 3 {
		t.Fatalf("nested size = %v, want 3", nested["sz"])
	}
}

func TestCompactWirePacketLeavesUnmappedValuesAloneOutsideContexts(t *testing.T) {
	input := map[string]any{
		"status": "world_delta",
		"lane":   "world",
		"type":   "session_full",
		"notes": []any{
			"overlay_delta",
			map[string]any{"snapshot_kind": "delta"},
		},
	}

	got := CompactWirePacket(input)

	if got["stat"] != "world_delta" {
		t.Fatalf("status = %v, want world_delta", got["stat"])
	}
	if got["l"] != "w" {
		t.Fatalf("lane = %v, want w", got["l"])
	}
	if got["t"] != "sf" {
		t.Fatalf("type = %v, want sf", got["t"])
	}
	notes := got["notes"].([]any)
	if notes[0] != "overlay_delta" {
		t.Fatalf("notes[0] = %v, want overlay_delta", notes[0])
	}
	if notes[1].(map[string]any)["k"] != "d" {
		t.Fatalf("notes[1].snapshot_kind = %v, want d", notes[1].(map[string]any)["k"])
	}
}

func TestCompactWirePacketDoesNotMutateInput(t *testing.T) {
	original := map[string]any{
		"type": "world_delta",
		"lane": "world",
		"ship_updates": []any{
			map[string]any{"id": "ship-1", "x": 1, "y": 2},
		},
	}

	got := CompactWirePacket(original)

	if original["type"] != "world_delta" {
		t.Fatalf("original type mutated: %v", original["type"])
	}
	if original["lane"] != "world" {
		t.Fatalf("original lane mutated: %v", original["lane"])
	}
	if original["ship_updates"].([]any)[0].(map[string]any)["id"] != "ship-1" {
		t.Fatalf("original nested record mutated")
	}
	if got["t"] != "wd" {
		t.Fatalf("compacted packet not returned as expected: %#v", got)
	}
}

func TestCompactWirePacketTuplePacksWorldFullAsteroids(t *testing.T) {
	input := map[string]any{
		"type": "world_full",
		"asteroids": []any{
			map[string]any{
				"id":      "asteroid-1",
				"x":       10,
				"y":       20,
				"size":    2,
				"health":  90,
				"scale":   1500,
				"variant": 3,
			},
		},
	}

	got := CompactWirePacket(input)

	if got["t"] != "wf" {
		t.Fatalf("type = %v, want wf", got["t"])
	}

	asteroids, ok := got["asteroids"].([]any)
	if !ok || len(asteroids) != 1 {
		t.Fatalf("asteroids = %#v, want one tuple-packed asteroid", got["asteroids"])
	}

	tuple, ok := asteroids[0].([]any)
	if !ok {
		t.Fatalf("asteroids[0] = %#v, want tuple array", asteroids[0])
	}

	want := []any{1, 10, 20, 2, 90, 1500, 3}
	if len(tuple) != len(want) {
		t.Fatalf("tuple len = %d, want %d (%#v)", len(tuple), len(want), tuple)
	}
	for i := range want {
		if tuple[i] != want[i] {
			t.Fatalf("tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
		}
	}
	for i, item := range tuple {
		if nested, ok := item.(map[string]any); ok {
			for _, key := range []string{"i", "sz", "h", "sl", "v"} {
				if _, exists := nested[key]; exists {
					t.Fatalf("tuple[%d] unexpectedly exposed compact asteroid key %q: %#v", i, key, tuple)
				}
			}
		}
	}

	if _, ok := asteroids[0].(map[string]any); ok {
		t.Fatalf("asteroids[0] unexpectedly remained a map: %#v", asteroids[0])
	}
}

func TestCompactWirePacketTuplePacksWorldFullBullets(t *testing.T) {
	input := map[string]any{
		"type": "world_full",
		"bullets": []any{
			map[string]any{
				"id":             "bullet-1",
				"owner_id":       "player-1",
				"x":              10,
				"y":              20,
				"rotation":       30,
				"weapon_id":      "pulse",
				"projectile_type": "laser",
			},
		},
	}

	got := CompactWirePacket(input)

	if got["t"] != "wf" {
		t.Fatalf("type = %v, want wf", got["t"])
	}

	bullets, ok := got["bullets"].([]any)
	if !ok || len(bullets) != 1 {
		t.Fatalf("bullets = %#v, want one tuple-packed bullet", got["bullets"])
	}

	tuple, ok := bullets[0].([]any)
	if !ok {
		t.Fatalf("bullets[0] = %#v, want tuple array", bullets[0])
	}

	want := []any{1, 1, 10, 20, 30, "pulse", "laser"}
	if len(tuple) != len(want) {
		t.Fatalf("tuple len = %d, want %d (%#v)", len(tuple), len(want), tuple)
	}
	for i := range want {
		if tuple[i] != want[i] {
			t.Fatalf("tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
		}
	}
}

func TestCompactWirePacketTuplePacksWorldFullShips(t *testing.T) {
	input := map[string]any{
		"type": "world_full",
		"ships": []any{
			map[string]any{
				"id":          "player-1",
				"ship_type":   "v_wing",
				"x":           10,
				"y":           20,
				"rotation":    30,
				"health":      100,
				"shields":     50,
				"thrusting":   true,
				"target_kind": "player",
				"target_id":   "player-2",
			},
		},
	}

	got := CompactWirePacket(input)
	ships, ok := got["ships"].([]any)
	if !ok || len(ships) != 1 {
		t.Fatalf("ships = %#v, want one tuple-packed ship", got["ships"])
	}
	tuple, ok := ships[0].([]any)
	if !ok {
		t.Fatalf("ships[0] = %#v, want tuple array", ships[0])
	}
	want := []any{1, "v_wing", 10, 20, 30, 100, 50, true, "player", 2}
	if len(tuple) != len(want) {
		t.Fatalf("tuple len = %d, want %d (%#v)", len(tuple), len(want), tuple)
	}
	for i := range want {
		if tuple[i] != want[i] {
			t.Fatalf("tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
		}
	}
}


func TestCompactWirePacketTuplePacksWorldDeltaShipCreates(t *testing.T) {
	input := map[string]any{
		"type": "world_delta",
		"ship_creates": []any{
			map[string]any{
				"id":          "player-1",
				"ship_type":   "v_wing",
				"x":           10,
				"y":           20,
				"rotation":    30,
				"health":      100,
				"shields":     50,
				"thrusting":   true,
				"target_kind": "player",
				"target_id":   "player-2",
			},
		},
	}

	got := CompactWirePacket(input)
	creates := got["sc"].([]any)
	want := []any{1, "v_wing", 10, 20, 30, 100, 50, true, "player", 2}
	if tuple, ok := creates[0].([]any); !ok || len(tuple) != len(want) {
		t.Fatalf("ship create = %#v, want full tuple %#v", creates[0], want)
	} else {
		for i := range want {
			if tuple[i] != want[i] {
				t.Fatalf("create tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
			}
		}
	}
}

func TestCompactWirePacketTuplePacksWorldDeltaShipUpdatesXYRotationThrusting(t *testing.T) {
	input := map[string]any{
		"type": "world_delta",
		"ship_updates": []any{
			map[string]any{
				"id":        "player-1",
				"x":         10,
				"y":         20,
				"rotation":  30,
				"thrusting": true,
			},
		},
	}

	got := CompactWirePacket(input)
	updates := got["su"].([]any)
	want := []any{1, 10, 20, 30, true}
	if tuple, ok := updates[0].([]any); !ok || len(tuple) != len(want) {
		t.Fatalf("ship xy rotation thrusting update = %#v, want tuple %#v", updates[0], want)
	} else {
		for i := range want {
			if tuple[i] != want[i] {
				t.Fatalf("update tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
			}
		}
	}
}

func TestCompactWirePacketTuplePacksWorldDeltaShipUpdatesXOnly(t *testing.T) {
	input := map[string]any{
		"type": "world_delta",
		"ship_updates": []any{
			map[string]any{
				"id": "player-1",
				"x":  10,
			},
		},
	}

	got := CompactWirePacket(input)
	updates := got["su"].([]any)
	want := []any{1, 10}
	if tuple, ok := updates[0].([]any); !ok || len(tuple) != len(want) {
		t.Fatalf("ship x-only update = %#v, want tuple %#v", updates[0], want)
	} else {
		for i := range want {
			if tuple[i] != want[i] {
				t.Fatalf("x-only tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
			}
		}
	}
}

func TestCompactWirePacketTuplePacksWorldDeltaShipUpdatesYOnly(t *testing.T) {
	input := map[string]any{
		"type": "world_delta",
		"ship_updates": []any{
			map[string]any{
				"id": "player-1",
				"y":  20,
			},
		},
	}

	got := CompactWirePacket(input)
	updates := got["su"].([]any)
	want := []any{1, nil, 20}
	if tuple, ok := updates[0].([]any); !ok || len(tuple) != len(want) {
		t.Fatalf("ship y-only update = %#v, want tuple %#v", updates[0], want)
	} else {
		for i := range want {
			if tuple[i] != want[i] {
				t.Fatalf("y-only tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
			}
		}
	}
}

func TestCompactWirePacketTuplePacksWorldDeltaShipUpdatesRotationOnly(t *testing.T) {
	input := map[string]any{
		"type": "world_delta",
		"ship_updates": []any{
			map[string]any{
				"id":       "player-1",
				"rotation": 30,
			},
		},
	}

	got := CompactWirePacket(input)
	updates := got["su"].([]any)
	want := []any{1, nil, nil, 30}
	if tuple, ok := updates[0].([]any); !ok || len(tuple) != len(want) {
		t.Fatalf("ship rotation-only update = %#v, want tuple %#v", updates[0], want)
	} else {
		for i := range want {
			if tuple[i] != want[i] {
				t.Fatalf("rotation-only tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
			}
		}
	}
}

func TestCompactWirePacketTuplePacksWorldDeltaShipUpdatesThrustingOnlyFalse(t *testing.T) {
	input := map[string]any{
		"type": "world_delta",
		"ship_updates": []any{
			map[string]any{
				"id":        "player-1",
				"thrusting": false,
			},
		},
	}

	got := CompactWirePacket(input)
	updates := got["su"].([]any)
	want := []any{1, nil, nil, nil, false}
	if tuple, ok := updates[0].([]any); !ok || len(tuple) != len(want) {
		t.Fatalf("ship thrusting-only update = %#v, want tuple %#v", updates[0], want)
	} else {
		for i := range want {
			if tuple[i] != want[i] {
				t.Fatalf("thrusting-only tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
			}
		}
	}
}

func TestCompactWirePacketTuplePacksWorldDeltaShipUpdatesPreserveZeroValues(t *testing.T) {
	input := map[string]any{
		"type": "world_delta",
		"ship_updates": []any{
			map[string]any{
				"id":        "player-1",
				"x":         0,
				"y":         0,
				"rotation":  0,
				"thrusting": false,
			},
		},
	}

	got := CompactWirePacket(input)
	updates := got["su"].([]any)
	want := []any{1, 0, 0, 0, false}
	if tuple, ok := updates[0].([]any); !ok || len(tuple) != len(want) {
		t.Fatalf("ship zero update = %#v, want tuple %#v", updates[0], want)
	} else {
		for i := range want {
			if tuple[i] != want[i] {
				t.Fatalf("zero tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
			}
		}
	}
}

func TestCompactWirePacketTuplePacksWorldDeltaShipUpdatesSparseFieldsPreserveLaterValues(t *testing.T) {
	input := map[string]any{
		"type": "world_delta",
		"ship_updates": []any{
			map[string]any{
				"id": "player-1",
				"y": 20,
			},
		},
	}

	got := CompactWirePacket(input)
	updates := got["su"].([]any)
	tuple := updates[0].([]any)
	want := []any{1, nil, 20}
	if len(tuple) != len(want) {
		t.Fatalf("ship y-only tuple len = %d, want %d (%#v)", len(tuple), len(want), tuple)
	}
	for i := range want {
		if tuple[i] != want[i] {
			t.Fatalf("ship y-only tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
		}
	}

	input = map[string]any{
		"type": "world_delta",
		"ship_updates": []any{
			map[string]any{
				"id": "player-1",
				"rotation": 30,
			},
		},
	}
	got = CompactWirePacket(input)
	tuple = got["su"].([]any)[0].([]any)
	want = []any{1, nil, nil, 30}
	if len(tuple) != len(want) {
		t.Fatalf("ship rotation-only tuple len = %d, want %d (%#v)", len(tuple), len(want), tuple)
	}
	for i := range want {
		if tuple[i] != want[i] {
			t.Fatalf("ship rotation-only tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
		}
	}

	input = map[string]any{
		"type": "world_delta",
		"ship_updates": []any{
			map[string]any{
				"id": "player-1",
				"thrusting": true,
			},
		},
	}
	got = CompactWirePacket(input)
	tuple = got["su"].([]any)[0].([]any)
	want = []any{1, nil, nil, nil, true}
	if len(tuple) != len(want) {
		t.Fatalf("ship thrusting-only tuple len = %d, want %d (%#v)", len(tuple), len(want), tuple)
	}
	for i := range want {
		if tuple[i] != want[i] {
			t.Fatalf("ship thrusting-only tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
		}
	}
}


func TestCompactWirePacketCompactsWorldDeltaShipDeletes(t *testing.T) {
	input := map[string]any{
		"type":           "world_delta",
		"ship_deletes": []any{"player-1", "ship-legacy", "player-2"},
	}

	got := CompactWirePacket(input)
	deletes := got["sx"].([]any)
	if len(deletes) != 3 || deletes[0] != 1 || deletes[1] != "ship-legacy" || deletes[2] != 2 {
		t.Fatalf("ship deletes changed: %#v", deletes)
	}
}

func TestCompactWirePacketTuplePacksSessionFullPlayers(t *testing.T) {
	input := map[string]any{
		"type": "session_full",
		"players": []any{
			map[string]any{
				"id":                  "player-1",
				"ship_type":           "v_wing",
				"score":               100,
				"lives":               3,
				"respawn_cooldown":    250,
				"primary_weapon_id":   "pulse",
				"primary_ammo_policy": "limited",
				"secondary_weapon_id": "mine",
				"secondary_ammo_policy": "limited",
				"spawn_x":             10,
				"spawn_y":             20,
			},
		},
	}

	got := CompactWirePacket(input)
	players := got["pl"].([]any)
	want := []any{1, "v_wing", 100, 3, 250, "pulse", "limited", "mine", "limited", 10, 20}
	if tuple, ok := players[0].([]any); !ok || len(tuple) != len(want) {
		t.Fatalf("session player = %#v, want full tuple %#v", players[0], want)
	} else {
		for i := range want {
			if tuple[i] != want[i] {
				t.Fatalf("player tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
			}
		}
	}
}


func TestCompactWirePacketTuplePacksSessionDeltaPlayerCreates(t *testing.T) {
	input := map[string]any{
		"type": "session_delta",
		"pl": []any{
			map[string]any{
				"id":                  "player-1",
				"ship_type":           "v_wing",
				"score":               100,
				"lives":               3,
				"respawn_cooldown":    250,
				"primary_weapon_id":   "pulse",
				"primary_ammo_policy": "limited",
				"secondary_weapon_id": "mine",
				"secondary_ammo_policy": "limited",
				"spawn_x":             10,
				"spawn_y":             20,
			},
		},
	}

	got := CompactWirePacket(input)
	players := got["pl"].([]any)
	want := []any{1, "v_wing", 100, 3, 250, "pulse", "limited", "mine", "limited", 10, 20}
	if tuple, ok := players[0].([]any); !ok || len(tuple) != len(want) {
		t.Fatalf("session delta player = %#v, want full tuple %#v", players[0], want)
	} else {
		for i := range want {
			if tuple[i] != want[i] {
				t.Fatalf("player create tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
			}
		}
	}
}

func TestCompactWirePacketTuplePacksSessionDeltaPlayerSessionUpdates(t *testing.T) {
	input := map[string]any{
		"type": "session_delta",
		"psu": []any{
			map[string]any{
				"id":               "player-1",
				"score":            100,
				"lives":            2,
				"respawn_cooldown": 0,
			},
		},
	}

	got := CompactWirePacket(input)
	updates := got["psu"].([]any)
	tuple := updates[0].([]any)
	if tuple[0] != 1 {
		t.Fatalf("update id = %#v, want 1", tuple[0])
	}
	if len(tuple) != 7 || tuple[1] != "sco" || tuple[2] != 100 || tuple[3] != "lv" || tuple[4] != 2 || tuple[5] != "rcd" || tuple[6] != 0 {
		t.Fatalf("player session update tuple = %#v", tuple)
	}
}

func TestCompactWirePacketCompactsSessionDeltaPlayerSessionDeletes(t *testing.T) {
	input := map[string]any{
		"type": "session_delta",
		"psx": []any{"player-1", "player-bad", "player-2"},
	}

	got := CompactWirePacket(input)
	deletes := got["psx"].([]any)
	if len(deletes) != 3 || deletes[0] != 1 || deletes[1] != "player-bad" || deletes[2] != 2 {
		t.Fatalf("player session deletes changed: %#v", deletes)
	}
}

func TestCompactWirePacketTuplePacksSessionLifecycleCreatesAndUpdates(t *testing.T) {
	create := CompactWirePacket(map[string]any{
		"type": "session_full",
		"plc": []any{
			map[string]any{"player_id": "player-1", "status": "active"},
		},
	})
	update := CompactWirePacket(map[string]any{
		"type": "session_delta",
		"plu": []any{
			map[string]any{"pid": "player-1", "stat": "respawning"},
		},
	})

	if got := create["plc"].([]any)[0]; !reflect.DeepEqual(got, []any{1, "active"}) {
		t.Fatalf("player lifecycle create = %#v, want %#v", got, []any{1, "active"})
	}
	if got := update["plu"].([]any)[0]; !reflect.DeepEqual(got, []any{1, "respawning"}) {
		t.Fatalf("player lifecycle update = %#v, want %#v", got, []any{1, "respawning"})
	}
}

func TestCompactWirePacketCompactsSessionLifecycleDeletes(t *testing.T) {
	input := map[string]any{
		"type": "session_delta",
		"plx": []any{"player-1"},
	}

	got := CompactWirePacket(input)
	deletes := got["plx"].([]any)
	if len(deletes) != 1 || deletes[0] != 1 {
		t.Fatalf("player lifecycle deletes changed: %#v", deletes)
	}
}

func TestCompactWirePacketTuplePacksWorldDeltaAsteroidCreates(t *testing.T) {
	input := map[string]any{
		"type": "world_delta",
		"asteroid_creates": []any{
			map[string]any{
				"id":      "asteroid-1",
				"x":       10,
				"y":       20,
				"size":    2,
				"health":  90,
				"scale":   1500,
				"variant": 3,
			},
		},
	}

	got := CompactWirePacket(input)
	creates := got["ac"].([]any)
	want := []any{1, 10, 20, 2, 90, 1500, 3}
	if tuple, ok := creates[0].([]any); !ok || len(tuple) != len(want) {
		t.Fatalf("asteroid create = %#v, want full tuple %#v", creates[0], want)
	} else {
		for i := range want {
			if tuple[i] != want[i] {
				t.Fatalf("create tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
			}
		}
	}
}

func TestCompactWirePacketTuplePacksWorldDeltaBulletUpdatesXYRotation(t *testing.T) {
	input := map[string]any{
		"type": "world_delta",
		"bullet_updates": []any{
			map[string]any{
				"id":       "bullet-1",
				"x":        10,
				"y":        20,
				"rotation": 30,
			},
		},
	}

	got := CompactWirePacket(input)
	updates := got["bu"].([]any)
	want := []any{1, 10, 20, 30}
	if tuple, ok := updates[0].([]any); !ok || len(tuple) != len(want) {
		t.Fatalf("bullet xy rotation update = %#v, want tuple %#v", updates[0], want)
	} else {
		for i := range want {
			if tuple[i] != want[i] {
				t.Fatalf("update tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
			}
		}
	}
}

func TestCompactWirePacketTuplePacksWorldDeltaBulletUpdatesXOnly(t *testing.T) {
	input := map[string]any{
		"type": "world_delta",
		"bullet_updates": []any{
			map[string]any{
				"id": "bullet-1",
				"x":  10,
			},
		},
	}

	got := CompactWirePacket(input)
	updates := got["bu"].([]any)
	want := []any{1, 10}
	if tuple, ok := updates[0].([]any); !ok || len(tuple) != len(want) {
		t.Fatalf("bullet x-only update = %#v, want tuple %#v", updates[0], want)
	} else {
		for i := range want {
			if tuple[i] != want[i] {
				t.Fatalf("x-only tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
			}
		}
	}
}

func TestCompactWirePacketTuplePacksWorldDeltaBulletUpdatesYOnly(t *testing.T) {
	input := map[string]any{
		"type": "world_delta",
		"bullet_updates": []any{
			map[string]any{
				"id": "bullet-1",
				"y":  20,
			},
		},
	}

	got := CompactWirePacket(input)
	updates := got["bu"].([]any)
	want := []any{1, nil, 20}
	if tuple, ok := updates[0].([]any); !ok || len(tuple) != len(want) {
		t.Fatalf("bullet y-only update = %#v, want tuple %#v", updates[0], want)
	} else {
		for i := range want {
			if tuple[i] != want[i] {
				t.Fatalf("y-only tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
			}
		}
	}
}

func TestCompactWirePacketTuplePacksWorldDeltaBulletUpdatesRotationOnly(t *testing.T) {
	input := map[string]any{
		"type": "world_delta",
		"bullet_updates": []any{
			map[string]any{
				"id":       "bullet-1",
				"rotation": 30,
			},
		},
	}

	got := CompactWirePacket(input)
	updates := got["bu"].([]any)
	want := []any{1, nil, nil, 30}
	if tuple, ok := updates[0].([]any); !ok || len(tuple) != len(want) {
		t.Fatalf("bullet rotation-only update = %#v, want tuple %#v", updates[0], want)
	} else {
		for i := range want {
			if tuple[i] != want[i] {
				t.Fatalf("rotation-only tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
			}
		}
	}
}

func TestCompactWirePacketTuplePacksWorldDeltaBulletUpdatesPreserveZeroValues(t *testing.T) {
	input := map[string]any{
		"type": "world_delta",
		"bullet_updates": []any{
			map[string]any{
				"id":       "bullet-1",
				"x":        0,
				"y":        0,
				"rotation": 0,
			},
		},
	}

	got := CompactWirePacket(input)
	updates := got["bu"].([]any)
	want := []any{1, 0, 0, 0}
	if tuple, ok := updates[0].([]any); !ok || len(tuple) != len(want) {
		t.Fatalf("bullet zero update = %#v, want tuple %#v", updates[0], want)
	} else {
		for i := range want {
			if tuple[i] != want[i] {
				t.Fatalf("zero tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
			}
		}
	}
}

func TestCompactWirePacketTuplePacksWorldDeltaBulletUpdatesSparseFieldsPreserveLaterValues(t *testing.T) {
	input := map[string]any{
		"type": "world_delta",
		"bullet_updates": []any{
			map[string]any{
				"id": "bullet-1",
				"y": 20,
			},
		},
	}

	got := CompactWirePacket(input)
	updates := got["bu"].([]any)
	tuple := updates[0].([]any)
	want := []any{1, nil, 20}
	if len(tuple) != len(want) {
		t.Fatalf("bullet y-only tuple len = %d, want %d (%#v)", len(tuple), len(want), tuple)
	}
	for i := range want {
		if tuple[i] != want[i] {
			t.Fatalf("bullet y-only tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
		}
	}

	input = map[string]any{
		"type": "world_delta",
		"bullet_updates": []any{
			map[string]any{
				"id": "bullet-1",
				"rotation": 30,
			},
		},
	}
	got = CompactWirePacket(input)
	tuple = got["bu"].([]any)[0].([]any)
	want = []any{1, nil, nil, 30}
	if len(tuple) != len(want) {
		t.Fatalf("bullet rotation-only tuple len = %d, want %d (%#v)", len(tuple), len(want), tuple)
	}
	for i := range want {
		if tuple[i] != want[i] {
			t.Fatalf("bullet rotation-only tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
		}
	}
}


func TestCompactWirePacketTuplePacksWorldDeltaAsteroidUpdatesXY(t *testing.T) {
	input := map[string]any{
		"type": "world_delta",
		"asteroid_updates": []any{
			map[string]any{
				"id": "asteroid-1",
				"x":  10,
				"y":  20,
			},
		},
	}

	got := CompactWirePacket(input)
	updates := got["au"].([]any)
	want := []any{1, 10, 20}
	if tuple, ok := updates[0].([]any); !ok || len(tuple) != len(want) {
		t.Fatalf("asteroid update = %#v, want tuple %#v", updates[0], want)
	} else {
		for i := range want {
			if tuple[i] != want[i] {
				t.Fatalf("update tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
			}
		}
	}
}

func TestCompactWirePacketTuplePacksWorldDeltaAsteroidUpdatesYOnly(t *testing.T) {
	input := map[string]any{
		"type": "world_delta",
		"asteroid_updates": []any{
			map[string]any{
				"id": "asteroid-1",
				"y":  20,
			},
		},
	}

	got := CompactWirePacket(input)
	updates := got["au"].([]any)
	want := []any{1, nil, 20}
	if tuple, ok := updates[0].([]any); !ok || len(tuple) != len(want) {
		t.Fatalf("asteroid y-only update = %#v, want tuple %#v", updates[0], want)
	} else {
		for i := range want {
			if tuple[i] != want[i] {
				t.Fatalf("y-only tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
			}
		}
	}
}

func TestCompactWirePacketTuplePacksWorldDeltaAsteroidUpdatesPreserveZeroValues(t *testing.T) {
	input := map[string]any{
		"type": "world_delta",
		"asteroid_updates": []any{
			map[string]any{
				"id": "asteroid-1",
				"x":  0,
				"y":  0,
			},
		},
	}

	got := CompactWirePacket(input)
	updates := got["au"].([]any)
	want := []any{1, 0, 0}
	if tuple, ok := updates[0].([]any); !ok || len(tuple) != len(want) {
		t.Fatalf("asteroid zero update = %#v, want tuple %#v", updates[0], want)
	} else {
		for i := range want {
			if tuple[i] != want[i] {
				t.Fatalf("zero tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
			}
		}
	}
}

func TestCompactWirePacketAsteroidDeltaCompactsTupleUpdates(t *testing.T) {
	input := map[string]any{
		"type": "asteroid_delta",
		"sequence": 33,
		"server_sent_msec": 12345,
		"asteroid_updates": []any{
			map[string]any{
				"id": "asteroid-123",
				"x":  10,
				"y":  20,
			},
		},
	}

	got := CompactWirePacket(input)
	assertStringValue(t, got, "t", "ad")
	assertIntValue(t, got, "q", 33)
	assertIntValue(t, got, "ms", 12345)
	updates := got["au"].([]any)
	if len(updates) != 1 {
		t.Fatalf("expected one asteroid update, got %#v", updates)
	}
	tuple := updates[0].([]any)
	want := []any{123, 10, 20}
	if len(tuple) != len(want) {
		t.Fatalf("asteroid tuple = %#v, want %#v", tuple, want)
	}
	for i := range want {
		if tuple[i] != want[i] {
			t.Fatalf("asteroid tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
		}
	}
	for _, key := range []string{"ac", "ax", "bc", "bx"} {
		assertNotContainsKey(t, got, key)
	}
}

func TestCompactWirePacketBulletDeltaCompactsTupleUpdates(t *testing.T) {
	input := map[string]any{
		"type": "bullet_delta",
		"sequence": 34,
		"server_sent_msec": 54321,
		"bullet_updates": []any{
			map[string]any{
				"id":       "bullet-123",
				"x":        11,
				"y":        22,
				"rotation": 33,
			},
		},
	}

	got := CompactWirePacket(input)
	assertStringValue(t, got, "t", "bd")
	assertIntValue(t, got, "q", 34)
	assertIntValue(t, got, "ms", 54321)
	updates := got["bu"].([]any)
	if len(updates) != 1 {
		t.Fatalf("expected one bullet update, got %#v", updates)
	}
	tuple := updates[0].([]any)
	want := []any{123, 11, 22, 33}
	if len(tuple) != len(want) {
		t.Fatalf("bullet tuple = %#v, want %#v", tuple, want)
	}
	for i := range want {
		if tuple[i] != want[i] {
			t.Fatalf("bullet tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
		}
	}
	for _, key := range []string{"ac", "ax", "bc", "bx"} {
		assertNotContainsKey(t, got, key)
	}
}

func TestCompactWirePacketLeavesMalformedAsteroidIDsUnchanged(t *testing.T) {
	input := map[string]any{
		"type": "world_delta",
		"asteroid_updates": []any{
			map[string]any{
				"id": "asteroid-not-a-number",
				"x":  10,
				"y":  20,
			},
		},
	}

	got := CompactWirePacket(input)
	updates := got["au"].([]any)
	want := []any{"asteroid-not-a-number", 10, 20}
	if tuple, ok := updates[0].([]any); !ok || len(tuple) != len(want) {
		t.Fatalf("malformed asteroid update = %#v, want tuple %#v", updates[0], want)
	} else {
		for i := range want {
			if tuple[i] != want[i] {
				t.Fatalf("malformed tuple[%d] = %#v, want %#v (full tuple %#v)", i, tuple[i], want[i], tuple)
			}
		}
	}
}

func TestCompactWirePacketCompactsWorldDeltaBulletDeletes(t *testing.T) {
	input := map[string]any{
		"type":           "world_delta",
		"bullet_deletes": []any{"bullet-1", "bullet-not-a-number", "bullet-2"},
	}

	got := CompactWirePacket(input)
	deletes := got["bx"].([]any)
	if len(deletes) != 3 || deletes[0] != 1 || deletes[1] != "bullet-not-a-number" || deletes[2] != 2 {
		t.Fatalf("bullet deletes changed: %#v", deletes)
	}
}

func TestCompactWirePacketCompactsWorldDeltaAsteroidDeletes(t *testing.T) {
	input := map[string]any{
		"type":             "world_delta",
		"asteroid_deletes": []any{"asteroid-1", "asteroid-not-a-number", "asteroid-2"},
	}

	got := CompactWirePacket(input)
	deletes := got["ax"].([]any)
	if len(deletes) != 3 || deletes[0] != 1 || deletes[1] != "asteroid-not-a-number" || deletes[2] != 2 {
		t.Fatalf("asteroid deletes changed: %#v", deletes)
	}
}

func TestCompactWirePacketCompactsReadableWorldDeltaMap(t *testing.T) {
	input := map[string]any{
		"type":              "world_delta",
		"sequence":          int64(7),
		"baseline_sequence": int64(5),
		"server_sent_msec":  int64(123),
		"ship_updates": []any{
			map[string]any{
				"id":        "ship-1",
				"x":         10,
				"y":         20,
				"rotation":  3142,
				"thrusting": false,
			},
		},
	}

	got := CompactWirePacket(input)

	for _, key := range []string{"t", "q", "bq", "ms", "su"} {
		if _, ok := got[key]; !ok {
			t.Fatalf("expected compact key %q to be present, got %#v", key, got)
		}
	}
	for _, key := range []string{"l", "k", "sid", "ci", "cc", "fc", "b"} {
		if _, ok := got[key]; ok {
			t.Fatalf("did not expect legacy compact key %q in compact runtime output: %#v", key, got)
		}
	}
	for _, key := range []string{"type", "sequence", "baseline_sequence", "server_sent_msec", "ship_updates"} {
		if _, ok := got[key]; ok {
			t.Fatalf("did not expect readable key %q in compact output: %#v", key, got)
		}
	}
}

func TestCompactWirePacketCompactsEventBatchAndNestedEventRecords(t *testing.T) {
	input := map[string]any{
		"type": "event_batch",
		"batch_id": "event-batch-11",
		"events": []any{
			map[string]any{"event_id": "presentation-event-1", "type": "bullet_blast", "x": 10, "y": 20},
			map[string]any{"event_id": "presentation-event-2", "type": "ship_death", "player_id": "player-1", "lives": 2, "respawn_delay": 1.25, "x": 30, "y": 40},
			map[string]any{"event_id": "presentation-event-3", "type": "damage_applied", "source_type": "projectile", "source_id": "bullet-1", "effect_type": "blast", "amount": 17, "x": 50, "y": 60},
			map[string]any{"event_id": "presentation-event-4", "type": "damage_over_time_started", "source_type": "asteroid", "source_id": "hazard-1", "effect_type": "radioactive", "amount": 2},
			map[string]any{"event_id": "presentation-event-5", "type": "damage_over_time_tick", "source_type": "asteroid", "source_id": "hazard-1", "effect_type": "radioactive", "amount": 3, "x": 70, "y": 80},
			map[string]any{"event_id": "presentation-event-6", "type": "radial_effect_started", "source_type": "pickup", "source_id": "pickup-1", "effect_type": "pulse", "x": 90, "y": 100},
			map[string]any{"event_id": "presentation-event-7", "type": "pickup_collected", "player_id": "player-1", "pickup_id": "pickup-1", "pickup_type": "shield", "x": 110, "y": 120},
			map[string]any{"event_id": "presentation-event-8", "type": "pickup_effect_applied", "player_id": "player-1", "pickup_id": "pickup-1", "pickup_type": "shield", "effect_type": "repair", "amount": 4, "lives_after": 3},
			map[string]any{"event_id": "presentation-event-9", "type": "pickup_expired", "pickup_id": "pickup-1", "pickup_type": "shield", "x": 130, "y": 140},
			map[string]any{"event_id": "presentation-event-10", "type": "pickup_dropped", "pickup_id": "pickup-1", "pickup_type": "shield", "source_type": "ship", "source_id": "ship-1", "table_id": "table-1", "x": 150, "y": 160},
		},
	}

	got := CompactWirePacket(input)

	if got["t"] != "eb" {
		t.Fatalf("type = %v, want eb", got["t"])
	}
	if got["bid"] != 11 {
		t.Fatalf("batch_id = %v, want 11", got["bid"])
	}
	events := got["ev"].([]any)
	if len(events) != 10 {
		t.Fatalf("events = %#v, want 10 items", events)
	}
	if _, ok := events[0].([]any); !ok {
		t.Fatalf("bullet_blast event = %#v, want tuple array", events[0])
	}
	if _, ok := events[1].([]any); !ok {
		t.Fatalf("ship_death event = %#v, want tuple array", events[1])
	}
	if _, ok := events[2].([]any); !ok {
		t.Fatalf("damage_applied event = %#v, want tuple array", events[2])
	}
	if _, ok := events[3].([]any); !ok {
		t.Fatalf("damage_over_time_started event = %#v, want tuple array", events[3])
	}
	if _, ok := events[4].([]any); !ok {
		t.Fatalf("damage_over_time_tick event = %#v, want tuple array", events[4])
	}
	if _, ok := events[5].([]any); !ok {
		t.Fatalf("radial_effect_started event = %#v, want tuple array", events[5])
	}
	if _, ok := events[6].([]any); !ok {
		t.Fatalf("pickup_collected event = %#v, want tuple array", events[6])
	}
	if _, ok := events[7].([]any); !ok {
		t.Fatalf("pickup_effect_applied event = %#v, want tuple array", events[7])
	}
	if _, ok := events[8].([]any); !ok {
		t.Fatalf("pickup_expired event = %#v, want tuple array", events[8])
	}
	if _, ok := events[9].([]any); !ok {
		t.Fatalf("pickup_dropped event = %#v, want tuple array", events[9])
	}
	if events[0].([]any)[0] != "bb" || events[0].([]any)[1] != 1 {
		t.Fatalf("bullet_blast tuple = %#v", events[0])
	}
	if events[1].([]any)[0] != "shd" || events[1].([]any)[1] != 2 || events[1].([]any)[2] != 1 {
		t.Fatalf("ship_death tuple = %#v", events[1])
	}
	if events[2].([]any)[0] != "dmg" || events[2].([]any)[1] != 3 || events[2].([]any)[3] != 1 {
		t.Fatalf("damage_applied tuple = %#v", events[2])
	}
	if events[3].([]any)[0] != "dots" || events[3].([]any)[1] != 4 || events[3].([]any)[3] != "hazard-1" {
		t.Fatalf("damage_over_time_started tuple = %#v", events[3])
	}
	if events[4].([]any)[0] != "dott" || events[4].([]any)[1] != 5 || events[4].([]any)[3] != "hazard-1" {
		t.Fatalf("damage_over_time_tick tuple = %#v", events[4])
	}
	if events[5].([]any)[0] != "rfx" || events[5].([]any)[1] != 6 || events[5].([]any)[3] != 1 {
		t.Fatalf("radial_effect_started tuple = %#v", events[5])
	}
	if events[6].([]any)[0] != "pcol" || events[6].([]any)[1] != 7 || events[6].([]any)[2] != 1 || events[6].([]any)[3] != 1 {
		t.Fatalf("pickup_collected tuple = %#v", events[6])
	}
	if events[7].([]any)[0] != "pea" || events[7].([]any)[1] != 8 || events[7].([]any)[2] != 1 || events[7].([]any)[3] != 1 {
		t.Fatalf("pickup_effect_applied tuple = %#v", events[7])
	}
	if events[8].([]any)[0] != "pexp" || events[8].([]any)[1] != 9 || events[8].([]any)[2] != 1 {
		t.Fatalf("pickup_expired tuple = %#v", events[8])
	}
	if events[9].([]any)[0] != "pdr" || events[9].([]any)[1] != 10 || events[9].([]any)[2] != 1 || events[9].([]any)[6] != 1 {
		t.Fatalf("pickup_dropped tuple = %#v", events[9])
	}
}

func TestCompactWirePacketLeavesUnknownEventRecordsMapShaped(t *testing.T) {
	input := map[string]any{
		"type": "event_batch",
		"batch_id": "event-batch-11",
		"events": []any{
			map[string]any{"event_id": "presentation-event-11", "type": "new_future_event", "source_id": "ship-1", "note": "kept"},
		},
	}

	got := CompactWirePacket(input)
	if got["t"] != "eb" {
		t.Fatalf("type = %v, want eb", got["t"])
	}
	if got["bid"] != 11 {
		t.Fatalf("batch_id = %v, want 11", got["bid"])
	}
	events := got["ev"].([]any)
	if len(events) != 1 {
		t.Fatalf("events = %#v, want 1 item", events)
	}
	record, ok := events[0].(map[string]any)
	if !ok {
		t.Fatalf("unknown event = %#v, want map shaped record", events[0])
	}
	if record["ei"] != "presentation-event-11" {
		t.Fatalf("event_id = %#v, want presentation-event-11", record["ei"])
	}
	if record["t"] != "new_future_event" {
		t.Fatalf("event type = %#v, want new_future_event", record["t"])
	}
	if record["src"] != "ship-1" || record["note"] != "kept" {
		t.Fatalf("unknown event aliases changed: %#v", record)
	}
}

func TestCompactWirePacketCompactsEventTypeAliases(t *testing.T) {
	input := map[string]any{
		"events": []any{
			map[string]any{"type": "radial_effect_started"},
			map[string]any{"type": "pickup_collected"},
			map[string]any{"type": "pickup_effect_applied"},
			map[string]any{"type": "pickup_expired"},
			map[string]any{"type": "pickup_dropped"},
			map[string]any{"type": "damage_over_time_started"},
			map[string]any{"type": "damage_over_time_tick"},
		},
	}

	got := CompactWirePacket(input)
	events := got["ev"].([]any)
	want := []string{"rfx", "pcol", "pea", "pexp", "pdr", "dots", "dott"}
	for i, compactType := range want {
		if events[i].(map[string]any)["t"] != compactType {
			t.Fatalf("event %d type = %v, want %v", i, events[i].(map[string]any)["t"], compactType)
		}
	}
}

func TestCompactWirePacketCompactAliasCollisionGuard(t *testing.T) {
	seen := map[string]string{}
	for key, value := range realtimewire.RealtimeWireKeyCompactByReadable {
		if prior, ok := seen[value]; ok {
			t.Fatalf("compact alias %q used by both %q and %q", value, prior, key)
		}
		seen[value] = key
	}
}





func TestCompactWirePacketTuplePacksWorldFullShipUnknownTargetKindUsesTaggedTargetID(t *testing.T) {
	input := map[string]any{
		"type": "world_full",
		"ships": []any{
			map[string]any{
				"id":          "player-3",
				"ship_type":   "v_wing",
				"target_kind": "mystery",
				"target_id":   "player-2",
			},
		},
	}

	got := CompactWirePacket(input)
	ships := got["ships"].([]any)
	tuple := ships[0].([]any)
	if tuple[0] != 3 || !reflect.DeepEqual(tuple[9], []any{"p", 2}) {
		t.Fatalf("unexpected compact ship tuple: %#v", tuple)
	}
}


func TestCompactWirePacketCompactsKnownEventSourceIDsByTupleContext(t *testing.T) {
	input := map[string]any{
		"type": "event_batch",
		"batch_id": "event-batch-1",
		"events": []any{
			map[string]any{"event_id": "presentation-event-1", "type": "damage_applied", "source_type": "mystery", "source_id": "player-2", "effect_type": "blast", "amount": 1},
		},
	}

	got := CompactWirePacket(input)
	event := got["ev"].([]any)[0].([]any)
	if !reflect.DeepEqual(event[3], []any{"p", 2}) {
		t.Fatalf("damage_applied source_id = %#v, want tagged player id", event[3])
	}
}


