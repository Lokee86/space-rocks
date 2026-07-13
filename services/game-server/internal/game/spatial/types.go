package spatial

import "github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"

type Kind uint8

const (
	KindPlayer Kind = 1 << iota
	KindProjectile
	KindAsteroid
	KindEnemy
	KindPickup
)

type KindMask uint8

const AllKinds KindMask = KindMask(KindPlayer | KindProjectile | KindAsteroid | KindEnemy | KindPickup)

func (mask KindMask) Allows(kind Kind) bool {
	return kind != 0 && mask&KindMask(kind) != 0
}

type Ref struct {
	Kind Kind
	ID   string
}

type Entry struct {
	Ref      Ref
	Position physics.Vector2
	Radius   float64
}

type Rect struct {
	Center      physics.Vector2
	HalfExtents physics.Vector2
}
