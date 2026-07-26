extends GutTest

const SpectateMenuStateScript := preload("res://scripts/gameplay/spectate/spectate_menu_state.gd")


func test_realtime_nested_state_exposes_only_active_remote_world_ships() -> void:
	var state := SpectateMenuStateScript.new()
	state.apply_gameplay_state({
		"overlay": {"self_id": "player-1"},
		"world": {
			"ships": {
				"player-2": {},
				"player-3": {},
				"player-4": {},
			}
		},
		"session": {
			"player_lifecycle": {
				"player-1": "eliminated",
				"player-2": "pending_respawn",
				"player-3": "active",
				"player-4": "eliminated",
			}
		},
	})

	assert_eq(state.self_id, "player-1")
	assert_eq(state.spectate_target_ids(), ["player-3"])
	assert_true(state.has_spectate_targets())


func test_world_ship_without_lifecycle_record_remains_spectatable() -> void:
	var state := SpectateMenuStateScript.new()
	state.apply_gameplay_state({
		"overlay": {"self_id": "player-1"},
		"world": {"ships": {"player-2": {}}},
		"session": {"player_lifecycle": {}},
	})

	assert_eq(state.begin_spectating(), "player-2")


func test_legacy_flat_state_remains_supported() -> void:
	var state := SpectateMenuStateScript.new()
	state.apply_gameplay_state({
		"self_id": "player-1",
		"ships": {"player-2": {}},
		"player_lifecycle": {
			"player-1": "eliminated",
			"player-2": "active",
		},
	})

	assert_eq(state.begin_spectating(), "player-2")
