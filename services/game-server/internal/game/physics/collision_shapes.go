package physics

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
)

const (
	DefaultShipCollisionShapeID = "v_wing"
	collisionShapesRelativePath = "shared/collisions/collision_shapes.json"
)

var (
	executablePath   = os.Executable
	workingDirectory = os.Getwd
)

// BoundingRadius returns a conservative radius enclosing the local collision shape.
func BoundingRadius(shape CollisionShape) float64 {
	switch shape.Type {
	case CollisionShapeCircle:
		return shape.Radius
	case CollisionShapeCapsule:
		return math.Max(shape.Height*0.5, shape.Radius)
	case CollisionShapeRectangle:
		return shape.Size.Multiply(0.5).Length()
	case CollisionShapePolygon:
		radius := 0.0
		for _, point := range shape.Points {
			radius = math.Max(radius, point.Length())
		}
		return radius
	default:
		return 0
	}
}

type CollisionShapeCatalog struct {
	Bullet    ImportedCollisionShape            `json:"bullet"`
	Ship      ImportedCollisionShape            `json:"ship"`
	Asteroids []ImportedCollisionShape          `json:"asteroids"`
	Pickups   map[string]ImportedCollisionShape `json:"pickups"`
}

type ImportedCollisionShape struct {
	Name   string      `json:"name"`
	Type   string      `json:"type"`
	Radius float64     `json:"radius"`
	Height float64     `json:"height"`
	Size   []float64   `json:"size"`
	Points [][]float64 `json:"points"`
}

func LoadCollisionShapeCatalog() (CollisionShapeCatalog, error) {
	path, err := findSharedCollisionShapesPath()
	if err != nil {
		return CollisionShapeCatalog{}, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return CollisionShapeCatalog{}, err
	}

	var catalog CollisionShapeCatalog
	if err := json.Unmarshal(content, &catalog); err != nil {
		return CollisionShapeCatalog{}, err
	}

	return catalog, nil
}

func (catalog CollisionShapeCatalog) BulletShape() (CollisionShape, error) {
	return catalog.Bullet.ToCollisionShape(1)
}

func (catalog CollisionShapeCatalog) ShipShape() (CollisionShape, error) {
	return catalog.Ship.ToCollisionShape(1)
}

func (catalog CollisionShapeCatalog) ShipShapeByID(shapeID string) (CollisionShape, error) {
	switch shapeID {
	case "", DefaultShipCollisionShapeID:
		return catalog.ShipShape()
	default:
		return catalog.ShipShape()
	}
}

func (catalog CollisionShapeCatalog) AsteroidShape(variant int, size int) (CollisionShape, error) {
	if len(catalog.Asteroids) == 0 {
		return CollisionShape{}, fmt.Errorf("no asteroid collision shapes loaded")
	}

	scale := float64(size) * constants.AsteroidSizeScale
	return catalog.Asteroids[wrapIndex(variant, len(catalog.Asteroids))].ToCollisionShape(scale)
}

func (catalog CollisionShapeCatalog) PickupShape(pickupType string) (CollisionShape, error) {
	if len(catalog.Pickups) == 0 {
		return CollisionShape{}, fmt.Errorf("no pickup collision shapes loaded")
	}

	shape, ok := catalog.Pickups[pickupType]
	if !ok {
		return CollisionShape{}, fmt.Errorf("unknown pickup collision shape %q", pickupType)
	}

	return shape.ToCollisionShape(1)
}

func (shape ImportedCollisionShape) ToCollisionShape(scale float64) (CollisionShape, error) {
	switch shape.Type {
	case "circle":
		if shape.Radius <= 0 {
			return CollisionShape{}, fmt.Errorf("invalid circle radius for %s", shape.Name)
		}
		return NewCircleShape(shape.Radius * scale), nil
	case "capsule":
		if shape.Radius <= 0 || shape.Height <= 0 {
			return CollisionShape{}, fmt.Errorf("invalid capsule shape for %s", shape.Name)
		}
		return NewCapsuleShape(shape.Radius*scale, shape.Height*scale), nil
	case "rectangle":
		if len(shape.Size) != 2 {
			return CollisionShape{}, fmt.Errorf("invalid rectangle size for %s", shape.Name)
		}
		return NewRectangleShape(shape.Size[0]*scale, shape.Size[1]*scale), nil
	case "polygon":
		points := make([]Vector2, 0, len(shape.Points))
		for _, point := range shape.Points {
			if len(point) != 2 {
				return CollisionShape{}, fmt.Errorf("invalid polygon point for %s", shape.Name)
			}
			points = append(points, Vector2{X: point[0] * scale, Y: point[1] * scale})
		}
		return NewPolygonShape(points), nil
	default:
		return CollisionShape{}, fmt.Errorf("unsupported collision shape type %q for %s", shape.Type, shape.Name)
	}
}

func findSharedCollisionShapesPath() (string, error) {
	searched := make([]string, 0, 2)

	if executable, err := executablePath(); err == nil {
		executableDirectory := filepath.Dir(executable)
		searched = append(searched, executableDirectory)
		if path, found := findCollisionShapesFromRoot(executableDirectory); found {
			return path, nil
		}
	}

	if directory, err := workingDirectory(); err == nil {
		if len(searched) == 0 || filepath.Clean(directory) != filepath.Clean(searched[0]) {
			searched = append(searched, directory)
			if path, found := findCollisionShapesFromRoot(directory); found {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("shared collision shapes not found from %v", searched)
}

func findCollisionShapesFromRoot(root string) (string, bool) {
	current := root
	for {
		path := filepath.Join(current, filepath.FromSlash(collisionShapesRelativePath))
		if _, err := os.Stat(path); err == nil {
			return path, true
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", false
		}
		current = parent
	}
}

func wrapIndex(index int, count int) int {
	if count == 0 {
		return 0
	}

	wrapped := index % count
	if wrapped < 0 {
		wrapped += count
	}

	return wrapped
}
