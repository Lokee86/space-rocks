package game

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Lokee86/space-rocks/server/internal/constants"
)

type Game struct {
	mu     sync.Mutex
	nextID int
	state  GameState
}

func New() *Game {
	return &Game{
		state: NewGameState(),
	}
}

func (game *Game) Start() {
	go game.runSimulation()
}

func (game *Game) AddPlayer() string {
	game.mu.Lock()
	defer game.mu.Unlock()

	playerIndex := game.nextID
	game.nextID++

	playerID := fmt.Sprintf("player-%d", game.nextID)
	game.state.Players[playerID] = &Ship{
		ID: playerID,
		X:  576 + float64(playerIndex%4)*80,
		Y:  320 + float64(playerIndex/4)*80,
	}

	return playerID
}

func (game *Game) RemovePlayer(playerID string) {
	game.mu.Lock()
	defer game.mu.Unlock()

	delete(game.state.Players, playerID)
}

func (game *Game) HandlePacket(playerID string, packet InputPacket) {
	game.mu.Lock()
	defer game.mu.Unlock()

	player, ok := game.state.Players[playerID]
	if !ok {
		return
	}

	switch packet.Type {
	case "input":
		player.Input = packet.Input

		if packet.Input.Shoot {
			log.Println("shoot")
		}
	}
}

func (game *Game) State(playerID string) []byte {
	game.mu.Lock()
	defer game.mu.Unlock()

	response, err := json.Marshal(game.statePacket(playerID))
	if err != nil {
		log.Println(err)
		return nil
	}

	return response
}

func (game *Game) runSimulation() {
	ticker := time.NewTicker(time.Second / time.Duration(constants.ServerTickRate))
	defer ticker.Stop()

	delta := 1.0 / float64(constants.ServerTickRate)
	for range ticker.C {
		game.Step(delta)
	}
}

func (game *Game) Step(delta float64) {
	game.mu.Lock()
	defer game.mu.Unlock()

	for _, player := range game.state.Players {
		player.applyInput(delta)
	}
}

func (game *Game) statePacket(playerID string) StatePacket {
	players := make(map[string]ShipState, len(game.state.Players))
	for id, player := range game.state.Players {
		players[id] = player.State()
	}

	return StatePacket{
		Type:    "state",
		SelfID:  playerID,
		Players: players,
	}
}
