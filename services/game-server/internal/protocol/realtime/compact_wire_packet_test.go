package realtime

import "testing"

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
}

func TestCompactWirePacketCompactsNestedWorldUpdatesRecursively(t *testing.T) {
	input := map[string]any{
		"type": "world_delta",
		"lane": "world",
		"ship_updates": []any{
			map[string]any{
				"id":        "ship-1",
				"x":         10,
				"y":         11,
				"rotation":  12,
				"thrusting": false,
				"nested": map[string]any{
					"owner_id": "player-9",
					"size":     3,
				},
			},
		},
	}

	got := CompactWirePacket(input)

	updates := got["su"].([]any)
	first := updates[0].(map[string]any)
	if first["i"] != "ship-1" {
		t.Fatalf("id = %v, want ship-1", first["i"])
	}
	if first["x"] != 10 || first["y"] != 11 {
		t.Fatalf("position changed: %#v", first)
	}
	if first["r"] != 12 {
		t.Fatalf("rotation = %v, want 12", first["r"])
	}
	if first["th"] != false {
		t.Fatalf("thrusting = %v, want false", first["th"])
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
func TestCompactWirePacketCompactsReadableWorldDeltaMap(t *testing.T) {
	input := map[string]any{
		"type":             "world_delta",
		"lane":             "world",
		"sequence":         int64(7),
		"baseline_id":      "player-1",
		"snapshot_id":      "player-1",
		"server_sent_msec": int64(123),
		"snapshot_kind":    "delta",
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

	for _, key := range []string{"t", "l", "q", "b", "sid", "ms", "k", "su"} {
		if _, ok := got[key]; !ok {
			t.Fatalf("expected compact key %q to be present, got %#v", key, got)
		}
	}
	for _, key := range []string{"type", "lane", "sequence", "baseline_id", "snapshot_id", "server_sent_msec", "snapshot_kind", "ship_updates"} {
		if _, ok := got[key]; ok {
			t.Fatalf("did not expect readable key %q in compact output: %#v", key, got)
		}
	}
}