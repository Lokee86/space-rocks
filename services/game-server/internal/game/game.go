package game

import (
	"sync"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/drops"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/effects/radial"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rng"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/scoring"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spatial"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spatial/grid"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spawning"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

type Game struct {
	mu                        sync.Mutex
	rngSource                 *rng.Source
	stopSimulation            chan struct{}
	startSimulationOnce       sync.Once
	stopSimulationOnce        sync.Once
	nextID                    int
	nextPickupID              int
	nextPresentationEventID   int
	matchID                   string
	matchTraceID              string
	spawner                   *spawning.Spawner
	scoringPolicy             scoring.Policy
	dropTables                drops.Tables
	radialEffects             radial.Store
	asteroidSpawnElapsed      float64
	worldSimulationOptions    WorldSimulationOptions
	collisionShapes           physics.CollisionShapeCatalog
	entities                  runtime.EntityStore
	spatialIndex              spatial.Index
	spatialEntries            []spatial.Entry
	spatialRefs               []spatial.Ref
	collisionPlayerIDs        []string
	collisionProjectileIDs    []string
	simulationStepObservers   []simulationStepObserver
	cameraViews               map[string]*runtime.CameraView
	playerSessions            map[string]*playerSession
	pendingPresentationEvents map[string][]PendingPresentationEvent
	presentationFrame         *gameplayPresentationFrame
}

func New() *Game {
	return newGame(rng.NewProduction())
}

func NewWithSeed(seed int64) *Game {
	return newGame(rng.New(seed))
}

func newGame(source *rng.Source) *Game {
	collisionShapes, err := physics.LoadCollisionShapeCatalog()
	if err != nil {
		logging.Emit(observability.Request{
			Event:  observability.EventNameCollisionShapeMissing,
			Fields: observability.Fields{"failure_mode": "collision_shape_catalog_unavailable"},
		})
	}

	game := &Game{
		rngSource:                 source,
		collisionShapes:           collisionShapes,
		stopSimulation:            make(chan struct{}),
		cameraViews:               make(map[string]*runtime.CameraView),
		playerSessions:            make(map[string]*playerSession),
		pendingPresentationEvents: make(map[string][]PendingPresentationEvent),
		spawner:                   spawning.New(source),
		scoringPolicy:             scoring.NewDefaultPolicy(),
		dropTables:                drops.GeneratedTables,
		radialEffects:             radial.NewStore(),
		entities:                  runtime.NewEntityStore(),
		spatialIndex:              grid.New(space.DefaultBounds(), defaultSpatialCellSize),
	}
	game.publishPresentationFrameLocked()
	return game
}

func (game *Game) SimulationSeed() int64 {
	return game.rngSource.Seed()
}

func (game *Game) SetMatchContext(matchID string, traceID string) {
	game.mu.Lock()
	defer game.mu.Unlock()
	game.matchID = matchID
	game.matchTraceID = traceID
}

func (game *Game) Start() {
	game.startSimulationOnce.Do(func() {
		go game.runSimulation()
	})
}

func (game *Game) Stop() {
	game.stopSimulationOnce.Do(func() {
		close(game.stopSimulation)
	})
}
