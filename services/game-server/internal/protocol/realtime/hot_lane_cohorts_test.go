package realtime

import "testing"

func TestHotLaneCohortStateRemoveMissingAsteroids(t *testing.T) {
	state := HotLaneCohortState{
		AsteroidRoutes: map[string]HotUpdateRoute{
			"asteroid-1": HotUpdateRouteWorld,
			"asteroid-2": HotUpdateRouteAsteroids,
		},
	}

	state.RemoveMissingAsteroids(map[string]bool{
		"asteroid-2": true,
	})

	if _, ok := state.AsteroidRoutes["asteroid-1"]; ok {
		t.Fatal("expected stale asteroid route to be removed")
	}
	if got := state.AsteroidRoutes["asteroid-2"]; got != HotUpdateRouteAsteroids {
		t.Fatalf("expected active asteroid route to remain, got %#v", got)
	}
}

func TestHotLaneCohortStateRemoveMissingBullets(t *testing.T) {
	state := HotLaneCohortState{
		BulletRoutes: map[string]HotUpdateRoute{
			"bullet-1": HotUpdateRouteWorld,
			"bullet-2": HotUpdateRouteBullets,
		},
	}

	state.RemoveMissingBullets(map[string]bool{
		"bullet-2": true,
	})

	if _, ok := state.BulletRoutes["bullet-1"]; ok {
		t.Fatal("expected stale bullet route to be removed")
	}
	if got := state.BulletRoutes["bullet-2"]; got != HotUpdateRouteBullets {
		t.Fatalf("expected active bullet route to remain, got %#v", got)
	}
}
