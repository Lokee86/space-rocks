package game

import (
	"sync"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/awards"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/bots"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/damage"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/drops"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/effects/radial"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/encounterlifecycle"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/encounterspawn"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/modes"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/objectives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/participation"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rng"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/scoring"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spatial"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spatial/grid"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spawning"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/measurement"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

// Game keeps active runtime participation in playerSessions and durable match
// facts for everyone who participated in participantRecords.
type Game struct {
	mu                         sync.Mutex
	inputMu                    sync.Mutex
	pendingPlayerInputs        map[string]runtime.InputState
	rngSource                  *rng.Source
	stopSimulation             chan struct{}
	startSimulationOnce        sync.Once
	stopSimulationOnce         sync.Once
	nextID                     int
	nextPickupID               int
	nextPresentationEventID    int
	nextAwardEventID           int
	awardClock                 float64
	matchID                    string
	matchTraceID               string
	modeID                     string
	resolvedMatchRules         modes.ResolvedMatchRules
	matchElapsed               float64
	scoreCompletionTimes       map[string]float64
	scoreSuccessOrders         map[string]int
	nextScoreSuccessOrder      int
	teamStructure              teams.Structure
	spawner                    *spawning.Spawner
	damagePolicy               damage.Policy
	damageOverTimeRuntime      *damage.DamageOverTimeRuntime
	scoringPolicy              scoring.Policy
	awardPolicy                awards.Policy
	awardRuntime               *awards.Runtime
	dropTables                 drops.Tables
	radialEffects              radial.Store
	encounterSpawnRuntime      *encounterspawn.Runtime
	asteroidSpawnElapsed       float64
	worldSimulationOptions     WorldSimulationOptions
	collisionShapes            physics.CollisionShapeCatalog
	entities                   runtime.EntityStore
	spatialIndex               spatial.Index
	spatialEntries             []spatial.Entry
	spatialRefs                []spatial.Ref
	collisionPlayerIDs         []string
	collisionProjectileIDs     []string
	simulationStepObservers    []simulationStepObserver
	simulationStepObserverKeys map[string]struct{}
	awardEventObservers        []func(awards.EventResult)
	objectiveRuntime           *objectives.Runtime
	objectiveEventObservers    []func(objectives.Event)
	lockedFinalMatchState      *FinalMatchState
	encounterLifecycleRuntime  *encounterlifecycle.Runtime
	cameraViews                map[string]*runtime.CameraView
	playerSessions             map[string]*playerSession
	botControllers             map[string]*bots.Controller
	lifeRuntime                *lives.Runtime
	participationRuntime       *participation.Runtime
	participantRecords         map[string]*participantRecord
	pendingPresentationEvents  map[string][]PendingPresentationEvent
	presentationFrame          *gameplayPresentationFrame
	presentationDerivedMu      sync.Mutex
	presentationDerived        map[string][]presentationDerivedEntry
	runtimeMeasurementMu       sync.RWMutex
	runtimeMeasurements        map[uint64]measurement.SimulationObserver
	nextMeasurementObserverID  uint64
}

func New() *Game {
	game, err := NewWithPolicies(lives.NewBaselinePolicy(), participation.NewDefaultAFKPolicy())
	if err != nil {
		panic(err)
	}
	return game
}

func NewWithSeed(seed int64) *Game {
	game, err := NewWithSeedAndPolicies(seed, lives.NewBaselinePolicy(), participation.NewDefaultAFKPolicy())
	if err != nil {
		panic(err)
	}
	return game
}

func newGame(source *rng.Source) *Game {
	game, err := newGameWithPolicies(source, lives.NewBaselinePolicy(), participation.NewDefaultAFKPolicy())
	if err != nil {
		panic(err)
	}
	return game
}

func NewWithPolicy(policy lives.Policy) (*Game, error) {
	return NewWithPolicies(policy, participation.NewDefaultAFKPolicy())
}

func NewWithSeedAndPolicy(seed int64, policy lives.Policy) (*Game, error) {
	return NewWithSeedAndPolicies(seed, policy, participation.NewDefaultAFKPolicy())
}

func newGameWithPolicy(source *rng.Source, policy lives.Policy) (*Game, error) {
	return newGameWithPolicies(source, policy, participation.NewDefaultAFKPolicy())
}

func NewWithPolicies(policy lives.Policy, afkPolicy participation.AFKPolicy) (*Game, error) {
	return newGameWithPolicies(rng.NewProduction(), policy, afkPolicy)
}

func NewWithSeedAndPolicies(seed int64, policy lives.Policy, afkPolicy participation.AFKPolicy) (*Game, error) {
	return newGameWithPolicies(rng.New(seed), policy, afkPolicy)
}

func newGameWithPolicies(source *rng.Source, policy lives.Policy, afkPolicy participation.AFKPolicy) (*Game, error) {
	collisionShapes, err := physics.LoadCollisionShapeCatalog()
	if err != nil {
		logging.Emit(observability.Request{
			Event:  observability.EventNameCollisionShapeMissing,
			Fields: observability.Fields{"failure_mode": "collision_shape_catalog_unavailable"},
		})
	}
	lifeRuntime, err := lives.NewRuntime(policy)
	if err != nil {
		return nil, err
	}
	participationRuntime, err := participation.NewRuntime(afkPolicy)
	if err != nil {
		return nil, err
	}
	encounterSpawnRuntime, err := newEncounterSpawnRuntime()
	if err != nil {
		return nil, err
	}
	resolvedRules, err := modes.Resolve(modes.DefaultRoomModeConfig(), teams.Config{Structure: teams.StructureFFA})
	if err != nil {
		return nil, err
	}
	resolvedRules.LivesPolicy = policy

	game := &Game{
		rngSource:                  source,
		pendingPlayerInputs:        make(map[string]runtime.InputState),
		simulationStepObserverKeys: make(map[string]struct{}),
		collisionShapes:            collisionShapes,
		stopSimulation:             make(chan struct{}),
		cameraViews:                make(map[string]*runtime.CameraView),
		playerSessions:             make(map[string]*playerSession),
		botControllers:             make(map[string]*bots.Controller),
		lifeRuntime:                lifeRuntime,
		participationRuntime:       participationRuntime,
		modeID:                     string(resolvedRules.ModeID),
		resolvedMatchRules:         modes.CloneResolvedMatchRules(resolvedRules),
		scoreCompletionTimes:       make(map[string]float64),
		scoreSuccessOrders:         make(map[string]int),
		participantRecords:         make(map[string]*participantRecord),
		teamStructure:              resolvedRules.TeamConfig.Structure,
		pendingPresentationEvents:  make(map[string][]PendingPresentationEvent),
		presentationDerived:        make(map[string][]presentationDerivedEntry),
		runtimeMeasurements:        make(map[uint64]measurement.SimulationObserver),
		spawner:                    spawning.New(source),
		encounterSpawnRuntime:      encounterSpawnRuntime,
		encounterLifecycleRuntime:  encounterlifecycle.NewRuntime(),
		damagePolicy:               damage.NewStandardPolicy(),
		damageOverTimeRuntime:      damage.NewDamageOverTimeRuntime(),
		scoringPolicy:              scoring.NewDefaultPolicy(),
		awardPolicy:                awards.NewStandardPolicy(),
		awardRuntime:               awards.NewRuntime(),
		objectiveRuntime:           objectives.NewRuntime(),
		dropTables:                 drops.GeneratedTables,
		radialEffects:              radial.NewStore(),
		entities:                   runtime.NewEntityStore(),
		spatialIndex:               grid.New(space.DefaultBounds(), defaultSpatialCellSize),
	}
	game.publishPresentationFrameLocked()
	return game, nil
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

func (game *Game) SetModeContext(modeID string) {
	game.mu.Lock()
	defer game.mu.Unlock()
	game.modeID = modeID
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
