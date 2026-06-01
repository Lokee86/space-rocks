extends GutTest

const DevtoolsPlayerTargetModel := preload("res://scripts/devtools/devtools_player_target_model.gd")
const DevtoolsTargetResolver := preload("res://scripts/devtools/devtools_target_resolver.gd")


func test_invincible_target_rows_include_game_target_first() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-1": {},
			"player-2": {},
		},
		"player_lifecycle": {
			"player-1": "active",
			"player-2": "active",
		},
		"debug_statuses": {},
	})

	var rows: Array = model.invincible_target_rows()

	assert_true(rows.size() > 0)
	assert_eq(rows[0]["player_id"], DevtoolsTargetResolver.TARGET_GAME)
	assert_eq(rows[0]["label"], DevtoolsTargetResolver.TARGET_GAME_LABEL)


func test_invincible_target_rows_keep_actual_players_after_game_target() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-1": {},
			"player-2": {},
		},
		"player_lifecycle": {
			"player-1": "active",
			"player-2": "active",
		},
		"debug_statuses": {},
	})

	var rows: Array = model.invincible_target_rows()

	var ids: Array = []
	for row in rows:
		ids.append(str(row.get("player_id", "")))

	assert_true(ids.has("player-1"))
	assert_true(ids.has("player-2"))


func test_apply_gameplay_state_reads_game_target_from_local_player_synced_state() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-1": {"target_player_id": "player-2"},
			"player-2": {},
		},
		"player_lifecycle": {
			"player-1": "active",
			"player-2": "active",
		},
		"debug_statuses": {},
	})

	assert_eq(model.game_target_player_id, "player-2")


func test_apply_gameplay_state_missing_local_player_clears_game_target_player_id() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-2": {"target_player_id": "player-3"},
		},
		"player_lifecycle": {
			"player-2": "active",
		},
		"debug_statuses": {},
	})

	assert_eq(model.game_target_player_id, "")
