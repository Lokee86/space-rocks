package realtime

import (
	"reflect"
	"testing"
)

type taggedWireFixture struct {
	SchemaName string `json:"schema_name"`
	LegacyName string
	Excluded   string `json:"-"`
	Optional   bool   `json:"optional,omitempty"`
	private    string
}

func TestWireStructToMapPrefersJSONTagsAndPreservesZeroValues(t *testing.T) {
	got := wireStructToMap(taggedWireFixture{
		SchemaName: "tagged",
		LegacyName: "legacy",
		Excluded:   "secret",
	})
	want := map[string]any{
		"schema_name": "tagged",
		"legacy_name": "legacy",
		"optional":    false,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("wireStructToMap() = %#v, want %#v", got, want)
	}
}

func TestWireValueHandlesPointersSlicesMapsAndNil(t *testing.T) {
	value := taggedWireFixture{SchemaName: "one"}
	got := wireValue(map[string]any{
		"records": []*taggedWireFixture{&value, nil},
	})
	want := map[string]any{
		"records": []any{
			map[string]any{"schema_name": "one", "legacy_name": "", "optional": false},
			nil,
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("wireValue() = %#v, want %#v", got, want)
	}
}

func TestWireStructToMapRejectsUnsupportedTopLevelValues(t *testing.T) {
	if got := wireStructToMap("not-a-record"); len(got) != 0 {
		t.Fatalf("wireStructToMap() = %#v, want empty map", got)
	}
}
