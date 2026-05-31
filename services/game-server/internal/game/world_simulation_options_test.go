package game

import "testing"

func TestWorldSimulationOptionsSetFreezeWorldSetsAllFlags(t *testing.T) {
	var options WorldSimulationOptions

	options.SetFreezeWorld(true)
	if !options.FreezeAsteroids {
		t.Fatal("expected SetFreezeWorld(true) to set FreezeAsteroids")
	}
	if !options.FreezeBullets {
		t.Fatal("expected SetFreezeWorld(true) to set FreezeBullets")
	}
	if !options.FreezeSpawning {
		t.Fatal("expected SetFreezeWorld(true) to set FreezeSpawning")
	}
	if !options.FreezeCollisions {
		t.Fatal("expected SetFreezeWorld(true) to set FreezeCollisions")
	}

	options.SetFreezeWorld(false)
	if options.FreezeAsteroids {
		t.Fatal("expected SetFreezeWorld(false) to clear FreezeAsteroids")
	}
	if options.FreezeBullets {
		t.Fatal("expected SetFreezeWorld(false) to clear FreezeBullets")
	}
	if options.FreezeSpawning {
		t.Fatal("expected SetFreezeWorld(false) to clear FreezeSpawning")
	}
	if options.FreezeCollisions {
		t.Fatal("expected SetFreezeWorld(false) to clear FreezeCollisions")
	}
}

func TestWorldSimulationOptionsIsWorldFrozenRequiresAllFlags(t *testing.T) {
	options := WorldSimulationOptions{
		FreezeAsteroids:  true,
		FreezeBullets:    true,
		FreezeSpawning:   true,
		FreezeCollisions: true,
	}

	if !options.IsWorldFrozen() {
		t.Fatal("expected world to be frozen when all freeze flags are true")
	}

	options.FreezeAsteroids = false
	if options.IsWorldFrozen() {
		t.Fatal("expected world not to be frozen when FreezeAsteroids is false")
	}
	options.FreezeAsteroids = true

	options.FreezeBullets = false
	if options.IsWorldFrozen() {
		t.Fatal("expected world not to be frozen when FreezeBullets is false")
	}
	options.FreezeBullets = true

	options.FreezeSpawning = false
	if options.IsWorldFrozen() {
		t.Fatal("expected world not to be frozen when FreezeSpawning is false")
	}
	options.FreezeSpawning = true

	options.FreezeCollisions = false
	if options.IsWorldFrozen() {
		t.Fatal("expected world not to be frozen when FreezeCollisions is false")
	}
}

func TestWorldSimulationOptionsToggleFreezeWorldFromPartialFreezeEnablesAllFlags(t *testing.T) {
	options := WorldSimulationOptions{
		FreezeAsteroids: true,
	}

	enabled := options.ToggleFreezeWorld()
	if !enabled {
		t.Fatal("expected ToggleFreezeWorld from partial freeze to enable full freeze")
	}
	if !options.FreezeAsteroids {
		t.Fatal("expected FreezeAsteroids to remain enabled")
	}
	if !options.FreezeBullets {
		t.Fatal("expected FreezeBullets to be enabled")
	}
	if !options.FreezeSpawning {
		t.Fatal("expected FreezeSpawning to be enabled")
	}
	if !options.FreezeCollisions {
		t.Fatal("expected FreezeCollisions to be enabled")
	}
}

func TestWorldSimulationOptionsToggleFreezeWorldFromFullFreezeDisablesAllFlags(t *testing.T) {
	options := WorldSimulationOptions{
		FreezeAsteroids:  true,
		FreezeBullets:    true,
		FreezeSpawning:   true,
		FreezeCollisions: true,
	}

	enabled := options.ToggleFreezeWorld()
	if enabled {
		t.Fatal("expected ToggleFreezeWorld from full freeze to disable full freeze")
	}
	if options.FreezeAsteroids {
		t.Fatal("expected FreezeAsteroids to be disabled")
	}
	if options.FreezeBullets {
		t.Fatal("expected FreezeBullets to be disabled")
	}
	if options.FreezeSpawning {
		t.Fatal("expected FreezeSpawning to be disabled")
	}
	if options.FreezeCollisions {
		t.Fatal("expected FreezeCollisions to be disabled")
	}
}
