package game

import (
	"math/rand"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func TestAsteroidSpawnClearanceRejectsPositionBarelyOutsideSecondaryCamera(t *testing.T) {
	game := NewWithSeed(1)
	config := runtime.ClientConfig{VisibleWorldWidth: 100, VisibleWorldHeight: 100}
	game.cameraViews["target"] = &runtime.CameraView{X: 0, Y: 0, Config: config}
	game.cameraViews["secondary"] = &runtime.CameraView{X: 159, Y: 0, Config: config}

	spawn := physics.Vector2{X: 210, Y: 0}
	if isInsideCameraView(game.cameraViews["secondary"], spawn) {
		t.Fatal("test spawn center must remain outside the secondary camera")
	}
	if !game.isInsideAsteroidSpawnClearanceForAnyCamera(spawn) {
		t.Fatal("expected spawn within the secondary camera clearance margin to be rejected")
	}
}

func TestAsteroidSpawnClearanceAllowsExactMarginBoundary(t *testing.T) {
	view := &runtime.CameraView{
		X: 0,
		Y: 0,
		Config: runtime.ClientConfig{
			VisibleWorldWidth:  100,
			VisibleWorldHeight: 100,
		},
	}
	spawn := physics.Vector2{X: 210, Y: 0}

	if isInsideCameraViewWithMargin(view, spawn, 160) {
		t.Fatal("expected exact spawn-margin boundary to remain eligible")
	}
}

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
