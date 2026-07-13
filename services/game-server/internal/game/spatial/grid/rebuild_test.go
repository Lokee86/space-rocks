package grid

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spatial"
)

func testGrid() *Index {
	return New(space.Bounds{Width: 10, Height: 10}, 2)
}

func testEntry(id string, x, y, radius float64) spatial.Entry {
	return spatial.Entry{Ref: spatial.Ref{Kind: spatial.KindAsteroid, ID: id}, Position: physics.Vector2{X: x, Y: y}, Radius: radius}
}

func TestNewAllocatesCeilingDimensions(t *testing.T) {
	index := New(space.Bounds{Width: 10, Height: 7}, 3)
	if index.cellsX != 4 || index.cellsY != 3 {
		t.Fatalf("dimensions = %dx%d, want 4x3", index.cellsX, index.cellsY)
	}
	if len(index.buckets) != 12 {
		t.Fatalf("bucket count = %d, want 12", len(index.buckets))
	}
}

func TestRebuildNormalizesPositionAndClampsRadius(t *testing.T) {
	index := testGrid()
	index.Rebuild([]spatial.Entry{testEntry("normalized", -1, 11, -2)})

	if len(index.buckets[0]) != 1 {
		t.Fatalf("normalized entry was not inserted into bucket 0")
	}
	entry := index.buckets[0][0]
	if entry.Position != (physics.Vector2{X: 9, Y: 1}) || entry.Radius != 0 {
		t.Fatalf("stored entry = %#v, want normalized position and zero radius", entry)
	}
}

func TestRebuildInsertsMultipleCells(t *testing.T) {
	index := testGrid()
	index.Rebuild([]spatial.Entry{testEntry("wide", 5, 5, 2.1)})
	if len(index.touched) != 9 {
		t.Fatalf("touched buckets = %d, want 9", len(index.touched))
	}
}

func TestRebuildWrapsInsertionAcrossEdges(t *testing.T) {
	index := testGrid()
	index.Rebuild([]spatial.Entry{testEntry("wrapped", 0.5, 0.5, 1)})

	for _, bucket := range []int{0, 4, 20, 24} {
		if len(index.buckets[bucket]) != 1 {
			t.Fatalf("bucket %d has %d entries, want 1", bucket, len(index.buckets[bucket]))
		}
	}
}

func TestRebuildPreventsFullAxisDuplicates(t *testing.T) {
	index := testGrid()
	index.Rebuild([]spatial.Entry{testEntry("full", 5, 5, 10)})
	for bucket, entries := range index.buckets {
		if len(entries) != 1 {
			t.Fatalf("bucket %d has %d entries, want 1", bucket, len(entries))
		}
	}
}

func TestRebuildRemovesStaleEntries(t *testing.T) {
	index := testGrid()
	index.Rebuild([]spatial.Entry{testEntry("old", 2, 2, 0)})
	index.Rebuild([]spatial.Entry{testEntry("new", 8, 8, 0)})

	for _, entries := range index.buckets {
		for _, entry := range entries {
			if entry.Ref.ID == "old" {
				t.Fatal("stale entry remained after rebuild")
			}
		}
	}
}

func TestRebuildClearsQueryDeduplicationState(t *testing.T) {
	index := testGrid()
	index.Rebuild([]spatial.Entry{testEntry("old", 2, 2, 10)})
	oldRefs := index.QueryCircle(nil, physics.Vector2{X: 2, Y: 2}, 10, spatial.AllKinds)
	if len(oldRefs) != 1 || oldRefs[0].ID != "old" {
		t.Fatalf("old refs = %#v", oldRefs)
	}

	index.Rebuild([]spatial.Entry{testEntry("new", 8, 8, 10)})
	newRefs := index.QueryCircle(nil, physics.Vector2{X: 8, Y: 8}, 10, spatial.AllKinds)
	if len(newRefs) != 1 || newRefs[0].ID != "new" {
		t.Fatalf("new refs = %#v", newRefs)
	}
	if len(index.querySeen) != 1 || index.queryGeneration != 1 {
		t.Fatalf("query state = %d refs, generation %d", len(index.querySeen), index.queryGeneration)
	}
}

func TestRebuildAllocatesNoTemporaryStorageAfterWarmup(t *testing.T) {
	index := testGrid()
	entries := []spatial.Entry{
		testEntry("one", 1, 1, 0),
		testEntry("two", 3, 3, 1),
		testEntry("three", 6, 6, 2.1),
	}
	index.Rebuild(entries)

	allocations := testing.AllocsPerRun(100, func() {
		index.Rebuild(entries)
	})
	if allocations != 0 {
		t.Fatalf("repeated rebuild allocations = %v, want 0", allocations)
	}
}

func TestRebuildUsesExactNonDivisibleToroidalTiling(t *testing.T) {
	index := New(space.Bounds{Width: 10, Height: 7}, 3)
	index.Rebuild([]spatial.Entry{
		testEntry("right", 9.8, 1, 0.4),
		testEntry("bottom", 3, 6.8, 0.4),
		testEntry("corner", 9.8, 6.8, 0.4),
	})

	if !bucketContains(index.buckets[3], "right") || !bucketContains(index.buckets[0], "right") {
		t.Fatal("right-edge entry did not cross the horizontal seam")
	}
	if !bucketContains(index.buckets[9], "bottom") || !bucketContains(index.buckets[1], "bottom") {
		t.Fatal("bottom-edge entry did not cross the vertical seam")
	}
	for _, bucket := range []int{11, 8, 3, 0} {
		if !bucketContains(index.buckets[bucket], "corner") {
			t.Fatalf("bottom-right corner entry missing from bucket %d", bucket)
		}
	}
}

func bucketContains(entries []spatial.Entry, id string) bool {
	for _, entry := range entries {
		if entry.Ref.ID == id {
			return true
		}
	}
	return false
}
