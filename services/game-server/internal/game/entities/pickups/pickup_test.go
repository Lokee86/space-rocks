package pickups

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestPickupCollisionBodyUsesPickupClassShapeKeys(t *testing.T) {
	catalog := physics.CollisionShapeCatalog{
		Pickups: map[string]physics.ImportedCollisionShape{
			"powerup": {
				Name:   "CollisionShape2D",
				Type:   string(physics.CollisionShapeCircle),
				Radius: 50,
			},
			"weapon": {
				Name:   "CollisionShape2D",
				Type:   string(physics.CollisionShapeCircle),
				Radius: 30,
			},
		},
	}

	tests := []struct {
		name          string
		pickup        Pickup
		wantRadius    float64
	}{
		{
			name: "one up uses powerup class shape",
			pickup: Pickup{
				ID:   "pickup-1",
				Type: TypeOneUp,
			},
			wantRadius:   50,
		},
		{
			name: "torpedo uses weapon class shape",
			pickup: Pickup{
				ID:   "pickup-2",
				Type: TypeTorpedo,
			},
			wantRadius:   30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, ok := tt.pickup.CollisionBody(catalog)
			if !ok {
				t.Fatalf("CollisionBody() ok = false, want true")
			}
			if body.ID != tt.pickup.ID {
				t.Fatalf("CollisionBody().ID = %q, want %q", body.ID, tt.pickup.ID)
			}
			if body.Shape.Type != physics.CollisionShapeCircle {
				t.Fatalf("CollisionBody().Shape.Type = %q, want %q", body.Shape.Type, physics.CollisionShapeCircle)
			}
			if body.Shape.Radius != tt.wantRadius {
				t.Fatalf("CollisionBody().Shape.Radius = %v, want %v", body.Shape.Radius, tt.wantRadius)
			}
		})
	}
}
