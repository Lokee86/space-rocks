package realtime

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

type realtimeWireFixture struct {
	Name         string         `json:"name"`
	Readable     map[string]any `json:"readable"`
	Compact      map[string]any `json:"compact"`
	Expanded     map[string]any `json:"expanded"`
	ServerEncode *bool          `json:"server_encode,omitempty"`
}

func TestRealtimeWireFixturesMatchCurrentCompactor(t *testing.T) {
	fixtures := loadRealtimeWireFixtures(t)
	if len(fixtures) == 0 {
		t.Fatal("no realtime wire fixtures found")
	}
	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(fixture.Name, func(t *testing.T) {
			if fixture.ServerEncode != nil && !*fixture.ServerEncode {
				t.Skip("fixture records a client-only legacy decode input")
			}
			got := normalizeJSONValue(t, CompactWirePacket(fixture.Readable))
			want := normalizeJSONValue(t, fixture.Compact)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("compact fixture mismatch\ngot:  %#v\nwant: %#v", got, want)
			}
		})
	}
}

func loadRealtimeWireFixtures(t *testing.T) []realtimeWireFixture {
	t.Helper()
	root := filepath.Clean(filepath.Join("..", "..", "..", "..", "..", "shared", "packets", "fixtures", "realtime_wire"))
	var paths []string
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			paths = append(paths, path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk realtime wire fixtures: %v", err)
	}
	sort.Strings(paths)
	fixtures := make([]realtimeWireFixture, 0, len(paths))
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read fixture %s: %v", path, err)
		}
		var fixture realtimeWireFixture
		if err := json.Unmarshal(data, &fixture); err != nil {
			t.Fatalf("parse fixture %s: %v", path, err)
		}
		if fixture.Name == "" || fixture.Readable == nil || fixture.Compact == nil || fixture.Expanded == nil {
			t.Fatalf("fixture %s must contain name, readable, compact, and expanded", path)
		}
		fixtures = append(fixtures, fixture)
	}
	return fixtures
}

func normalizeJSONValue(t *testing.T, value any) any {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal fixture value: %v", err)
	}
	var normalized any
	if err := json.Unmarshal(data, &normalized); err != nil {
		t.Fatalf("normalize fixture value: %v", err)
	}
	return normalized
}
