package devtools

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

type collisionTelemetryTargetFake struct {
	bodies map[string][]physics.CollisionBody
}

func (target *collisionTelemetryTargetFake) CollisionBodiesByKind() map[string][]physics.CollisionBody {
	return target.bodies
}

func TestCollisionBodiesUsesDeterministicKindOrdering(t *testing.T) {
	target := &collisionTelemetryTargetFake{bodies: map[string][]physics.CollisionBody{
		"pickup": {{ID: "pickup-1", Shape: physics.CollisionShape{Type: physics.CollisionShapeCircle, Radius: 3}}},
		"bullet": {{ID: "bullet-1", Shape: physics.CollisionShape{Type: physics.CollisionShapeCircle, Radius: 2}}},
		"player": {{ID: "player-1", Shape: physics.CollisionShape{Type: physics.CollisionShapeRectangle, Size: physics.Vector2{X: 4, Y: 2}}}},
		"asteroid": {{ID: "asteroid-1", Shape: physics.CollisionShape{Type: physics.CollisionShapeCircle, Radius: 8}}},
	}}

	bodies := CollisionBodies(target)
	kinds := make([]string, 0, len(bodies))
	for _, body := range bodies {
		kinds = append(kinds, body.Kind)
	}

	want := []string{"player", "asteroid", "bullet", "pickup"}
	if !reflect.DeepEqual(kinds, want) {
		t.Fatalf("CollisionBodies() kinds = %v, want %v", kinds, want)
	}
}

func TestCollisionBodiesProjectsOutlinePoints(t *testing.T) {
	target := &collisionTelemetryTargetFake{bodies: map[string][]physics.CollisionBody{
		"player": {{ID: "player-1", Shape: physics.CollisionShape{Type: physics.CollisionShapeRectangle, Size: physics.Vector2{X: 4, Y: 2}}}},
	}}

	bodies := CollisionBodies(target)
	if len(bodies) != 1 {
		t.Fatalf("expected 1 body, got %d", len(bodies))
	}
	if len(bodies[0].Points) != 4 {
		t.Fatalf("expected 4 outline points, got %d", len(bodies[0].Points))
	}
	if math.Abs(bodies[0].Points[0].X+2) > 1e-9 || math.Abs(bodies[0].Points[0].Y+1) > 1e-9 {
		t.Fatalf("expected projected outline point near (-2,-1), got %#v", bodies[0].Points[0])
	}
}

func TestCollisionBodyMarshalsWithLowercaseJSONFields(t *testing.T) {
	data, err := json.Marshal(CollisionBody{Kind: "player", ID: "Player-1", Shape: "rectangle", Points: []CollisionPoint{{X: 1, Y: 2}}})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	jsonText := string(data)
	for _, needle := range []string{`"kind"`, `"id"`, `"shape"`, `"points"`, `"x"`, `"y"`} {
		if !strings.Contains(jsonText, needle) {
			t.Fatalf("expected %q to contain %q", jsonText, needle)
		}
	}
}

func TestCollisionBodiesNilTargetReturnsNil(t *testing.T) {
	if got := CollisionBodies(nil); got != nil {
		t.Fatalf("expected nil result, got %v", got)
	}
}
