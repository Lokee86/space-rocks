package game

import "testing"

func TestPresentationDerivedCachesAdjacentGenerations(t *testing.T) {
	game := New()
	builds := 0
	build := func(value string) func() (any, error) {
		return func() (any, error) {
			builds++
			return value, nil
		}
	}

	first, err := game.PresentationDerived(10, "world", build("ten"))
	if err != nil || first != "ten" {
		t.Fatalf("build first generation: value=%v err=%v", first, err)
	}
	second, err := game.PresentationDerived(11, "world", build("eleven"))
	if err != nil || second != "eleven" {
		t.Fatalf("build second generation: value=%v err=%v", second, err)
	}
	firstAgain, err := game.PresentationDerived(10, "world", build("unexpected"))
	if err != nil || firstAgain != "ten" {
		t.Fatalf("reuse adjacent generation: value=%v err=%v", firstAgain, err)
	}
	if builds != 2 {
		t.Fatalf("build count = %d, want 2", builds)
	}

	if _, err := game.PresentationDerived(12, "world", build("twelve")); err != nil {
		t.Fatalf("build third generation: %v", err)
	}
	if _, err := game.PresentationDerived(10, "world", build("ten-rebuilt")); err != nil {
		t.Fatalf("rebuild evicted generation: %v", err)
	}
	if builds != 4 {
		t.Fatalf("build count after eviction = %d, want 4", builds)
	}
}
