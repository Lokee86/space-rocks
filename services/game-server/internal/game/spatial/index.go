package spatial

import "github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"

type Index interface {
	Rebuild(entries []Entry)
	QueryCircle(dst []Ref, center physics.Vector2, radius float64, mask KindMask) []Ref
	QueryRect(dst []Ref, rect Rect, mask KindMask) []Ref
}