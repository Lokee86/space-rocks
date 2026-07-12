extends GutTest

const DevtoolsWindowTargetSelectors := preload("res://scripts/devtools/devtools_window_target_selectors.gd")


func _configured_selectors() -> Array:
	var controls: Array[OptionButton] = []
	for _index in range(9):
		controls.append(OptionButton.new())
	var selectors := DevtoolsWindowTargetSelectors.new()
	selectors.configure(
		controls[0],
		controls[1],
		controls[2],
		controls[3],
		controls[4],
		controls[5],
		controls[6],
		controls[7],
		controls[8]
	)
	return [selectors, controls]


func test_refresh_preserves_selected_player_by_metadata() -> void:
	var configured := _configured_selectors()
	var selectors: DevtoolsWindowTargetSelectors = configured[0]
	var controls: Array[OptionButton] = configured[1]
	var rows := [
		{"label": "Local Player", "player_id": "Player-1"},
		{"label": "Player 2", "player_id": "Player-2"},
	]

	selectors.refresh_invincible_targets(rows)
	controls[0].select(1)
	selectors.refresh_invincible_targets([
		{"label": "Player 2 Updated", "player_id": "Player-2"},
		{"label": "Local Player", "player_id": "Player-1"},
	])

	assert_eq(controls[0].get_selected(), 0)
	assert_eq(str(controls[0].get_item_metadata(controls[0].get_selected())), "Player-2")
	assert_eq(controls[0].get_item_text(controls[0].get_selected()), "Player 2 Updated")


func test_refresh_falls_back_to_game_target_when_previous_selection_is_missing() -> void:
	var configured := _configured_selectors()
	var selectors: DevtoolsWindowTargetSelectors = configured[0]
	var controls: Array[OptionButton] = configured[1]

	selectors.refresh_invincible_targets([
		{"label": "Game Target", "player_id": DevtoolsTargetResolver.TARGET_GAME},
		{"label": "Player 1", "player_id": "Player-1"},
	])
	controls[0].select(1)
	selectors.refresh_invincible_targets([
		{"label": "Game Target", "player_id": DevtoolsTargetResolver.TARGET_GAME},
	])

	assert_eq(controls[0].get_selected(), 0)
	assert_eq(str(controls[0].get_item_metadata(0)), DevtoolsTargetResolver.TARGET_GAME)


func test_empty_targets_clear_selector_without_selection() -> void:
	var configured := _configured_selectors()
	var selectors: DevtoolsWindowTargetSelectors = configured[0]
	var controls: Array[OptionButton] = configured[1]

	selectors.refresh_kill_player_targets([
		{"label": "Player 1", "player_id": "Player-1"},
	])
	selectors.refresh_kill_player_targets([])

	assert_eq(controls[3].get_item_count(), 0)
	assert_eq(controls[3].get_selected(), -1)


func test_kill_targets_preserve_selected_game_target() -> void:
	var configured := _configured_selectors()
	var selectors: DevtoolsWindowTargetSelectors = configured[0]
	var controls: Array[OptionButton] = configured[1]
	var rows := [
		{"label": "Game Target", "player_id": DevtoolsTargetResolver.TARGET_GAME},
		{"label": "Player 1", "player_id": "Player-1"},
	]

	selectors.refresh_kill_player_targets(rows)
	controls[3].select(0)
	selectors.refresh_kill_player_targets([
		{"label": "Player 1 Updated", "player_id": "Player-1"},
		{"label": "Game Target Updated", "player_id": DevtoolsTargetResolver.TARGET_GAME},
	])

	assert_eq(controls[3].get_selected(), 1)
	assert_eq(str(controls[3].get_item_metadata(1)), DevtoolsTargetResolver.TARGET_GAME)
	assert_eq(controls[3].get_item_text(1), "Game Target Updated")


func test_local_player_label_and_metadata_are_preserved() -> void:
	var configured := _configured_selectors()
	var selectors: DevtoolsWindowTargetSelectors = configured[0]
	var controls: Array[OptionButton] = configured[1]

	selectors.refresh_respawn_player_targets([
		{"label": "Local Player", "player_id": "Player-1"},
	])

	assert_eq(controls[4].get_item_text(0), "Local Player")
	assert_eq(str(controls[4].get_item_metadata(0)), "Player-1")


func test_counter_selectors_receive_the_same_target_set() -> void:
	var configured := _configured_selectors()
	var selectors: DevtoolsWindowTargetSelectors = configured[0]
	var controls: Array[OptionButton] = configured[1]
	var rows := [
		{"label": "Game Target", "player_id": DevtoolsTargetResolver.TARGET_GAME},
		{"label": "Local Player", "player_id": "Player-1"},
	]

	selectors.refresh_counter_player_targets(rows)

	for control_index in range(5, 9):
		assert_eq(controls[control_index].get_item_count(), 2)
		assert_eq(controls[control_index].get_item_text(0), "Game Target")
		assert_eq(str(controls[control_index].get_item_metadata(1)), "Player-1")
		assert_eq(controls[control_index].get_selected(), 0)


func test_all_player_target_selectors_receive_the_same_rows() -> void:
	var configured := _configured_selectors()
	var selectors: DevtoolsWindowTargetSelectors = configured[0]
	var controls: Array[OptionButton] = configured[1]
	var rows := [
		{"label": "Game Target", "player_id": DevtoolsTargetResolver.TARGET_GAME},
		{"label": "Local Player", "player_id": "Player-1"},
	]

	selectors.refresh_invincible_targets(rows)
	selectors.refresh_infinite_lives_targets(rows)
	selectors.refresh_player_frozen_targets(rows)

	for control_index in range(3):
		assert_eq(controls[control_index].get_item_count(), 2)
		assert_eq(controls[control_index].get_item_text(1), "Local Player")
		assert_eq(str(controls[control_index].get_item_metadata(1)), "Player-1")