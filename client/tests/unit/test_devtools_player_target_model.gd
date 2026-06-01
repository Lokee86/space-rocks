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


func test_invincible_target_rows_with_player_target_player_2_include_compact_game_target_first_row() -> void:
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

	assert_true(rows.size() > 0)
	assert_eq(rows[0]["player_id"], DevtoolsTargetResolver.TARGET_GAME)
	assert_eq(rows[0]["label"], "Target : P2")

	var ids: Array = []
	for row in rows:
		ids.append(str(row.get("player_id", "")))

	assert_eq(ids[0], DevtoolsTargetResolver.TARGET_GAME)
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

	assert_true(rows.size() > 0)
	assert_eq(rows[0]["player_id"], DevtoolsTargetResolver.TARGET_GAME)
	assert_eq(rows[0]["label"], "Target : P10")


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


func test_active_player_target_rows_with_player_target_include_compact_game_target_first_row() -> void:
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

	assert_true(rows.size() > 0)
	assert_eq(rows[0]["player_id"], DevtoolsTargetResolver.TARGET_GAME)
	assert_eq(rows[0]["label"], "Target : P2")
	assert_eq(ids[0], DevtoolsTargetResolver.TARGET_GAME)
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


func test_kill_player_target_rows_with_player_target_include_compact_game_target_first_row() -> void:
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

	assert_true(rows.size() > 0)
	assert_eq(rows[0]["player_id"], DevtoolsTargetResolver.TARGET_GAME)
	assert_eq(rows[0]["label"], "Target : P2")
	assert_eq(ids[0], DevtoolsTargetResolver.TARGET_GAME)


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
