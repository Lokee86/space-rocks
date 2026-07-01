package quantize

import "testing"

type rawFloatLeaf struct {
	Value float64
}

type rawFloatNested struct {
	Child rawFloatLeaf
}

type safeStruct struct {
	Count  int
	Label  string
	Active bool
}

func TestCollectRawFloatLeaksDetectsStructFloat(t *testing.T) {
	leaks := CollectRawFloatLeaks("packet", rawFloatLeaf{Value: 12.5})
	if len(leaks) != 1 {
		t.Fatalf("expected 1 leak, got %d", len(leaks))
	}
	if leaks[0].FieldPath != "packet.value" {
		t.Fatalf("expected field path packet.value, got %q", leaks[0].FieldPath)
	}
}

func TestCollectRawFloatLeaksDetectsNestedStructFloat(t *testing.T) {
	leaks := CollectRawFloatLeaks("packet", rawFloatNested{Child: rawFloatLeaf{Value: 9.25}})
	if len(leaks) != 1 {
		t.Fatalf("expected 1 leak, got %d", len(leaks))
	}
	if leaks[0].FieldPath != "packet.child.value" {
		t.Fatalf("expected field path packet.child.value, got %q", leaks[0].FieldPath)
	}
}

func TestCollectRawFloatLeaksIgnoresSafeStruct(t *testing.T) {
	if leaks := CollectRawFloatLeaks("packet", safeStruct{Count: 3, Label: "ok", Active: true}); len(leaks) != 0 {
		t.Fatalf("expected no leaks, got %d", len(leaks))
	}
}

func TestCollectRawFloatLeaksDetectsMapFloat(t *testing.T) {
	leaks := CollectRawFloatLeaks("packet", map[string]any{"value": 2.5})
	if len(leaks) != 1 {
		t.Fatalf("expected 1 leak, got %d", len(leaks))
	}
	if leaks[0].FieldPath != "packet.value" {
		t.Fatalf("expected field path packet.value, got %q", leaks[0].FieldPath)
	}
}
