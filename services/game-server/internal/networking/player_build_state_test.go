package networking

import (
	"testing"

	"github.com/Lokee86/space-rocks/player-data/playerdata"
	"github.com/Lokee86/space-rocks/player-data/protocol"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/modes"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/playerbuild"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

type testInventoryLoader struct {
	inventory protocol.HangarInventory
}

func (loader testInventoryLoader) Load(protocol.PlayerDataIdentity, protocol.PlayerDataRequestContext) (protocol.PlayerDataLoadHangarInventoryResult, error) {
	return protocol.PlayerDataLoadHangarInventoryResult{Found: true, Inventory: loader.inventory}, nil
}

func TestLoadPlayerBuildOptionsResolvesFallbackSelection(t *testing.T) {
	inventory := protocol.HangarInventory{
		InventoryVersion: 3,
		OwnedShips: []protocol.OwnedShip{{
			OwnedShipID: "owned-v-wing", ShipID: playerbuild.ShipVWing, State: "normal",
		}},
		OwnedWeapons: []protocol.OwnedWeapon{{
			OwnedWeaponID: "owned-pulse", WeaponID: playerbuild.WeaponPulse, State: "normal",
		}},
		DefaultOwnedShipID: "owned-v-wing",
	}
	service, err := playerbuild.NewService(testInventoryLoader{inventory: inventory}, playerbuild.DefaultCatalog())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	session := newWebSocketSession(nil, nil, nil, nil, service)

	if err := session.loadPlayerBuildOptions("pilot-1", playerdata.PlayModeSinglePlayer, string(modes.ModeArcadeSurvival), "trace-1"); err != nil {
		t.Fatalf("load options: %v", err)
	}

	options, selection := session.playerBuildPacketStates()
	if len(options.EligibleShips) != 1 || options.EligibleShips[0].OwnedShipID != "owned-v-wing" {
		t.Fatalf("unexpected ship options: %#v", options.EligibleShips)
	}
	if len(options.EligibleWeapons) != 1 || options.EligibleWeapons[0].WeaponPoint != string(playerbuild.Primary1) {
		t.Fatalf("unexpected weapon options: %#v", options.EligibleWeapons)
	}
	if !selection.Valid || selection.SelectedOwnedShipID != "owned-v-wing" {
		t.Fatalf("unexpected fallback selection: %#v", selection)
	}
	if got := selection.SelectedWeaponsByPoint[string(playerbuild.Primary1)]; got != "owned-pulse" {
		t.Fatalf("expected primary fallback weapon, got %q", got)
	}
}

func TestLoadoutSubmissionIsLockedAfterMatchStart(t *testing.T) {
	inventory := protocol.HangarInventory{
		OwnedShips:         []protocol.OwnedShip{{OwnedShipID: "owned-v-wing", ShipID: playerbuild.ShipVWing, State: "normal"}},
		OwnedWeapons:       []protocol.OwnedWeapon{{OwnedWeaponID: "owned-pulse", WeaponID: playerbuild.WeaponPulse, State: "normal"}},
		DefaultOwnedShipID: "owned-v-wing",
	}
	service, err := playerbuild.NewService(testInventoryLoader{inventory: inventory}, playerbuild.DefaultCatalog())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	session := newWebSocketSession(nil, nil, nil, nil, service)
	if err := session.loadPlayerBuildOptions("pilot-1", playerdata.PlayModeSinglePlayer, string(modes.ModeArcadeSurvival), "trace-1"); err != nil {
		t.Fatalf("load options: %v", err)
	}
	session.bindRoom(rooms.NewRoom("active-room", rooms.RoomStateInGame, nil))

	session.handleSetLoadoutRequest("trace-2", "owned-v-wing", map[string]string{string(playerbuild.Primary1): "owned-pulse"}, map[string]string{}, map[string]int{})
	_, selection := session.playerBuildPacketStates()
	if selection.ErrorCode != loadoutErrorLocked {
		t.Fatalf("expected locked loadout error, got %#v", selection)
	}
}

func TestInvalidLoadoutKeepsLastResolvedBuild(t *testing.T) {
	inventory := protocol.HangarInventory{
		OwnedShips:         []protocol.OwnedShip{{OwnedShipID: "owned-v-wing", ShipID: playerbuild.ShipVWing, State: "normal"}},
		OwnedWeapons:       []protocol.OwnedWeapon{{OwnedWeaponID: "owned-pulse", WeaponID: playerbuild.WeaponPulse, State: "normal"}},
		DefaultOwnedShipID: "owned-v-wing",
	}
	service, err := playerbuild.NewService(testInventoryLoader{inventory: inventory}, playerbuild.DefaultCatalog())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	session := newWebSocketSession(nil, nil, nil, nil, service)
	if err := session.loadPlayerBuildOptions("pilot-1", playerdata.PlayModeSinglePlayer, string(modes.ModeArcadeSurvival), "trace-1"); err != nil {
		t.Fatalf("load options: %v", err)
	}
	before := session.resolvedBuildTemplate()

	session.handleSetLoadoutRequest("trace-2", "missing-ship", map[string]string{}, map[string]string{}, map[string]int{})
	after := session.resolvedBuildTemplate()
	_, selection := session.playerBuildPacketStates()

	if after.SelectedOwnedShipID != before.SelectedOwnedShipID {
		t.Fatalf("invalid submission replaced accepted build: before=%q after=%q", before.SelectedOwnedShipID, after.SelectedOwnedShipID)
	}
	if selection.ErrorCode != loadoutErrorInvalid {
		t.Fatalf("expected invalid loadout error, got %#v", selection)
	}
}
