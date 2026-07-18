package game

import (
	"sync"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/measurement"
)

func TestRuntimeMeasurementIsDormantUntilAttached(t *testing.T) {
	game := NewWithSeed(1)
	called := 0
	game.Step(1.0 / 60.0)
	if called != 0 || game.HasRuntimeMeasurements() {
		t.Fatalf("runtime measurement should be dormant by default")
	}

	detach := game.AttachRuntimeMeasurement(measurement.SimulationObserverFunc(func(time.Duration, measurement.EntityCounts) {
		called++
	}))
	game.Step(1.0 / 60.0)
	detach()
	game.Step(1.0 / 60.0)
	if called != 1 {
		t.Fatalf("expected one attached-step observation, got %d", called)
	}
}

func TestRuntimeMeasurementReportsCountsAfterMatchOverStep(t *testing.T) {
	game := newMatchOverTestGame()
	game.entities.Players["player-1"] = &runtime.Ship{ID: "player-1"}
	game.entities.Enemies["enemy-1"] = &runtime.Ship{ID: "enemy-1"}
	game.entities.Asteroids["asteroid-1"] = &runtime.Asteroid{ID: "asteroid-1"}
	game.entities.Projectiles["bullet-1"] = &runtime.Bullet{ID: "bullet-1"}
	for i := 0; i < 9; i++ {
		game.spawner.NextAsteroidID(game.entities.Asteroids)
	}

	var observed measurement.EntityCounts
	detach := game.AttachRuntimeMeasurement(measurement.SimulationObserverFunc(func(_ time.Duration, counts measurement.EntityCounts) {
		observed = counts
		game.mu.Lock()
		game.mu.Unlock()
	}))
	game.Step(1.0 / 60.0)
	detach()

	if observed.Players != 1 || observed.PlayerSessions != 1 || observed.Enemies != 1 || observed.AsteroidsSpawnedTotal != 10 {
		t.Fatalf("unexpected match-over entity observation: %#v", observed)
	}
}

func TestRuntimeMeasurementRunsRemainIndependent(t *testing.T) {
	game := NewWithSeed(2)
	first := measurement.NewRun(measurement.RunContext{RunID: "first"})
	second := measurement.NewRun(measurement.RunContext{RunID: "second"})
	detachFirst := game.AttachRuntimeMeasurement(first)
	detachSecond := game.AttachRuntimeMeasurement(second)

	game.Step(1.0 / 60.0)
	detachFirst()
	game.Step(1.0 / 60.0)
	detachSecond()

	if first.Finalize().Ticks.Count != 1 || second.Finalize().Ticks.Count != 2 {
		t.Fatalf("detaching one run should not affect another: first=%d second=%d", first.Finalize().Ticks.Count, second.Finalize().Ticks.Count)
	}
}

func TestRuntimeMeasurementCountsUsePlayerSessionsSeparately(t *testing.T) {
	game := NewWithSeed(3)
	game.playerSessions["player-1"] = newPlayerSession("player-1", physics.Vector2{})
	game.entities.Players["player-1"] = &runtime.Ship{ID: "player-1"}
	var observed measurement.EntityCounts
	detach := game.AttachRuntimeMeasurement(measurement.SimulationObserverFunc(func(_ time.Duration, counts measurement.EntityCounts) { observed = counts }))
	game.Step(0)
	detach()
	if observed.Players != 1 || observed.PlayerSessions != 1 {
		t.Fatalf("expected separate player and session counts, got %#v", observed)
	}
}

func TestRuntimeMeasurementAttachDetachAndStepAreRaceSafe(t *testing.T) {
	game := NewWithSeed(4)
	var wait sync.WaitGroup
	wait.Add(2)
	go func() {
		defer wait.Done()
		for i := 0; i < 100; i++ {
			game.Step(1.0 / 60.0)
		}
	}()
	go func() {
		defer wait.Done()
		for i := 0; i < 100; i++ {
			detach := game.AttachRuntimeMeasurement(measurement.SimulationObserverFunc(func(time.Duration, measurement.EntityCounts) {}))
			if i%2 == 0 {
				detach()
			}
		}
	}()
	wait.Wait()
}
