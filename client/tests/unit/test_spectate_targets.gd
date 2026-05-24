extends GutTest

const SpectateTargetsScript := preload("res://scripts/spectate_targets.gd")


func test_select_target_keeps_current_valid_target() -> void:
	var positions := {
		"player-b": Vector2(20.0, 20.0),
		"player-a": Vector2(10.0, 10.0),
	}

	assert_eq(
		SpectateTargetsScript.select_target("player-local", "player-b", positions),
		"player-b"
	)


func test_select_target_chooses_deterministic_target_without_current() -> void:
	var positions := {
		"player-c": Vector2(30.0, 30.0),
		"player-a": Vector2(10.0, 10.0),
		"player-b": Vector2(20.0, 20.0),
	}

	assert_eq(
		SpectateTargetsScript.select_target("player-local", "", positions),
		"player-a"
	)


func test_select_target_switches_when_current_disappears() -> void:
	var positions := {
		"player-c": Vector2(30.0, 30.0),
		"player-a": Vector2(10.0, 10.0),
	}

	assert_eq(
		SpectateTargetsScript.select_target("player-local", "player-b", positions),
		"player-a"
	)


func test_cycle_target_selects_next_target() -> void:
	var positions := {
		"player-a": Vector2(10.0, 10.0),
		"player-b": Vector2(20.0, 20.0),
		"player-c": Vector2(30.0, 30.0),
	}

	assert_eq(
		SpectateTargetsScript.cycle_target("player-local", "player-a", positions),
		"player-b"
	)


func test_cycle_target_wraps_to_first_target() -> void:
	var positions := {
		"player-a": Vector2(10.0, 10.0),
		"player-b": Vector2(20.0, 20.0),
	}

	assert_eq(
		SpectateTargetsScript.cycle_target("player-local", "player-b", positions),
		"player-a"
	)


func test_select_target_returns_empty_without_remote_players() -> void:
	var positions := {
		"player-local": Vector2(5.0, 5.0),
	}

	assert_eq(SpectateTargetsScript.select_target("player-local", "", positions), "")
	assert_eq(SpectateTargetsScript.cycle_target("player-local", "", positions), "")
