package grid

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spatial"
)

const equivalenceSeed int64 = 1407

func TestCircleBroadPhaseHasNoExactCollisionFalseNegatives(t *testing.T) {
	const (
		iterations = 80
		targets    = 36
		queries    = 24
	)
	bounds := space.Bounds{Width: 37, Height: 29}
	index := New(bounds, 6)
	random := rand.New(rand.NewSource(equivalenceSeed))

	for iteration := 0; iteration < iterations; iteration++ {
		entries := make([]spatial.Entry, 0, targets)
		bodies := make([]physics.CollisionBody, 0, targets)
		for target := 0; target < targets; target++ {
			body := randomEquivalenceBody(random, target, bounds)
			bodies = append(bodies, body)
			entries = append(entries, spatial.Entry{
				Ref:      spatial.Ref{Kind: spatial.KindAsteroid, ID: fmt.Sprintf("target-%02d", target)},
				Position: body.Position,
				Radius:   physics.BoundingRadius(body.Shape),
			})
		}
		index.Rebuild(entries)

		for query := 0; query < queries; query++ {
			queryBody := randomEquivalenceBody(random, query, bounds)
			candidateRefs := index.QueryCircle(nil, queryBody.Position, physics.BoundingRadius(queryBody.Shape), spatial.AllKinds)
			candidates := make(map[string]bool, len(candidateRefs))
			for _, ref := range candidateRefs {
				candidates[ref.ID] = true
			}

			for target, targetBody := range bodies {
				delta := space.ShortestDelta(queryBody.Position, targetBody.Position, bounds)
				localTarget := targetBody
				localTarget.Position = queryBody.Position.Add(delta)
				if _, collided := physics.DetectCollision(queryBody, localTarget); !collided {
					continue
				}
				id := fmt.Sprintf("target-%02d", target)
				if !candidates[id] {
					t.Fatalf("seed=%d iteration=%d query=%d target=%s omitted exact collision; query=%#v target=%#v", equivalenceSeed, iteration, query, id, queryBody, targetBody)
				}
			}
		}
	}
}

func TestCircleBroadPhaseCornerWrapExplicit(t *testing.T) {
	bounds := space.Bounds{Width: 37, Height: 29}
	index := New(bounds, 6)
	target := physics.CollisionBody{Position: physics.Vector2{X: 36.2, Y: 28.2}, Shape: physics.NewCircleShape(1.5)}
	query := physics.CollisionBody{Position: physics.Vector2{X: 0.3, Y: 0.3}, Shape: physics.NewCircleShape(1.5)}
	index.Rebuild([]spatial.Entry{{Ref: spatial.Ref{Kind: spatial.KindAsteroid, ID: "corner"}, Position: target.Position, Radius: physics.BoundingRadius(target.Shape)}})

	refs := index.QueryCircle(nil, query.Position, physics.BoundingRadius(query.Shape), spatial.AllKinds)
	if len(refs) != 1 || refs[0].ID != "corner" {
		t.Fatalf("corner-wrap refs = %#v", refs)
	}
	delta := space.ShortestDelta(query.Position, target.Position, bounds)
	target.Position = query.Position.Add(delta)
	if _, ok := physics.DetectCollision(query, target); !ok {
		t.Fatal("explicit corner-wrap narrow collision was not detected")
	}
}

func randomEquivalenceBody(random *rand.Rand, id int, bounds space.Bounds) physics.CollisionBody {
	position := randomEquivalencePosition(random, id, bounds)
	rotation := random.Float64() * math.Pi * 2
	switch random.Intn(4) {
	case 0:
		return physics.CollisionBody{
			ID:       fmt.Sprintf("body-%d", id),
			Position: position,
			Rotation: rotation,
			Shape:    physics.NewCircleShape(0.2 + random.Float64()*3),
		}
	case 1:
		return physics.CollisionBody{
			ID:       fmt.Sprintf("body-%d", id),
			Position: position,
			Rotation: rotation,
			Shape:    physics.NewCapsuleShape(0.2+random.Float64()*1.5, 1+random.Float64()*6),
		}
	case 2:
		return physics.CollisionBody{
			ID:       fmt.Sprintf("body-%d", id),
			Position: position,
			Rotation: rotation,
			Shape:    physics.NewRectangleShape(0.5+random.Float64()*5, 0.5+random.Float64()*5),
		}
	default:
		return physics.CollisionBody{
			ID:       fmt.Sprintf("body-%d", id),
			Position: position,
			Rotation: rotation,
			Shape: physics.NewPolygonShape([]physics.Vector2{
				{X: -1, Y: -1},
				{X: 1.5, Y: -0.8},
				{X: 1.2, Y: 1.1},
				{X: -0.8, Y: 1.4},
			}),
		}
	}
}

func randomEquivalencePosition(random *rand.Rand, id int, bounds space.Bounds) physics.Vector2 {
	switch id % 8 {
	case 0:
		return physics.Vector2{X: random.Float64() * 1.5, Y: random.Float64() * bounds.Height}
	case 1:
		return physics.Vector2{X: bounds.Width - random.Float64()*1.5, Y: random.Float64() * bounds.Height}
	case 2:
		return physics.Vector2{X: random.Float64() * bounds.Width, Y: random.Float64() * 1.5}
	case 3:
		return physics.Vector2{X: random.Float64() * bounds.Width, Y: bounds.Height - random.Float64()*1.5}
	default:
		return physics.Vector2{X: random.Float64() * bounds.Width, Y: random.Float64() * bounds.Height}
	}
}
