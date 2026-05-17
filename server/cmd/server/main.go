package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/gorilla/websocket"
)

type InputPacket struct {
	Type  string     `json:"type"`
	Input InputState `json:"input"`
}

type InputState struct {
	Forward bool `json:"forward"`
	Back    bool `json:"back"`
	Right   bool `json:"right"`
	Left    bool `json:"left"`
	Shoot   bool `json:"shoot"`
}

type Player struct {
	X        float64   `json:"x"`
	Y        float64   `json:"y"`
	Rotation float64   `json:"rotation"`
	Velocity Vector2   `json:"-"`
	LastTick time.Time `json:"-"`
}

type PlayerState struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Rotation float64 `json:"rotation"`
}

type StatePacket struct {
	Type    string                 `json:"type"`
	SelfID  string                 `json:"self_id"`
	Players map[string]PlayerState `json:"players"`
}

type Game struct {
	mu      sync.Mutex
	nextID  int
	players map[string]*Player
}

type Vector2 struct {
	X float64
	Y float64
}

func main() {
	mux := http.NewServeMux()
	game := NewGame()

	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /ws", game.wsHandler)

	fmt.Println("Server starting on :8080")

	log.Fatal(http.ListenAndServe(":8080", mux))

}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func NewGame() *Game {
	return &Game{
		players: make(map[string]*Player),
	}
}

func (game *Game) wsHandler(w http.ResponseWriter, r *http.Request) {

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer conn.Close()

	playerID := game.addPlayer()
	defer game.removePlayer(playerID)

	log.Println("client connected:", playerID)

	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		var packet InputPacket

		err = json.Unmarshal(msg, &packet)
		if err != nil {
			log.Println("bad packet:", err)
			continue
		}

		returnMsg := game.packetHandler(playerID, packet)

		err = conn.WriteMessage(messageType, returnMsg)
		if err != nil {
			log.Println(err)
			break
		}
	}
}

func (game *Game) addPlayer() string {
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

func (game *Game) removePlayer(playerID string) {
	game.mu.Lock()
	defer game.mu.Unlock()

	delete(game.players, playerID)
	log.Println("client disconnected:", playerID)
}

func (game *Game) packetHandler(playerID string, input InputPacket) []byte {
	game.mu.Lock()
	defer game.mu.Unlock()

	player, ok := game.players[playerID]
	if !ok {
		return nil
	}

	switch input.Type {
	case "input":
		player.applyInput(input.Input)

		if input.Input.Shoot {
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

func (player *Player) State() PlayerState {
	return PlayerState{
		X:        player.X,
		Y:        player.Y,
		Rotation: player.Rotation,
	}
}

func (player *Player) applyInput(input InputState) {
	delta := player.nextDelta()
	rotationInput := axis(input.Left, input.Right)
	thrustInput := axis(input.Back, input.Forward)

	player.Rotation += rotationInput * constants.PlayerRotationSpeed * delta

	if thrustInput != 0 {
		player.Velocity.X += math.Sin(player.Rotation) * constants.PlayerThrustForce * thrustInput * delta
		player.Velocity.Y += -math.Cos(player.Rotation) * constants.PlayerThrustForce * thrustInput * delta
	}

	damping := math.Pow(constants.PlayerDamping, delta/(1.0/60.0))
	player.Velocity.X *= damping
	player.Velocity.Y *= damping
	player.Velocity = player.Velocity.limitLength(constants.PlayerMaxSpeed)

	player.X += player.Velocity.X * delta
	player.Y += player.Velocity.Y * delta
}

func (player *Player) nextDelta() float64 {
	now := time.Now()
	if player.LastTick.IsZero() {
		player.LastTick = now
		return 1.0 / 60.0
	}

	delta := now.Sub(player.LastTick).Seconds()
	player.LastTick = now

	return min(delta, 0.05)
}

func axis(negative bool, positive bool) float64 {
	var value float64
	if negative {
		value -= 1
	}
	if positive {
		value += 1
	}

	return max(-1, min(value, 1))
}

func (vector Vector2) limitLength(maxLength float64) Vector2 {
	length := math.Hypot(vector.X, vector.Y)
	if length <= maxLength {
		return vector
	}

	scale := maxLength / length
	return Vector2{
		X: vector.X * scale,
		Y: vector.Y * scale,
	}
}
