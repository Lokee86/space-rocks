extends GutTest

const DevtoolsPlayerTargetModel := preload("res://scripts/devtools/devtools_player_target_model.gd")
const DevtoolsTargetResolver := preload("res://scripts/devtools/devtools_target_resolver.gd")


func test_invincible_target_rows_without_active_game_target_do_not_include_game_target_option() -> void:
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

	assert_false(ids.has(DevtoolsTargetResolver.TARGET_GAME))


func test_invincible_target_rows_put_all_players_first() -> void:
	var model := _model_with_two_active_players()

	var rows: Array = model.invincible_target_rows()

	assert_true(rows.size() > 0)
	assert_eq(rows[0]["player_id"], DevtoolsTargetResolver.TARGET_ALL_PLAYERS)
	assert_eq(rows[0]["label"], DevtoolsTargetResolver.TARGET_ALL_PLAYERS_LABEL)


func test_invincible_target_rows_with_player_target_player_2_include_compact_game_target_after_all_players() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-1": {"target_kind": "player", "target_id": "player-2"},
			"player-2": {},
		},
		"player_lifecycle": {
			"player-1": "active",
			"player-2": "active",
		},
		"debug_statuses": {},
	})

	var rows: Array = model.invincible_target_rows()

	assert_true(rows.size() > 1)
	assert_eq(rows[0]["player_id"], DevtoolsTargetResolver.TARGET_ALL_PLAYERS)
	assert_eq(rows[1]["player_id"], DevtoolsTargetResolver.TARGET_GAME)
	assert_eq(rows[1]["label"], "Target : P2")

	var ids: Array = []
	for row in rows:
		ids.append(str(row.get("player_id", "")))

	assert_eq(ids[0], DevtoolsTargetResolver.TARGET_ALL_PLAYERS)
	assert_eq(ids[1], DevtoolsTargetResolver.TARGET_GAME)
	assert_true(ids.has("player-1"))
	assert_true(ids.has("player-2"))


func test_invincible_target_rows_with_player_target_player_10_use_compact_p10_label() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-1": {"target_kind": "player", "target_id": "player-10"},
			"player-2": {},
		},
		"player_lifecycle": {
			"player-1": "active",
			"player-2": "active",
		},
		"debug_statuses": {},
	})

	var rows: Array = model.invincible_target_rows()

	assert_true(rows.size() > 1)
	assert_eq(rows[0]["player_id"], DevtoolsTargetResolver.TARGET_ALL_PLAYERS)
	assert_eq(rows[1]["player_id"], DevtoolsTargetResolver.TARGET_GAME)
	assert_eq(rows[1]["label"], "Target : P10")


func test_apply_gameplay_state_reads_generic_player_game_target_from_local_player_synced_state() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-1": {"target_kind": "player", "target_id": "player-2"},
			"player-2": {},
		},
		"player_lifecycle": {
			"player-1": "active",
			"player-2": "active",
		},
		"debug_statuses": {},
	})

	assert_eq(model.game_target_kind, "player")
	assert_eq(model.game_target_id, "player-2")
	assert_eq(model.game_target_player_id, "player-2")


func test_apply_gameplay_state_reads_generic_asteroid_target_and_clears_compatibility_player_target() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-1": {"target_kind": "asteroid", "target_id": "asteroid-7"},
		},
		"player_lifecycle": {
			"player-1": "active",
		},
		"debug_statuses": {},
	})

	assert_eq(model.game_target_kind, "asteroid")
	assert_eq(model.game_target_id, "asteroid-7")
	assert_eq(model.game_target_player_id, "")


func test_apply_gameplay_state_reads_generic_bullet_target_and_clears_compatibility_player_target() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-1": {"target_kind": "bullet", "target_id": "bullet-9"},
		},
		"player_lifecycle": {
			"player-1": "active",
		},
		"debug_statuses": {},
	})

	assert_eq(model.game_target_kind, "bullet")
	assert_eq(model.game_target_id, "bullet-9")
	assert_eq(model.game_target_player_id, "")


func test_apply_gameplay_state_missing_target_fields_leave_target_fields_empty() -> void:
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

	assert_eq(model.game_target_kind, "")
	assert_eq(model.game_target_id, "")
	assert_eq(model.game_target_player_id, "")


func test_apply_gameplay_state_missing_local_player_clears_game_target_fields() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-2": {"target_kind": "player", "target_id": "player-3"},
		},
		"player_lifecycle": {
			"player-2": "active",
		},
		"debug_statuses": {},
	})

	assert_eq(model.game_target_kind, "")
	assert_eq(model.game_target_id, "")
	assert_eq(model.game_target_player_id, "")


func test_apply_gameplay_state_reads_local_player_target_fields() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	model.apply_gameplay_state({
		"self_id": "Player-1",
		"server_players": {
			"Player-1": {"target_kind": "player", "target_id": "Player-2"},
			"Player-2": {},
		},
		"player_lifecycle": {
			"Player-1": "active",
			"Player-2": "active",
		},
		"debug_statuses": {},
	})

	assert_eq(model.game_target_kind, "player")
	assert_eq(model.game_target_id, "Player-2")
	assert_eq(model.game_target_player_id, "Player-2")


func test_apply_gameplay_state_compatibility_target_player_id_fallback_populates_player_target_fields() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	model.apply_gameplay_state({
		"self_id": "Player-1",
		"server_players": {
			"Player-1": {"target_kind": "", "target_id": "", "target_player_id": "Player-2"},
			"Player-2": {},
		},
		"player_lifecycle": {
			"Player-1": "active",
			"Player-2": "active",
		},
		"debug_statuses": {},
	})

	assert_eq(model.game_target_kind, "player")
	assert_eq(model.game_target_id, "Player-2")
	assert_eq(model.game_target_player_id, "Player-2")


func test_invincible_target_rows_with_asteroid_target_do_not_include_game_target_option() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-1": {"target_kind": "asteroid", "target_id": "asteroid-1"},
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

	assert_false(ids.has(DevtoolsTargetResolver.TARGET_GAME))


func test_active_player_target_rows_without_active_game_target_do_not_include_game_target_option() -> void:
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

	var rows: Array = model.active_player_target_rows()
	var ids: Array = []
	for row in rows:
		ids.append(str(row.get("player_id", "")))

	assert_false(ids.has(DevtoolsTargetResolver.TARGET_GAME))


func test_active_player_target_rows_put_all_players_first() -> void:
	var model := _model_with_two_active_players()

	var rows: Array = model.active_player_target_rows()

	assert_true(rows.size() > 0)
	assert_eq(rows[0]["player_id"], DevtoolsTargetResolver.TARGET_ALL_PLAYERS)
	assert_eq(rows[0]["label"], DevtoolsTargetResolver.TARGET_ALL_PLAYERS_LABEL)


func test_active_player_target_rows_with_player_target_include_compact_game_target_after_all_players() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-1": {"target_kind": "player", "target_id": "player-2"},
			"player-2": {},
		},
		"player_lifecycle": {
			"player-1": "active",
			"player-2": "active",
		},
		"debug_statuses": {},
	})

	var rows: Array = model.active_player_target_rows()
	var ids: Array = []
	for row in rows:
		ids.append(str(row.get("player_id", "")))

	assert_true(rows.size() > 1)
	assert_eq(rows[0]["player_id"], DevtoolsTargetResolver.TARGET_ALL_PLAYERS)
	assert_eq(rows[1]["player_id"], DevtoolsTargetResolver.TARGET_GAME)
	assert_eq(rows[1]["label"], "Target : P2")
	assert_eq(ids[0], DevtoolsTargetResolver.TARGET_ALL_PLAYERS)
	assert_eq(ids[1], DevtoolsTargetResolver.TARGET_GAME)
	assert_true(ids.has("player-1"))
	assert_true(ids.has("player-2"))


func test_kill_player_target_rows_without_active_game_target_do_not_include_game_target_option() -> void:
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

	var rows: Array = model.kill_player_target_rows()
	var ids: Array = []
	for row in rows:
		ids.append(str(row.get("player_id", "")))

	assert_false(ids.has(DevtoolsTargetResolver.TARGET_GAME))


func test_kill_player_target_rows_put_all_players_first() -> void:
	var model := _model_with_two_active_players()

	var rows: Array = model.kill_player_target_rows()

	assert_true(rows.size() > 0)
	assert_eq(rows[0]["player_id"], DevtoolsTargetResolver.TARGET_ALL_PLAYERS)
	assert_eq(rows[0]["label"], DevtoolsTargetResolver.TARGET_ALL_PLAYERS_LABEL)


func test_kill_player_target_rows_with_player_target_include_compact_game_target_after_all_players() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-1": {"target_kind": "player", "target_id": "player-2"},
			"player-2": {},
		},
		"player_lifecycle": {
			"player-1": "active",
			"player-2": "active",
		},
		"debug_statuses": {},
	})

	var rows: Array = model.kill_player_target_rows()
	var ids: Array = []
	for row in rows:
		ids.append(str(row.get("player_id", "")))

	assert_true(rows.size() > 1)
	assert_eq(rows[0]["player_id"], DevtoolsTargetResolver.TARGET_ALL_PLAYERS)
	assert_eq(rows[1]["player_id"], DevtoolsTargetResolver.TARGET_GAME)
	assert_eq(rows[1]["label"], "Target : P2")
	assert_eq(ids[0], DevtoolsTargetResolver.TARGET_ALL_PLAYERS)
	assert_eq(ids[1], DevtoolsTargetResolver.TARGET_GAME)


func test_kill_player_target_rows_keep_actual_player_rows_after_conditional_game_target_row() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-1": {"target_kind": "player", "target_id": "player-2"},
			"player-2": {},
		},
		"player_lifecycle": {
			"player-1": "active",
			"player-2": "active",
		},
		"debug_statuses": {},
	})

	var rows: Array = model.kill_player_target_rows()
	var ids: Array = []
	for row in rows:
		ids.append(str(row.get("player_id", "")))

	assert_true(ids.has("player-1"))
	assert_true(ids.has("player-2"))


func test_kill_player_target_rows_with_asteroid_target_do_not_include_game_target_option() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-1": {"target_kind": "asteroid", "target_id": "asteroid-1"},
			"player-2": {},
		},
		"player_lifecycle": {
			"player-1": "active",
			"player-2": "active",
		},
		"debug_statuses": {},
	})

	var rows: Array = model.kill_player_target_rows()
	var ids: Array = []
	for row in rows:
		ids.append(str(row.get("player_id", "")))

	assert_false(ids.has(DevtoolsTargetResolver.TARGET_GAME))


func test_infinite_lives_target_rows_put_all_players_first() -> void:
	var model := _model_with_two_active_players()

	var rows: Array = model.infinite_lives_target_rows()

	assert_true(rows.size() > 0)
	assert_eq(rows[0]["player_id"], DevtoolsTargetResolver.TARGET_ALL_PLAYERS)
	assert_eq(rows[0]["label"], DevtoolsTargetResolver.TARGET_ALL_PLAYERS_LABEL)


func test_player_frozen_target_rows_put_all_players_first() -> void:
	var model := _model_with_two_active_players()

	var rows: Array = model.player_frozen_target_rows()

	assert_true(rows.size() > 0)
	assert_eq(rows[0]["player_id"], DevtoolsTargetResolver.TARGET_ALL_PLAYERS)
	assert_eq(rows[0]["label"], DevtoolsTargetResolver.TARGET_ALL_PLAYERS_LABEL)


func test_respawn_player_target_rows_put_all_players_first() -> void:
	var model := _model_with_two_active_players()

	var rows: Array = model.respawn_player_target_rows()
	var ids: Array = []
	for row in rows:
		ids.append(str(row.get("player_id", "")))

	assert_true(rows.size() > 0)
	assert_eq(rows[0]["player_id"], DevtoolsTargetResolver.TARGET_ALL_PLAYERS)
	assert_eq(rows[0]["label"], DevtoolsTargetResolver.TARGET_ALL_PLAYERS_LABEL)
	assert_true(ids.has("player-1"))
	assert_true(ids.has("player-2"))


func test_target_rows_remain_raw_and_do_not_include_all_players() -> void:
	var model := _model_with_two_active_players()

	var rows: Array = model.target_rows()
	var ids: Array = []
	for row in rows:
		ids.append(str(row.get("player_id", "")))

	assert_false(ids.has(DevtoolsTargetResolver.TARGET_ALL_PLAYERS))


func test_target_state_for_player_target_returns_raw_player_dictionary() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	var expected_player_state := {
		"score": 123,
		"lives": 2,
		"health": 1,
	}
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-1": {"target_kind": "player", "target_id": "player-2"},
			"player-2": expected_player_state,
		},
		"player_lifecycle": {
			"player-1": "active",
			"player-2": "active",
		},
		"debug_statuses": {},
	})

	assert_eq(model.target_state(), expected_player_state)


func test_local_player_state_for_state_packet_returns_raw_player_dictionary() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	var expected_player_state := {
		"score": 42,
		"lives": 3,
	}
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-1": expected_player_state,
		},
		"player_lifecycle": {
			"player-1": "active",
		},
		"debug_statuses": {},
	})

	assert_eq(model.local_player_state_for_source("state_packet"), expected_player_state)


func test_local_player_state_for_session_packet_returns_raw_player_session_dictionary() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	var expected_session_state := {
		"id": "player-1",
		"ship_type": "v_wing",
		"score": 17,
		"lives": 1,
	}
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {},
		"player_sessions": {
			"player-1": expected_session_state,
		},
		"player_lifecycle": {
			"player-1": "pending_respawn",
		},
		"debug_statuses": {},
	})

	assert_eq(model.local_player_state(), {})
	assert_eq(model.local_player_state_for_source("state_packet"), {})
	assert_eq(model.local_player_state_for_source("session_packet"), expected_session_state)


func test_target_state_for_session_packet_returns_raw_player_session_dictionary() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	var expected_session_state := {
		"id": "player-2",
		"ship_type": "falcon",
		"score": 23,
		"lives": 2,
	}
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-1": {"target_kind": "player", "target_id": "player-2"},
		},
		"player_sessions": {
			"player-2": expected_session_state,
		},
		"player_lifecycle": {
			"player-1": "active",
			"player-2": "pending_respawn",
		},
		"debug_statuses": {},
	})

	assert_eq(model.target_state(), {})
	assert_eq(model.target_state_for_source("state_packet"), {})
	assert_eq(model.target_state_for_source("session_packet"), expected_session_state)


func test_target_state_for_session_packet_returns_empty_dictionary_for_non_player_target() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-1": {"target_kind": "asteroid", "target_id": "asteroid-7"},
		},
		"server_asteroids": {
			"asteroid-7": {
				"x": 44.0,
				"y": 88.0,
			},
		},
		"player_lifecycle": {
			"player-1": "active",
		},
		"debug_statuses": {},
	})

	assert_eq(model.target_state_for_source("session_packet"), {})


func test_target_state_for_asteroid_target_returns_raw_asteroid_dictionary_when_server_asteroids_exists() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	var expected_asteroid_state := {
		"x": 44.0,
		"y": 88.0,
		"size": 3,
	}
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-1": {"target_kind": "asteroid", "target_id": "asteroid-7"},
		},
		"server_asteroids": {
			"asteroid-7": expected_asteroid_state,
		},
		"player_lifecycle": {
			"player-1": "active",
		},
		"debug_statuses": {},
	})

	assert_eq(model.target_state(), expected_asteroid_state)


func test_target_state_with_missing_or_empty_target_returns_empty_dictionary() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-1": {},
			"player-2": {"score": 10},
		},
		"player_lifecycle": {
			"player-1": "active",
			"player-2": "active",
		},
		"debug_statuses": {},
	})

	assert_eq(model.target_state(), {})


func test_local_player_state_uses_state_packet_by_default_even_when_session_data_exists() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	var expected_session_state := {
		"id": "player-1",
		"ship_type": "v_wing",
		"score": 7,
		"lives": 1,
	}
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-2": {},
		},
		"player_sessions": {
			"player-1": expected_session_state,
		},
		"player_lifecycle": {
			"player-1": "pending_respawn",
			"player-2": "active",
		},
		"debug_statuses": {},
	})

	assert_eq(model.local_player_state(), {})
	assert_eq(model.local_player_state_for_source("state_packet"), {})
	assert_eq(model.local_player_state_for_source("session_packet"), expected_session_state)


func test_target_state_for_player_uses_state_packet_by_default_even_when_session_data_exists() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	var expected_session_state := {
		"id": "player-2",
		"ship_type": "falcon",
		"score": 19,
		"lives": 2,
	}
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-1": {"target_kind": "player", "target_id": "player-2"},
		},
		"player_sessions": {
			"player-2": expected_session_state,
		},
		"player_lifecycle": {
			"player-1": "active",
			"player-2": "pending_respawn",
		},
		"debug_statuses": {},
	})

	assert_eq(model.target_state(), {})
	assert_eq(model.target_state_for_source("state_packet"), {})
	assert_eq(model.target_state_for_source("session_packet"), expected_session_state)


func test_target_state_for_player_prefers_state_packet_over_session_packet() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	var expected_player_state := {
		"score": 123,
		"lives": 2,
	}
	var session_state := {
		"id": "player-2",
		"ship_type": "v_wing",
		"score": 9,
		"lives": 0,
	}
	model.apply_gameplay_state({
		"self_id": "player-1",
		"server_players": {
			"player-1": {"target_kind": "player", "target_id": "player-2"},
			"player-2": expected_player_state,
		},
		"player_sessions": {
			"player-2": session_state,
		},
		"player_lifecycle": {
			"player-1": "active",
			"player-2": "active",
		},
		"debug_statuses": {},
	})

	assert_eq(model.target_state(), expected_player_state)
	assert_eq(model.target_state_for_source("state_packet"), expected_player_state)
	assert_eq(model.target_state_for_source("session_packet"), session_state)


func _model_with_two_active_players() -> RefCounted:
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
	return model
