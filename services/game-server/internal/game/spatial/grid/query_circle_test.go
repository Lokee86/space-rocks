package grid

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spatial"
)

func queryGrid() *Index { return New(space.Bounds{Width: 10, Height: 10}, 2) }

func queryEntry(kind spatial.Kind, id string, x, y, radius float64) spatial.Entry {
	return spatial.Entry{Ref: spatial.Ref{Kind: kind, ID: id}, Position: physics.Vector2{X: x, Y: y}, Radius: radius}
}

func queryIDs(refs []spatial.Ref) map[string]bool {
	ids := make(map[string]bool, len(refs))
	for _, ref := range refs { ids[ref.ID] = true }
	return ids
}

func TestQueryCircleFiltersKindsAndDeduplicates(t *testing.T) {
	index := queryGrid()
	index.Rebuild([]spatial.Entry{
		queryEntry(spatial.KindAsteroid, "near", 1, 1, 0),
		queryEntry(spatial.KindPlayer, "player", 3, 1, 0),
		queryEntry(spatial.KindPickup, "far", 8, 8, 0),
	})

	refs := index.QueryCircle(nil, physics.Vector2{X: 1, Y: 1}, 3, spatial.KindMask(spatial.KindAsteroid))
	if len(refs) != 1 || refs[0].ID != "near" { t.Fatalf("query = %#v, want near only", refs) }

	refs = index.QueryCircle(nil, physics.Vector2{X: 1, Y: 1}, 3, spatial.AllKinds)
	if len(refs) != 2 { t.Fatalf("all-kinds query returned %d refs, want 2", len(refs)) }
}

func TestQueryCircleWrapsAllSeams(t *testing.T) {
	index := queryGrid()
	index.Rebuild([]spatial.Entry{
		queryEntry(spatial.KindAsteroid, "horizontal", 9.5, 5, 0),
		queryEntry(spatial.KindAsteroid, "vertical", 5, 9.5, 0),
		queryEntry(spatial.KindAsteroid, "corner", 9.5, 9.5, 0),
	})
	refs := index.QueryCircle(nil, physics.Vector2{X: 0, Y: 0}, 1, spatial.AllKinds)
	ids := queryIDs(refs)
	for _, id := range []string{"horizontal", "vertical", "corner"} {
		if !ids[id] { t.Fatalf("missing wrapped result %q", id) }
	}
}

func TestQueryCircleLargeEntryAndDestinationAppend(t *testing.T) {
	index := queryGrid()
	index.Rebuild([]spatial.Entry{queryEntry(spatial.KindAsteroid, "large", 5, 5, 10)})
	dst := []spatial.Ref{{Kind: spatial.KindPickup, ID: "existing"}}
	refs := index.QueryCircle(dst, physics.Vector2{X: 5, Y: 5}, 0, spatial.AllKinds)
	if len(refs) != 2 || refs[0].ID != "existing" || refs[1].ID != "large" { t.Fatalf("query = %#v, want appended unique result", refs) }
}

func TestQueryCircleRejectsEmptyQueriesAndStaleEntries(t *testing.T) {
	index := queryGrid()
	index.Rebuild([]spatial.Entry{queryEntry(spatial.KindAsteroid, "old", 1, 1, 0)})
	index.Rebuild([]spatial.Entry{queryEntry(spatial.KindAsteroid, "new", 1, 1, 0)})
	if refs := index.QueryCircle(nil, physics.Vector2{X: 1, Y: 1}, 1, 0); len(refs) != 0 { t.Fatal("zero mask returned results") }
	if refs := index.QueryCircle(nil, physics.Vector2{X: 1, Y: 1}, -1, spatial.AllKinds); len(refs) != 0 { t.Fatal("negative radius returned results") }
	refs := index.QueryCircle(nil, physics.Vector2{X: 1, Y: 1}, 1, spatial.AllKinds)
	if len(refs) != 1 || refs[0].ID != "new" { t.Fatalf("stale query = %#v", refs) }
}

func TestQueryCircleAllocatesNoStorageAfterWarmup(t *testing.T) {
	index := queryGrid()
	entries := []spatial.Entry{queryEntry(spatial.KindAsteroid, "one", 1, 1, 0), queryEntry(spatial.KindAsteroid, "two", 3, 3, 1)}
	index.Rebuild(entries)
	dst := make([]spatial.Ref, 0, 4)
	index.QueryCircle(dst, physics.Vector2{X: 2, Y: 2}, 3, spatial.AllKinds)
	allocations := testing.AllocsPerRun(100, func() { index.QueryCircle(dst[:0], physics.Vector2{X: 2, Y: 2}, 3, spatial.AllKinds) })
	if allocations != 0 { t.Fatalf("query allocations = %v, want 0", allocations) }
}
