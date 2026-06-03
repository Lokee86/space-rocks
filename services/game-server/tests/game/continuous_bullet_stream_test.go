package gametests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestDevtoolsBeginContinuousBulletStreamTracksActiveStream(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	origin := physics.Vector2{X: 10, Y: 20}
	direction := physics.Vector2{X: 2, Y: 0}

	if !scenario.game.DevtoolsBeginContinuousBulletStream(playerID, origin, direction) {
		t.Fatal("expected continuous bullet stream to start")
	}

	streams := scenario.game.DevtoolsActiveContinuousBulletStreams()
	if len(streams) != 1 {
		t.Fatalf("expected 1 active continuous bullet stream, got %d", len(streams))
	}
	if streams[0].OwnerPlayerID != playerID {
		t.Fatalf("expected owner %q, got %q", playerID, streams[0].OwnerPlayerID)
	}
	if streams[0].Origin != origin {
		t.Fatalf("expected origin %+v, got %+v", origin, streams[0].Origin)
	}
	if streams[0].Direction != (physics.Vector2{X: 1, Y: 0}) {
		t.Fatalf("expected normalized direction, got %+v", streams[0].Direction)
	}
	if streams[0].CooldownRemaining != constants.BulletCooldown {
		t.Fatalf("expected cooldown %f, got %f", constants.BulletCooldown, streams[0].CooldownRemaining)
	}
}

func TestContinuousBulletStreamStepSpawnsBulletAfterCooldown(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	origin := physics.Vector2{X: 10, Y: 20}
	direction := physics.Vector2{X: 0, Y: -1}

	if !scenario.game.DevtoolsBeginContinuousBulletStream(playerID, origin, direction) {
		t.Fatal("expected continuous bullet stream to start")
	}
	scenario.step(constants.BulletCooldown)

	if bulletCount := scenario.bullets().Len(); bulletCount != 1 {
		t.Fatalf("expected 1 spawned bullet, got %d", bulletCount)
	}

	bullet := scenario.bullet("bullet-1")
	if bullet.OwnerID != playerID {
		t.Fatalf("expected bullet owner %q, got %q", playerID, bullet.OwnerID)
	}
}

func TestDevtoolsClearContinuousBulletStreamsStopsFutureBullets(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	if !scenario.game.DevtoolsBeginContinuousBulletStream(playerID, physics.Vector2{X: 10, Y: 20}, physics.Vector2{X: 0, Y: -1}) {
		t.Fatal("expected continuous bullet stream to start")
	}
	scenario.game.DevtoolsClearContinuousBulletStreams()
	scenario.step(constants.BulletCooldown)

	if streams := scenario.game.DevtoolsActiveContinuousBulletStreams(); len(streams) != 0 {
		t.Fatalf("expected no active continuous bullet streams, got %d", len(streams))
	}
	if bulletCount := scenario.bullets().Len(); bulletCount != 0 {
		t.Fatalf("expected no spawned bullets after clearing streams, got %d", bulletCount)
	}
}
