package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestDevtoolsBulletsCanMoveOnNewGameReturnsTrue(t *testing.T) {
	game := New()

	if got := game.DevtoolsBulletsCanMove(); !got {
		t.Fatalf("expected bullets to be movable on a new game, got %v", got)
	}
}

func TestDevtoolsSpawnBulletWithValidOwnerPlayerID(t *testing.T) {
	game := New()
	ownerID := game.AddPlayer()
	origin := physics.Vector2{X: 120, Y: 220}
	direction := physics.Vector2{X: 0, Y: -1}

	bullet, ok := game.DevtoolsSpawnBullet(ownerID, origin, direction)
	if !ok {
		t.Fatal("expected DevtoolsSpawnBullet to succeed")
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
