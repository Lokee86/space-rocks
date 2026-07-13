package game

import (
	"fmt"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spatial/grid"
)

func BenchmarkProjectileAsteroidCollisionBroadPhase(b *testing.B) {
	for _, size := range []struct {
		name       string
		asteroids   int
		projectiles int
	}{
		{name: "100x500", asteroids: 100, projectiles: 500},
		{name: "500x2000", asteroids: 500, projectiles: 2000},
	} {
		b.Run(size.name+"/brute_force", func(b *testing.B) {
			fixture := newCollisionBenchmarkFixture(size.asteroids, size.projectiles)
			b.ResetTimer()
			checks := 0
			for i := 0; i < b.N; i++ {
				for _, projectile := range fixture.projectiles {
					for _, asteroid := range fixture.asteroids {
						checks++
						detectProjectileAsteroidCollision(projectile, asteroid, fixture.game.collisionShapes)
					}
				}
			}
			b.StopTimer()
			b.ReportMetric(float64(checks)/float64(b.N), "narrow_checks/op")
			b.ReportMetric(0, "candidate_reduction_pct")
		})

		b.Run(size.name+"/spatial_candidates", func(b *testing.B) {
			fixture := newCollisionBenchmarkFixture(size.asteroids, size.projectiles)
			b.ResetTimer()
			checks := 0
			for i := 0; i < b.N; i++ {
				fixture.game.rebuildAsteroidSpatialIndex()
				for _, projectile := range fixture.projectiles {
					body, ok := projectile.CollisionBody(fixture.game.collisionShapes)
					if !ok {
						continue
					}
					for _, ref := range fixture.game.asteroidCollisionCandidates(body) {
						checks++
						detectProjectileAsteroidCollision(projectile, fixture.game.entities.Asteroids[ref.ID], fixture.game.collisionShapes)
					}
				}
			}
			b.StopTimer()
			b.ReportMetric(float64(checks)/float64(b.N), "narrow_checks/op")
			bruteChecks := float64(size.asteroids * size.projectiles)
			candidateChecks := float64(checks) / float64(b.N)
			b.ReportMetric((1-candidateChecks/bruteChecks)*100, "candidate_reduction_pct")
		})
	}
}

type collisionBenchmarkFixture struct {
	game        *Game
	asteroids   []*runtime.Asteroid
	projectiles []*runtime.Bullet
}

func newCollisionBenchmarkFixture(asteroidCount, projectileCount int) collisionBenchmarkFixture {
	game := &Game{
		collisionShapes: physics.CollisionShapeCatalog{
			Bullet:    physics.ImportedCollisionShape{Type: "circle", Radius: 3},
			Asteroids: []physics.ImportedCollisionShape{{Type: "circle", Radius: 12}},
		},
		entities:     runtime.NewEntityStore(),
		spatialIndex: grid.New(space.DefaultBounds(), defaultSpatialCellSize),
	}
	bounds := space.DefaultBounds()
	fixture := collisionBenchmarkFixture{
		game:        game,
		asteroids:   make([]*runtime.Asteroid, 0, asteroidCount),
		projectiles: make([]*runtime.Bullet, 0, projectileCount),
	}
	for i := 0; i < asteroidCount; i++ {
		position := deterministicBenchmarkPosition(i, asteroidCount, bounds)
		asteroid := &runtime.Asteroid{ID: fmt.Sprintf("asteroid-%d", i), X: position.X, Y: position.Y, Size: 1}
		fixture.asteroids = append(fixture.asteroids, asteroid)
		game.entities.Asteroids[asteroid.ID] = asteroid
	}
	for i := 0; i < projectileCount; i++ {
		position := deterministicBenchmarkPosition(i*7+3, projectileCount, bounds)
		projectile := &runtime.Bullet{ID: fmt.Sprintf("projectile-%d", i), X: position.X, Y: position.Y}
		fixture.projectiles = append(fixture.projectiles, projectile)
		game.entities.Projectiles[projectile.ID] = projectile
	}
	return fixture
}

func deterministicBenchmarkPosition(index, count int, bounds space.Bounds) physics.Vector2 {
	return physics.Vector2{
		X: float64((index*7919)%count) / float64(count) * bounds.Width,
		Y: float64((index*104729)%count) / float64(count) * bounds.Height,
	}
}


