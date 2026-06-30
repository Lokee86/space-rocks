extends GutTest

const DevtoolsPlayerTargetModel := preload("res://scripts/devtools/devtools_player_target_model.gd")
const DevtoolsTargetResolver := preload("res://scripts/devtools/devtools_target_resolver.gd")
const DevtoolsLaneStateAdapter := preload("res://scripts/protocol/realtime/devtools_lane_state_adapter.gd")


func test_model_constructs_and_resets_lane_native_state() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	model.self_id = "player-1"
	model.world_ships = {"player-1": {"id": "player-1"}}
	model.world_asteroids = {"asteroid-1": {"id": "asteroid-1"}}
	model.world_bullets = {"bullet-1": {"id": "bullet-1"}}
	model.world_pickups = {"pickup-1": {"id": "pickup-1"}}
	model.session_players = {"player-1": {"id": "player-1"}}
	model.session_player_lifecycle = {"player-1": "active"}
	model.overlay_self_id = "player-1"

	model.reset()

	assert_eq(model.self_id, "")
	assert_true(model.world_ships.is_empty())
	assert_true(model.world_asteroids.is_empty())
	assert_true(model.world_bullets.is_empty())
	assert_true(model.world_pickups.is_empty())
	assert_true(model.session_players.is_empty())
	assert_true(model.session_player_lifecycle.is_empty())
	assert_eq(model.overlay_self_id, "")


func test_apply_gameplay_state_reads_lane_native_groups() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	model.apply_gameplay_state({
		"self_id": "player-1",
		"world": {
			"ships": {
				"player-1": {"target_kind": "player", "target_id": "player-2"},
				"player-2": {"id": "player-2"},
			},
			"asteroids": {"asteroid-7": {"id": "asteroid-7"}},
			"bullets": {"bullet-9": {"id": "bullet-9"}},
			"pickups": {"pickup-1": {"id": "pickup-1"}},
		},
		"session": {
			"players": {
				"player-1": {"id": "player-1", "score": 42, "lives": 3},
				"player-2": {"id": "player-2", "score": 99, "lives": 2},
			},
			"player_lifecycle": {
				"player-1": "active",
				"player-2": "active",
			},
		},
		"overlay": {
			"self_id": "player-1",
		},
	})

	assert_eq(model.self_id, "player-1")
	assert_eq(model.overlay_self_id, "player-1")
	assert_eq(model.world_ships["player-1"]["target_id"], "player-2")
	assert_eq(model.world_asteroids["asteroid-7"]["id"], "asteroid-7")
	assert_eq(model.world_bullets["bullet-9"]["id"], "bullet-9")
	assert_eq(model.world_pickups["pickup-1"]["id"], "pickup-1")
	assert_eq(model.session_players["player-1"]["score"], 42)
	assert_eq(model.session_player_lifecycle["player-2"], "active")
	assert_eq(model.game_target_kind, "player")
	assert_eq(model.game_target_id, "player-2")
	assert_eq(model.game_target_player_id, "player-2")


func test_local_player_state_for_player_world_states_uses_session_players() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	var expected_session_state := {"id": "player-1", "score": 17, "lives": 1}
	model.apply_gameplay_state({
		"self_id": "player-1",
		"world": {"ships": {}, "asteroids": {}, "bullets": {}, "pickups": {}},
		"session": {"players": {"player-1": expected_session_state}, "player_lifecycle": {"player-1": "pending_respawn"}},
		"overlay": {"self_id": "player-1"},
	})

	assert_eq(model.local_player_state(), {})
	assert_eq(model.local_player_state_for_source("players"), {})
	assert_eq(model.local_player_state_for_source("player_world_states"), expected_session_state)


func test_target_state_for_player_world_states_returns_raw_player_session_dictionary() -> void:
	var model := DevtoolsPlayerTargetModel.new()
	var expected_session_state := {"id": "player-2", "score": 23, "lives": 2}
	model.apply_gameplay_state({
		"self_id": "player-1",
		"world": {
			"ships": {"player-1": {"target_kind": "player", "target_id": "player-2"}},
			"asteroids": {},
			"bullets": {},
			"pickups": {},
		},
		"session": {
			"players": {"player-2": expected_session_state},
			"player_lifecycle": {"player-1": "active", "player-2": "pending_respawn"},
		},
		"overlay": {"self_id": "player-1"},
	})

	assert_eq(model.target_state(), {})
	assert_eq(model.target_state_for_source("players"), {})
	assert_eq(model.target_state_for_source("player_world_states"), expected_session_state)


func test_apply_gameplay_state_with_lane_adapter_output_populates_target_rows() -> void:
	var adapter: DevtoolsLaneStateAdapter = DevtoolsLaneStateAdapter.new()
	var router := {
		"overlay_lane_state": {"self_id": "player-1"},
		"world_lane_state": {
			"ships": {
				"player-1": {"id": "player-1"},
				"player-2": {"id": "player-2"},
			},
			"asteroids": {"asteroid-1": {"id": "asteroid-1"}},
			"bullets": {"bullet-1": {"id": "bullet-1"}},
			"pickups": {"pickup-1": {"id": "pickup-1"}},
		},
		"session_lane_state": {
			"player_sessions": {
				"player-1": {"id": "player-1"},
				"player-2": {"id": "player-2"},
			},
			"player_lifecycle": {
				"player-1": "active",
				"player-2": "active",
			},
		},
	}
	var model := DevtoolsPlayerTargetModel.new()
	model.apply_gameplay_state(adapter.build_state(router))

	var rows: Array = model.target_rows()
	var ids: Array = []
	for row in rows:
		ids.append(str(row.get("player_id", "")))

	assert_true(ids.has("player-1"))
	assert_true(ids.has("player-2"))
	assert_eq(model.self_id, "player-1")
	assert_eq(model.world_asteroids["asteroid-1"]["id"], "asteroid-1")
	assert_eq(model.session_players["player-2"]["id"], "player-2")
	assert_eq(model.local_player_state_for_source("players"), {"id": "player-1"})
	assert_eq(model.local_player_state_for_source("player_world_states"), {"id": "player-1"})

	var active_rows: Array = model.active_player_target_rows()
	var active_ids: Array = []
	for row in active_rows:
		active_ids.append(str(row.get("player_id", "")))

	assert_true(active_ids.has(DevtoolsTargetResolver.TARGET_ALL_PLAYERS))
	assert_true(active_ids.has("player-1"))
	assert_true(active_ids.has("player-2"))
