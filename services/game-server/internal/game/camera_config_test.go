package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func TestAddPlayerStartsWithBaseCameraView(t *testing.T) {
	gameInstance := NewWithSeed(3)
	playerID := gameInstance.AddPlayer()

	gameInstance.mu.Lock()
	cameraView := gameInstance.cameraViews[playerID]
	sessionConfig := gameInstance.playerSessions[playerID].Config
	shipConfig := gameInstance.entities.Players[playerID].Config
	gameInstance.mu.Unlock()

	assertBaseCameraConfig(t, cameraView.Config)
	assertBaseCameraConfig(t, sessionConfig)
	assertBaseCameraConfig(t, shipConfig)
}

func TestPlayerCameraConfigIsClampedToSupportedZoomRange(t *testing.T) {
	gameInstance := NewWithSeed(5)
	playerID := gameInstance.AddPlayer()
	gameInstance.HandlePacket(playerID, ClientPacket{
		Type: PacketTypeClientConfig,
		Config: runtime.ClientConfig{
			VisibleWorldWidth:  9999,
			VisibleWorldHeight: 9999,
		},
	})

	gameInstance.mu.Lock()
	cameraConfig := gameInstance.cameraViews[playerID].Config
	sessionConfig := gameInstance.playerSessions[playerID].Config
	shipConfig := gameInstance.entities.Players[playerID].Config
	gameInstance.mu.Unlock()

	for name, config := range map[string]runtime.ClientConfig{
		"camera":  cameraConfig,
		"session": sessionConfig,
		"ship":    shipConfig,
	} {
		if config.VisibleWorldWidth != runtime.MaxVisibleWorldWidth || config.VisibleWorldHeight != runtime.MaxVisibleWorldHeight {
			t.Fatalf("expected %s config clamped to %vx%v, got %+v", name, runtime.MaxVisibleWorldWidth, runtime.MaxVisibleWorldHeight, config)
		}
	}
}

func TestBotCameraConfigRemainsLocked(t *testing.T) {
	gameInstance := NewWithSeed(7)
	botID := gameInstance.AddBot()
	gameInstance.HandlePacket(botID, ClientPacket{
		Type: PacketTypeClientConfig,
		Config: runtime.ClientConfig{
			VisibleWorldWidth:  runtime.MaxVisibleWorldWidth,
			VisibleWorldHeight: runtime.MaxVisibleWorldHeight,
		},
	})

	gameInstance.mu.Lock()
	cameraConfig := gameInstance.cameraViews[botID].Config
	sessionConfig := gameInstance.playerSessions[botID].Config
	shipConfig := gameInstance.entities.Players[botID].Config
	gameInstance.mu.Unlock()

	assertBaseCameraConfig(t, cameraConfig)
	assertBaseCameraConfig(t, sessionConfig)
	assertBaseCameraConfig(t, shipConfig)
}

func assertBaseCameraConfig(t *testing.T, config runtime.ClientConfig) {
	t.Helper()
	if config.VisibleWorldWidth != runtime.BaseVisibleWorldWidth || config.VisibleWorldHeight != runtime.BaseVisibleWorldHeight {
		t.Fatalf("expected base camera config %vx%v, got %+v", runtime.BaseVisibleWorldWidth, runtime.BaseVisibleWorldHeight, config)
	}
}
