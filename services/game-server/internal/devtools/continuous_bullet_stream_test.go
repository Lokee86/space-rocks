package devtools

import (
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
)

func TestHandleDebugBeginContinuousBulletStreamRegistersObserverAndSpawnsOnStep(t *testing.T) {
	gameInstance := game.New()
	control := game.NewControl(gameInstance)
	controller := NewController(Dependencies{Target: control})
	playerID := gameInstance.AddPlayer()

	command := DebugCommand{
		Type:         PacketTypeDebugBeginContinuousBulletStream,
		HasDirection: true,
		X:            10,
		Y:            20,
		DirectionX:   0,
		DirectionY:   -1,
	}

	if ok := controller.HandleCommand(playerID, command); !ok {
		t.Fatal("expected HandleCommand to return true")
	}

	beforeSnapshot := gameInstance.GameplayPresentationSnapshot(playerID)
	if len(beforeSnapshot.Bullets) != 0 {
		t.Fatalf("expected 0 bullets before stepping the game, got %d", len(beforeSnapshot.Bullets))
	}

	done := make(chan struct{})
	go func() {
		gameInstance.Step(constants.BasicCannonCooldown)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("continuous bullet stream simulation step deadlocked")
	}

	afterSnapshot := gameInstance.GameplayPresentationSnapshot(playerID)
	if len(afterSnapshot.Bullets) != 1 {
		t.Fatalf("expected 1 bullet after stepping the game, got %d", len(afterSnapshot.Bullets))
	}
}
