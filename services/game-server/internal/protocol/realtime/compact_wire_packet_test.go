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
	if _, ok := got["baseline_sequence"]; ok {
		t.Fatalf("did not expect readable baseline_sequence key in compact output: %#v", got)
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
			map[string]any{
				"event_id": "event-1",
				"type": "bullet_blast",
				"x": 10,
				"y": 20,
			},
			map[string]any{
				"event_id": "event-2",
				"type": "damage_applied",
				"source_type": "projectile",
				"source_id": "bullet-1",
				"target_type": "player",
				"target_id": "player-1",
				"damage_type": "explosive",
				"damage_cause": "impact",
				"base_amount": 20,
				"modified_amount": 17,
				"applied_to_health": 12,
				"absorbed_by_shield": 5,
				"remaining_health": 88,
				"remaining_shield": 0,
				"effect_type": "blast",
				"amount": 17,
			},
		},
	}

	got := CompactWirePacket(input)

	if got["t"] != "eb" {
		t.Fatalf("type = %v, want eb", got["t"])
	}
	if got["bid"] != "event-batch-11" {
		t.Fatalf("batch_id = %v, want event-batch-11", got["bid"])
	}
	events := got["ev"].([]any)
	if len(events) != 2 {
		t.Fatalf("events = %#v, want 2 items", events)
	}
	first := events[0].(map[string]any)
	if first["ei"] != "event-1" {
		t.Fatalf("event_id = %v, want event-1", first["ei"])
	}
	if first["t"] != "bb" {
		t.Fatalf("event type = %v, want bb", first["t"])
	}
	if first["x"] != 10 || first["y"] != 20 {
		t.Fatalf("event coordinates changed: %#v", first)
	}
	second := events[1].(map[string]any)
	if second["ei"] != "event-2" {
		t.Fatalf("event_id = %v, want event-2", second["ei"])
	}
	if second["t"] != "dmg" {
		t.Fatalf("event type = %v, want dmg", second["t"])
	}
	if second["srct"] != "projectile" || second["src"] != "bullet-1" {
		t.Fatalf("source aliases not compacted: %#v", second)
	}
	if second["tt"] != "player" || second["tid"] != "player-1" {
		t.Fatalf("target aliases not compacted: %#v", second)
	}
	if second["dt"] != "explosive" || second["dc"] != "impact" {
		t.Fatalf("damage aliases not compacted: %#v", second)
	}
	if second["ba"] != 20 || second["ma"] != 17 || second["ah"] != 12 || second["abs"] != 5 || second["rh"] != 88 || second["rs"] != 0 {
		t.Fatalf("damage value aliases not compacted: %#v", second)
	}
	if second["fx"] != "blast" || second["amt"] != 17 {
		t.Fatalf("effect/amount aliases not compacted: %#v", second)
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
	for key, value := range compactWireKeyMap {
		if prior, ok := seen[value]; ok {
			t.Fatalf("compact alias %q used by both %q and %q", value, prior, key)
		}
		seen[value] = key
	}
}
