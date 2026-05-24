package game

import (
	"sync"

	"github.com/Lokee86/space-rocks/server/internal/game/devtools"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/scoring"
	"github.com/Lokee86/space-rocks/server/internal/game/spawning"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

type Game struct {
	mu                        sync.Mutex
	stopSimulation            chan struct{}
	stopSimulationOnce        sync.Once
	nextID                    int
	spawner                   *spawning.Spawner
	scoringPolicy             scoring.Policy
	asteroidSpawnElapsed      float64
	worldDevTools             devtools.WorldOptions
	collisionShapes           physics.CollisionShapeCatalog
	state                     entities.GameState
	cameraViews               map[string]*entities.CameraView
	playerSessions            map[string]*playerSession
	pendingPresentationEvents map[string][]EventState
}

func New() *Game {
	collisionShapes, err := physics.LoadCollisionShapeCatalog()
	if err != nil {
		logging.Game.Warn("collision shapes unavailable", logging.FieldError, err)
	}

	return &Game{
		collisionShapes:           collisionShapes,
		stopSimulation:            make(chan struct{}),
		cameraViews:               make(map[string]*entities.CameraView),
		playerSessions:            make(map[string]*playerSession),
		pendingPresentationEvents: make(map[string][]EventState),
		spawner:                   spawning.New(),
		scoringPolicy:             scoring.NewDefaultPolicy(),
		state:                     entities.NewGameState(),
	}
}

func (game *Game) Start() {
	go game.runSimulation()
}

func (game *Game) Stop() {
	game.stopSimulationOnce.Do(func() {
		close(game.stopSimulation)
	})
}
