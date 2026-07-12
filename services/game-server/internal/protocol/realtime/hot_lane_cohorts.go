package realtime

type HotUpdateRoute string

const (
	HotUpdateRouteWorld     HotUpdateRoute = "world"
	HotUpdateRouteAsteroids HotUpdateRoute = "asteroids"
	HotUpdateRouteBullets   HotUpdateRoute = "bullets"
)

type HotLaneMode string

const (
	HotLaneModeInline        HotLaneMode = "inline"
	HotLaneModeOverflow      HotLaneMode = "overflow"
	HotLaneModeFullOwned60Hz HotLaneMode = "full_owned_60hz"
	HotLaneModeFullOwned30Hz HotLaneMode = "full_owned_30hz"
	HotLaneModeFullOwned20Hz HotLaneMode = "full_owned_20hz"
	HotLaneModeNeedsChunking HotLaneMode = "needs_chunking"
)

type HotLaneCohortState struct {
	AsteroidRoutes         map[string]HotUpdateRoute
	BulletRoutes           map[string]HotUpdateRoute
	AsteroidMode           HotLaneMode
	BulletMode             HotLaneMode
	StableLowPressureTicks int
}

func NewHotLaneCohortState() HotLaneCohortState {
	return HotLaneCohortState{
		AsteroidRoutes: make(map[string]HotUpdateRoute),
		BulletRoutes:   make(map[string]HotUpdateRoute),
		AsteroidMode:   HotLaneModeInline,
		BulletMode:     HotLaneModeInline,
	}
}

func (state *HotLaneCohortState) EnsureInitialized() {
	if state.AsteroidRoutes == nil {
		state.AsteroidRoutes = make(map[string]HotUpdateRoute)
	}
	if state.BulletRoutes == nil {
		state.BulletRoutes = make(map[string]HotUpdateRoute)
	}
	if state.AsteroidMode == "" {
		state.AsteroidMode = HotLaneModeInline
	}
	if state.BulletMode == "" {
		state.BulletMode = HotLaneModeInline
	}
}

func (state *HotLaneCohortState) RemoveMissingAsteroids(activeIDs map[string]bool) {
	state.EnsureInitialized()
	for asteroidID := range state.AsteroidRoutes {
		if !activeIDs[asteroidID] {
			delete(state.AsteroidRoutes, asteroidID)
		}
	}
}

func (state *HotLaneCohortState) RemoveMissingBullets(activeIDs map[string]bool) {
	state.EnsureInitialized()
	for bulletID := range state.BulletRoutes {
		if !activeIDs[bulletID] {
			delete(state.BulletRoutes, bulletID)
		}
	}
}
