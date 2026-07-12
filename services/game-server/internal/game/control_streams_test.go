package game

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
)

func TestDevtoolsBulletsCanMoveOnNewGameReturnsTrue(t *testing.T) {
	gameInstance := New()
	control := NewControl(gameInstance)

	if got := control.BulletsCanMove(); !got {
		t.Fatalf("expected bullets to be movable on a new game, got %v", got)
	}
}

func TestDevtoolsSpawnBulletWithValidOwnerPlayerID(t *testing.T) {
	gameInstance := New()
	control := NewControl(gameInstance)
	ownerID := gameInstance.AddPlayer()
	origin := physics.Vector2{X: 120, Y: 220}
	direction := physics.Vector2{X: 0, Y: -1}

	bullet, ok := control.SpawnBullet(ownerID, origin, direction)
	if !ok {
		t.Fatal("expected SpawnBullet to succeed")
	}
	if bullet == nil {
		t.Fatal("expected spawned bullet to be non-nil")
	}
	if bullet.OwnerID != ownerID {
		t.Fatalf("expected owner %q, got %q", ownerID, bullet.OwnerID)
	}
	if bullet.Position() != origin {
		t.Fatalf("expected origin %+v, got %+v", origin, bullet.Position())
	}
}

func TestControlPublicMethodsAreSafeDuringConcurrentStepping(t *testing.T) {
	gameInstance := New()
	control := NewControl(gameInstance)
	playerID := gameInstance.AddPlayer()
	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				control.WorldFrozen()
				control.BulletsCanMove()
				control.SetPlayerScore(playerID, 1)
				control.AddPlayerLives(playerID, 1)
				control.TargetPlayerIDs()
			}
		}()
	}
	for range 100 {
		gameInstance.Step(1.0 / 60.0)
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("concurrent Control calls did not complete")
	}
}

func TestControlDoesNotExposeLockAssumingObserverMethods(t *testing.T) {
	typeOfControl := reflect.TypeOf(&Control{})
	for _, name := range []string{"BulletsCanMoveLocked", "SpawnDebugBulletLocked"} {
		if _, ok := typeOfControl.MethodByName(name); ok {
			t.Fatalf("Control must not expose %s", name)
		}
	}
}
