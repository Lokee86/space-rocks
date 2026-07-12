package game

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	runtimepkg "github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

type presentationBenchmarkScenario struct {
	name      string
	players   int
	asteroids int
	bullets   int
}

var presentationBenchmarkScenarios = []presentationBenchmarkScenario{
	{name: "1p-100a-100b", players: 1, asteroids: 100, bullets: 100},
	{name: "8p-100a-500b", players: 8, asteroids: 100, bullets: 500},
	{name: "stress-16p-500a-2000b", players: 16, asteroids: 500, bullets: 2000},
}

func BenchmarkGameplayPresentationSnapshot(b *testing.B) {
	for _, scenario := range presentationBenchmarkScenarios {
		b.Run(scenario.name, func(b *testing.B) {
			game := newPresentationBenchmarkGame(scenario)
			b.ReportAllocs()
			b.ReportMetric(float64(scenario.players+scenario.asteroids+scenario.bullets), "entities/snapshot")
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = game.GameplayPresentationSnapshot("player-0")
			}
		})
	}
}

func BenchmarkGameplayPresentationFramePublication(b *testing.B) {
	for _, scenario := range presentationBenchmarkScenarios {
		b.Run(scenario.name, func(b *testing.B) {
			game := newPresentationBenchmarkGame(scenario)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				game.mu.Lock()
				game.publishPresentationFrameLocked()
				game.mu.Unlock()
			}
		})
	}
}

func BenchmarkGameStepWithPresentationSnapshotContention(b *testing.B) {
	for _, scenario := range presentationBenchmarkScenarios {
		for _, readers := range []int{0, 1, 4, 8} {
			b.Run(fmt.Sprintf("%s/%d-readers", scenario.name, readers), func(b *testing.B) {
				game := newPresentationBenchmarkGame(scenario)
				game.worldSimulationOptions.SetFreezeWorld(true)
				var stop atomic.Bool
				var snapshots atomic.Uint64
				var wg sync.WaitGroup
				wg.Add(readers)
				for i := 0; i < readers; i++ {
					go func() {
						defer wg.Done()
						for !stop.Load() {
							_ = game.GameplayPresentationSnapshot("player-0")
							snapshots.Add(1)
						}
					}()
				}

				b.ReportAllocs()
				b.ReportMetric(float64(scenario.players+scenario.asteroids+scenario.bullets), "entities/step")
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					game.Step(1.0 / 60.0)
				}
				b.StopTimer()
				stop.Store(true)
				wg.Wait()
				b.ReportMetric(float64(snapshots.Load())/float64(b.N), "snapshots/step")
			})
		}
	}
}

func newPresentationBenchmarkGame(scenario presentationBenchmarkScenario) *Game {
	game := New()
	for i := 0; i < scenario.players; i++ {
		id := fmt.Sprintf("player-%d", i)
		session := newPlayerSession(id, physics.Vector2{X: float64(i * 32), Y: float64(i * 32)})
		game.playerSessions[id] = session
		game.entities.Players[id] = session.NewShip(session.SpawnPosition)
	}
	for i := 0; i < scenario.asteroids; i++ {
		id := fmt.Sprintf("asteroid-%d", i)
		game.entities.Asteroids[id] = runtimepkg.NewAsteroid(id, physics.Vector2{X: float64(i), Y: float64(i * 2)}, physics.Vector2{X: 1, Y: -1}, 2, i%3)
	}
	for i := 0; i < scenario.bullets; i++ {
		id := fmt.Sprintf("bullet-%d", i)
		game.entities.Projectiles[id] = runtimepkg.NewBullet(id, "player-0", physics.Vector2{X: float64(i), Y: float64(i * 2)}, 0, physics.Vector2{X: 2, Y: 0}, 10)
	}
	game.mu.Lock()
	game.publishPresentationFrameLocked()
	game.mu.Unlock()
	return game
}
