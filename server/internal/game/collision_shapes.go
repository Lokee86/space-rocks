package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Lokee86/space-rocks/server/internal/constants"
)

type CollisionShapeCatalog struct {
	Bullet    ImportedCollisionShape   `json:"bullet"`
	Asteroids []ImportedCollisionShape `json:"asteroids"`
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

func (catalog CollisionShapeCatalog) AsteroidShape(variant int, size int) (CollisionShape, error) {
	if len(catalog.Asteroids) == 0 {
		return CollisionShape{}, fmt.Errorf("no asteroid collision shapes loaded")
	}

	scale := float64(size) * constants.AsteroidSizeScale
	return catalog.Asteroids[wrapIndex(variant, len(catalog.Asteroids))].ToCollisionShape(scale)
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
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", err
	}

	current := workingDirectory
	for {
		path := filepath.Join(current, "shared", "collisions", "collision_shapes.json")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("shared collision shapes not found from %s", workingDirectory)
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
