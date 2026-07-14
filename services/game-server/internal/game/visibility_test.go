package game

import (
	"math/rand"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func TestRandomOffscreenPositionUsesGameSeededRNG(t *testing.T) {
	const seed int64 = 24681357
	const margin = 12.5

	view := &runtime.CameraView{X: 400, Y: -125}
	game := NewWithSeed(seed)
	wantRng := rand.New(rand.NewSource(seed))

	gotFirst := game.randomOffscreenPosition(view, margin)
	wantFirst := expectedOffscreenPosition(wantRng, view, margin)
	if gotFirst != wantFirst {
		t.Fatalf("first randomOffscreenPosition() = %#v, want %#v", gotFirst, wantFirst)
	}

	gotSecond := game.randomOffscreenPosition(view, margin)
	wantSecond := expectedOffscreenPosition(wantRng, view, margin)
	if gotSecond != wantSecond {
		t.Fatalf("second randomOffscreenPosition() = %#v, want %#v", gotSecond, wantSecond)
	}
}

func expectedOffscreenPosition(rng *rand.Rand, view *runtime.CameraView, margin float64) physics.Vector2 {
	width := view.VisibleWorldWidth()
	height := view.VisibleWorldHeight()
	left := view.X - width*0.5
	right := view.X + width*0.5
	top := view.Y - height*0.5
	bottom := view.Y + height*0.5

	switch rng.Intn(4) {
	case 0:
		return physics.Vector2{X: left + rng.Float64()*(right-left), Y: top - margin}
	case 1:
		return physics.Vector2{X: right + margin, Y: top + rng.Float64()*(bottom-top)}
	case 2:
		return physics.Vector2{X: left + rng.Float64()*(right-left), Y: bottom + margin}
	default:
		return physics.Vector2{X: left - margin, Y: top + rng.Float64()*(bottom-top)}
	}
}
