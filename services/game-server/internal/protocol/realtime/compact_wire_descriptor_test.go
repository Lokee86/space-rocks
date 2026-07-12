package realtime

import (
	"reflect"
	"testing"
)

func TestCompactWireDescriptorYOnlySparsePlaceholder(t *testing.T) {
	packet := map[string]any{
		"type":         "world_delta",
		"ship_updates": []any{map[string]any{"id": "player-7", "y": 12}},
	}
	got := compactWirePacketFromDescriptors(packet)
	want := map[string]any{"t": "wd", "su": []any{[]any{7, nil, 12}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestCompactWireDescriptorPreservesSessionUnknownFields(t *testing.T) {
	packet := map[string]any{
		"type": "session_delta",
		"player_session_updates": []any{map[string]any{
			"id": "player-2", "score": 4, "future_field": "kept",
		}},
	}
	got := compactWirePacketFromDescriptors(packet)
	want := map[string]any{"t": "sd", "psu": []any{[]any{2, "sco", 4, "future_field", "kept"}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestCompactWireDescriptorAcceptsCompactSessionKeysAndFields(t *testing.T) {
	packet := map[string]any{
		"type": "session_delta",
		"pl":   []any{map[string]any{"i": "player-1", "st": "scout", "sco": 3}},
		"psu":  []any{map[string]any{"i": "player-2", "sco": 4, "future_field": "kept"}},
		"plu":  []any{map[string]any{"pid": "player-3", "stat": "active"}},
	}
	got := compactWirePacketFromDescriptors(packet)
	want := map[string]any{
		"t":   "sd",
		"pl":  []any{[]any{1, "scout", 3, nil, nil, nil, nil, nil, nil, nil, nil}},
		"psu": []any{[]any{2, "sco", 4, "future_field", "kept"}},
		"plu": []any{[]any{3, "active"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestCompactWireDescriptorAcceptsCompactLifecycleKeysAndFields(t *testing.T) {
	packet := map[string]any{
		"type": "asteroids_lifecycle",
		"ac": []any{map[string]any{
			"i": "asteroid-2", "h": 10, "sl": 1000, "sz": 1, "v": 0, "x": 1, "y": 2,
		}},
		"ax": []any{"asteroid-3", "asteroid-bad"},
	}
	got := compactWirePacketFromDescriptors(packet)
	want := map[string]any{
		"t":  "al",
		"ac": []any{map[string]any{"h": 10, "i": "asteroid-2", "sl": 1000, "sz": 1, "v": 0, "x": 1, "y": 2}},
		"ax": []any{"asteroid-3", "asteroid-bad"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestCompactWireDescriptorPreservesMalformedIDs(t *testing.T) {
	packet := map[string]any{
		"type":      "world_full",
		"asteroids": []any{map[string]any{"id": "asteroid-invalid", "x": 1}},
	}
	got := compactWirePacketFromDescriptors(packet)
	asteroids := got["asteroids"].([]any)
	if asteroids[0].([]any)[0] != "asteroid-invalid" {
		t.Fatalf("malformed ID changed: %#v", asteroids)
	}
}

func TestCompactWireDescriptorSelectorTaggedFallback(t *testing.T) {
	packet := map[string]any{
		"type": "world_full",
		"ships": []any{map[string]any{
			"id": "player-1", "target_kind": "unknown", "target_id": "player-9",
		}},
	}
	got := compactWirePacketFromDescriptors(packet)
	ship := got["ships"].([]any)[0].([]any)
	if !reflect.DeepEqual(ship[9], []any{"p", 9}) {
		t.Fatalf("selector fallback = %#v, want [p 9]", ship[9])
	}
}

func TestCompactWireDescriptorUnknownEventPassesThroughAsMap(t *testing.T) {
	packet := map[string]any{
		"type":   "event_batch",
		"events": []any{map[string]any{"type": "future_event", "id": "opaque-1"}},
	}
	got := compactWirePacketFromDescriptors(packet)
	event := got["ev"].([]any)[0].(map[string]any)
	if event["t"] != "future_event" || event["i"] != "opaque-1" {
		t.Fatalf("unknown event changed: %#v", event)
	}
}
