extends GutTest

const SpectateTargetsScript := preload("res://scripts/gameplay/spectate/spectate_targets.gd")


func test_select_target_keeps_current_active_target() -> void:
	var positions := {
		"player-b": Vector2(20.0, 20.0),
		"player-a": Vector2(10.0, 10.0),
	}
	var lifecycle := {
		"player-a": "active",
		"player-b": "active",
	}

	assert_eq(
		SpectateTargetsScript.select_target("player-local", "player-b", positions, lifecycle),
		"player-b"
	)


func test_select_target_chooses_deterministic_active_target_without_current() -> void:
	var positions := {
		"player-c": Vector2(30.0, 30.0),
		"player-a": Vector2(10.0, 10.0),
		"player-b": Vector2(20.0, 20.0),
	}
	var lifecycle := {
		"player-a": "active",
		"player-b": "active",
		"player-c": "active",
	}

	assert_eq(
		SpectateTargetsScript.select_target("player-local", "", positions, lifecycle),
		"player-a"
	)


func test_select_target_switches_when_current_disappears() -> void:
	var positions := {
		"player-c": Vector2(30.0, 30.0),
		"player-a": Vector2(10.0, 10.0),
	}
	var lifecycle := {
		"player-a": "active",
		"player-c": "active",
	}

	assert_eq(
		SpectateTargetsScript.select_target("player-local", "player-b", positions, lifecycle),
		"player-a"
	)


func test_select_target_switches_when_current_becomes_ineligible() -> void:
	var positions := {
		"player-a": Vector2(10.0, 10.0),
		"player-b": Vector2(20.0, 20.0),
	}
	var lifecycle := {
		"player-a": "active",
		"player-b": "eliminated",
	}

	assert_eq(
		SpectateTargetsScript.select_target("player-local", "player-b", positions, lifecycle),
		"player-a"
	)


func test_active_player_without_position_is_not_eligible() -> void:
	var positions := {
		"player-a": Vector2(10.0, 10.0),
	}
	var lifecycle := {
		"player-a": "pending_respawn",
		"player-b": "active",
	}

	assert_eq(SpectateTargetsScript.select_target("player-local", "", positions, lifecycle), "")


func test_pending_respawn_player_with_position_is_not_eligible() -> void:
	var positions := {
		"player-a": Vector2(10.0, 10.0),
	}
	var lifecycle := {
		"player-a": "pending_respawn",
	}

	assert_eq(SpectateTargetsScript.select_target("player-local", "", positions, lifecycle), "")


func test_eliminated_player_with_position_is_not_eligible() -> void:
	var positions := {
		"player-a": Vector2(10.0, 10.0),
	}
	var lifecycle := {
		"player-a": "eliminated",
	}

	assert_eq(SpectateTargetsScript.select_target("player-local", "", positions, lifecycle), "")


func test_missing_lifecycle_entry_is_not_eligible() -> void:
	var positions := {
		"player-a": Vector2(10.0, 10.0),
	}

	assert_eq(SpectateTargetsScript.select_target("player-local", "", positions, {}), "")


func test_cycle_target_selects_next_active_target() -> void:
	var positions := {
		"player-a": Vector2(10.0, 10.0),
		"player-b": Vector2(20.0, 20.0),
		"player-c": Vector2(30.0, 30.0),
	}
	var lifecycle := {
		"player-a": "active",
		"player-b": "active",
		"player-c": "active",
	}

	assert_eq(
		SpectateTargetsScript.cycle_target("player-local", "player-a", positions, lifecycle),
		"player-b"
	)


func test_cycle_target_wraps_to_first_active_target() -> void:
	var positions := {
		"player-a": Vector2(10.0, 10.0),
		"player-b": Vector2(20.0, 20.0),
	}
	var lifecycle := {
		"player-a": "active",
		"player-b": "active",
	}

	assert_eq(
		SpectateTargetsScript.cycle_target("player-local", "player-b", positions, lifecycle),
		"player-a"
	)


func test_select_target_returns_empty_without_remote_players() -> void:
	var positions := {
		"player-local": Vector2(5.0, 5.0),
	}
	var lifecycle := {
		"player-local": "active",
	}

	assert_eq(SpectateTargetsScript.select_target("player-local", "", positions, lifecycle), "")
	assert_eq(SpectateTargetsScript.cycle_target("player-local", "", positions, lifecycle), "")
