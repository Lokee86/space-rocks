package game

import (
	"sync"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/bots"
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
	"github.com/Lokee86/space-rocks/services/game-server/internal/measurement"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

// Game keeps active runtime participation in playerSessions and durable match
// facts for everyone who participated in participantRecords.
type Game struct {
	mu                        sync.Mutex
	inputMu                   sync.Mutex
	pendingPlayerInputs       map[string]runtime.InputState
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
	botControllers            map[string]*bots.Controller
	participantRecords        map[string]*participantRecord
	pendingPresentationEvents map[string][]PendingPresentationEvent
	presentationFrame         *gameplayPresentationFrame
	runtimeMeasurementMu      sync.RWMutex
	runtimeMeasurements       map[uint64]measurement.SimulationObserver
	nextMeasurementObserverID uint64
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
		pendingPlayerInputs:       make(map[string]runtime.InputState),
		collisionShapes:           collisionShapes,
		stopSimulation:            make(chan struct{}),
		cameraViews:               make(map[string]*runtime.CameraView),
		playerSessions:            make(map[string]*playerSession),
		botControllers:            make(map[string]*bots.Controller),
		participantRecords:        make(map[string]*participantRecord),
		pendingPresentationEvents: make(map[string][]PendingPresentationEvent),
		runtimeMeasurements:       make(map[uint64]measurement.SimulationObserver),
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
