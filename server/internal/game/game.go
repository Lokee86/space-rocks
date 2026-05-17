package game

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

type Game struct {
	mu      sync.Mutex
	nextID  int
	players map[string]*Player
}

func New() *Game {
	return &Game{
		players: make(map[string]*Player),
	}
}

func (game *Game) AddPlayer() string {
	game.mu.Lock()
	defer game.mu.Unlock()

	playerIndex := game.nextID
	game.nextID++

	playerID := fmt.Sprintf("player-%d", game.nextID)
	game.players[playerID] = &Player{
		X: 576 + float64(playerIndex%4)*80,
		Y: 320 + float64(playerIndex/4)*80,
	}

	return playerID
}

func (game *Game) RemovePlayer(playerID string) {
	game.mu.Lock()
	defer game.mu.Unlock()

	delete(game.players, playerID)
}

func (game *Game) HandlePacket(playerID string, packet InputPacket) []byte {
	game.mu.Lock()
	defer game.mu.Unlock()

	player, ok := game.players[playerID]
	if !ok {
		return nil
	}

	switch packet.Type {
	case "input":
		player.applyInput(packet.Input)

		if packet.Input.Shoot {
			log.Println("shoot")
		}
	}

	response, err := json.Marshal(game.statePacket(playerID))
	if err != nil {
		log.Println(err)
		return nil
	}

	return response
}

func (game *Game) statePacket(playerID string) StatePacket {
	players := make(map[string]PlayerState, len(game.players))
	for id, player := range game.players {
		players[id] = player.State()
	}

	return StatePacket{
		Type:    "state",
		SelfID:  playerID,
		Players: players,
	}
}
