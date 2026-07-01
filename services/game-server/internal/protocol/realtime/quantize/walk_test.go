package quantize

import (
	"math"
	"testing"
)

func TestQuantizeValueConvertsNestedFloats(t *testing.T) {
	input := map[string]any{
		"session": map[string]any{
			"elapsed": 1.234,
		},
		"overlay": map[string]any{
			"debug": map[string]any{
				"fps": 59.9,
			},
		},
	}

	got, err := QuantizeValue("session", "", input)
	if err != nil {
		t.Fatalf("quantize: %v", err)
	}

	root, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("got type %T, want map[string]any", got)
	}

	session := root["session"].(map[string]any)
	if session["elapsed"] != int64(1234) {
		t.Fatalf("session.elapsed = %v, want 1234", session["elapsed"])
	}

	overlay := root["overlay"].(map[string]any)
	debug := overlay["debug"].(map[string]any)
	if debug["fps"] != int64(59900) {
		t.Fatalf("overlay.debug.fps = %v, want 59900", debug["fps"])
	}
}

func TestQuantizeValuePreservesNonFloatValues(t *testing.T) {
	input := map[string]any{
		"string": "hello",
		"bool":   true,
		"int":    int64(42),
		"nil":    nil,
		"array": []any{
			"item",
			false,
			int32(7),
			nil,
			map[string]any{"nested": "value"},
		},
	}

	got, err := QuantizeValue("session", "", input)
	if err != nil {
		t.Fatalf("quantize: %v", err)
	}

	root := got.(map[string]any)
	if root["string"] != "hello" {
		t.Fatalf("string = %v", root["string"])
	}
	if root["bool"] != true {
		t.Fatalf("bool = %v", root["bool"])
	}
	if root["int"] != int64(42) {
		t.Fatalf("int = %v", root["int"])
	}
	if root["nil"] != nil {
		t.Fatalf("nil = %v", root["nil"])
	}

	array := root["array"].([]any)
	if array[0] != "item" || array[1] != false || array[2] != int32(7) || array[3] != nil {
		t.Fatalf("array preserved incorrectly: %#v", array)
	}
	nested := array[4].(map[string]any)
	if nested["nested"] != "value" {
		t.Fatalf("nested value = %v", nested["nested"])
	}
}

func TestQuantizeValueFallsBackToFloatGenericForUnmappedFloats(t *testing.T) {
	input := map[string]any{
		"unmapped": 2.5,
	}

	got, err := QuantizeValue("session", "", input)
	if err != nil {
		t.Fatalf("quantize: %v", err)
	}

	root := got.(map[string]any)
	if root["unmapped"] != int64(2500) {
		t.Fatalf("unmapped = %v, want 2500", root["unmapped"])
	}
}

func TestQuantizeValueRejectsInvalidFloats(t *testing.T) {
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		_, err := QuantizeValue("session", "session.elapsed", value)
		if err == nil {
			t.Fatalf("expected error for %v", value)
		}
	}
}

func TestQuantizeValueDoesNotMutateOriginalInput(t *testing.T) {
	original := map[string]any{
		"session": map[string]any{
			"elapsed": 1.234,
		},
		"array": []any{
			map[string]any{"nested": 9.99},
		},
	}

	got, err := QuantizeValue("session", "", original)
	if err != nil {
		t.Fatalf("quantize: %v", err)
	}

	if original["session"].(map[string]any)["elapsed"] != 1.234 {
		t.Fatalf("original session.elapsed was mutated")
	}
	if original["array"].([]any)[0].(map[string]any)["nested"] != 9.99 {
		t.Fatalf("original array element was mutated")
	}

	result := got.(map[string]any)
	if result["session"].(map[string]any)["elapsed"] != int64(1234) {
		t.Fatalf("quantized copy not returned as expected")
	}
}
