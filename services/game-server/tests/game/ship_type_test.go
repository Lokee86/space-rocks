package gametests

import (
	"encoding/json"
	"testing"

	servergame "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
)

func TestNewShipsDefaultToDefaultShipTypeID(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	ship := scenario.player(playerID)
	if ship.ShipTypeID != entities.DefaultShipTypeID {
		t.Fatalf("expected new ship type %q, got %q", entities.DefaultShipTypeID, ship.ShipTypeID)
	}
}

func TestNewPlayerSessionDefaultsToDefaultShipTypeID(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	shipTypeID := scenario.sessionField(playerID, "ShipTypeID").String()
	if shipTypeID != entities.DefaultShipTypeID {
		t.Fatalf("expected new session ship type %q, got %q", entities.DefaultShipTypeID, shipTypeID)
	}
}

func TestSessionCreatedShipsCopySessionShipTypeID(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	const testShipTypeID = "test_ship"
	scenario.sessionField(playerID, "ShipTypeID").SetString(testShipTypeID)
	scenario.removePlayerEntity(playerID)

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeRespawn})

	ship := scenario.player(playerID)
	if ship.ShipTypeID != testShipTypeID {
		t.Fatalf("expected respawned ship type %q, got %q", testShipTypeID, ship.ShipTypeID)
	}
}

func TestShipStateIncludesShipType(t *testing.T) {
	ship := entities.Ship{
		ID:         "player-1",
		ShipTypeID: "test_ship",
	}

	rawState, err := json.Marshal(ship.State())
	if err != nil {
		t.Fatalf("marshal ship state: %v", err)
	}

	var state map[string]any
	if err := json.Unmarshal(rawState, &state); err != nil {
		t.Fatalf("decode ship state: %v", err)
	}
	if state["ship_type"] != ship.ShipTypeID {
		t.Fatalf("expected ship_type %q, got %v", ship.ShipTypeID, state["ship_type"])
	}
}

func TestShipStateShipTypeEqualsShipTypeID(t *testing.T) {
	ship := entities.Ship{
		ID:         "player-1",
		ShipTypeID: "test_ship",
	}

	state := ship.State()
	if state.ShipType != ship.ShipTypeID {
		t.Fatalf("expected state ship type %q, got %q", ship.ShipTypeID, state.ShipType)
	}
}
