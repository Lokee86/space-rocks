extends GutTest

const DevtoolsLaneStateAdapter := preload("res://scripts/protocol/realtime/devtools_lane_state_adapter.gd")


func test_build_state_returns_devtools_compatible_keys_from_lane_state() -> void:
	var adapter := DevtoolsLaneStateAdapter.new()
	var router := {
		"overlay_lane_state": {"self_id": "player-1"},
		"world_lane_state": {
			"ships": {
				"player-1": {"id": "player-1", "x": 10},
				"player-2": {"id": "player-2", "x": 20},
			},
			"asteroids": {"asteroid-1": {"id": "asteroid-1"}},
			"bullets": {"bullet-1": {"id": "bullet-1"}},
			"pickups": {"pickup-1": {"id": "pickup-1"}},
		},
		"session_lane_state": {
			"player_sessions": {
				"player-1": {"id": "player-1", "score": 1},
				"player-2": {"id": "player-2", "score": 2},
			},
			"player_lifecycle": {
				"player-1": "active",
				"player-2": "active",
			},
		},
	}

	var state := adapter.build_state(router)

	assert_eq(state.get("self_id"), "player-1")
	assert_true(state.get("server_players", {}) is Dictionary)
	assert_true(state.get("player_sessions", {}) is Dictionary)
	assert_true(state.get("server_asteroids", {}) is Dictionary)
	assert_true(state.get("server_bullets", {}) is Dictionary)
	assert_true(state.get("server_pickups", {}) is Dictionary)
	assert_true(state.get("player_lifecycle", {}) is Dictionary)
	assert_eq(state["server_players"]["player-1"]["x"], 10)
	assert_eq(state["server_players"]["player-2"]["x"], 20)
	assert_eq(state["player_sessions"]["player-1"]["score"], 1)
	assert_eq(state["player_sessions"]["player-2"]["score"], 2)
	assert_eq(state["server_asteroids"].size(), 1)
	assert_eq(state["server_bullets"].size(), 1)
	assert_eq(state["server_pickups"].size(), 1)
	assert_eq(state["player_lifecycle"]["player-1"], "active")
	assert_eq(state["player_lifecycle"]["player-2"], "active")