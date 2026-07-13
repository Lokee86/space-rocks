package grid

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spatial"
)

func rectGrid() *Index { return New(space.Bounds{Width: 10, Height: 10}, 2) }

func rectEntry(kind spatial.Kind, id string, x, y, radius float64) spatial.Entry {
	return spatial.Entry{Ref: spatial.Ref{Kind: kind, ID: id}, Position: physics.Vector2{X: x, Y: y}, Radius: radius}
}

func rectIDs(refs []spatial.Ref) map[string]bool {
	ids := make(map[string]bool, len(refs))
	for _, ref := range refs { ids[ref.ID] = true }
	return ids
}

func TestQueryRectFiltersMasksAndAdjacentCells(t *testing.T) {
	index := rectGrid()
	index.Rebuild([]spatial.Entry{
		rectEntry(spatial.KindAsteroid, "near", 3, 1, 0),
		rectEntry(spatial.KindPlayer, "player", 1, 3, 0),
		rectEntry(spatial.KindPickup, "far", 8, 8, 0),
	})
	refs := index.QueryRect(nil, spatial.Rect{Center: physics.Vector2{X: 2, Y: 2}, HalfExtents: physics.Vector2{X: 2, Y: 2}}, spatial.KindMask(spatial.KindAsteroid))
	if len(refs) != 1 || refs[0].ID != "near" { t.Fatalf("query = %#v, want near only", refs) }
	refs = index.QueryRect(nil, spatial.Rect{Center: physics.Vector2{X: 2, Y: 2}, HalfExtents: physics.Vector2{X: 2, Y: 2}}, spatial.AllKinds)
	if len(refs) != 2 { t.Fatalf("all-kinds query returned %d refs, want 2", len(refs)) }
}

func TestQueryRectWrapsHorizontalVerticalAndCornerSeams(t *testing.T) {
	index := rectGrid()
	index.Rebuild([]spatial.Entry{
		rectEntry(spatial.KindAsteroid, "horizontal", 9.5, 0.5, 0),
		rectEntry(spatial.KindAsteroid, "vertical", 0.5, 9.5, 0),
		rectEntry(spatial.KindAsteroid, "corner", 9.5, 9.5, 0),
	})
	refs := index.QueryRect(nil, spatial.Rect{Center: physics.Vector2{}, HalfExtents: physics.Vector2{X: 1, Y: 1}}, spatial.AllKinds)
	ids := rectIDs(refs)
	for _, id := range []string{"horizontal", "vertical", "corner"} {
		if !ids[id] { t.Fatalf("missing wrapped result %q", id) }
	}
}

func TestQueryRectLargeEntriesAndFullWorld(t *testing.T) {
	index := rectGrid()
	index.Rebuild([]spatial.Entry{rectEntry(spatial.KindAsteroid, "large", 5, 5, 10)})
	refs := index.QueryRect(nil, spatial.Rect{Center: physics.Vector2{X: 5, Y: 5}}, spatial.AllKinds)
	if len(refs) != 1 || refs[0].ID != "large" { t.Fatalf("large entry query = %#v", refs) }
	index.Rebuild([]spatial.Entry{
		rectEntry(spatial.KindAsteroid, "one", 1, 1, 0),
		rectEntry(spatial.KindPickup, "two", 9, 9, 0),
	})
	refs = index.QueryRect(nil, spatial.Rect{HalfExtents: physics.Vector2{X: 5, Y: 5}}, spatial.AllKinds)
	if len(refs) != 2 { t.Fatalf("full-world query returned %d refs, want 2", len(refs)) }
}

func TestQueryRectNegativeExtentsAndDestinationAppend(t *testing.T) {
	index := rectGrid()
	index.Rebuild([]spatial.Entry{rectEntry(spatial.KindAsteroid, "point", 2, 2, 0)})
	dst := []spatial.Ref{{Kind: spatial.KindPickup, ID: "existing"}}
	refs := index.QueryRect(dst, spatial.Rect{Center: physics.Vector2{X: 2, Y: 2}, HalfExtents: physics.Vector2{X: -1, Y: -1}}, spatial.AllKinds)
	if len(refs) != 2 || refs[0].ID != "existing" || refs[1].ID != "point" { t.Fatalf("query = %#v", refs) }
	if refs = index.QueryRect(dst, spatial.Rect{}, 0); len(refs) != 1 { t.Fatal("zero mask changed destination") }
}

func TestQueryRectAllocatesNoStorageAfterWarmup(t *testing.T) {
	index := rectGrid()
	entries := []spatial.Entry{rectEntry(spatial.KindAsteroid, "one", 1, 1, 0), rectEntry(spatial.KindAsteroid, "two", 3, 3, 1)}
	index.Rebuild(entries)
	dst := make([]spatial.Ref, 0, 4)
	index.QueryRect(dst, spatial.Rect{Center: physics.Vector2{X: 2, Y: 2}, HalfExtents: physics.Vector2{X: 3, Y: 3}}, spatial.AllKinds)
	allocations := testing.AllocsPerRun(100, func() { index.QueryRect(dst[:0], spatial.Rect{Center: physics.Vector2{X: 2, Y: 2}, HalfExtents: physics.Vector2{X: 3, Y: 3}}, spatial.AllKinds) })
	if allocations != 0 { t.Fatalf("query allocations = %v, want 0", allocations) }
}
