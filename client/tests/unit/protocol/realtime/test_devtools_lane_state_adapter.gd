extends GutTest

const DevtoolsLaneStateAdapter := preload("res://scripts/protocol/realtime/devtools_lane_state_adapter.gd")


func test_build_state_returns_lane_native_nested_state() -> void:
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

	assert_true(state.has("world"))
	assert_true(state.has("session"))
	assert_true(state.has("overlay"))
	assert_eq(state["overlay"]["self_id"], "player-1")
	assert_eq(state["world"]["ships"]["player-1"]["x"], 10)
	assert_eq(state["world"]["ships"]["player-2"]["x"], 20)
	assert_eq(state["world"]["asteroids"].size(), 1)
	assert_eq(state["world"]["bullets"].size(), 1)
	assert_eq(state["world"]["pickups"].size(), 1)
	assert_eq(state["session"]["players"]["player-1"]["score"], 1)
	assert_eq(state["session"]["players"]["player-2"]["score"], 2)
	assert_eq(state["session"]["player_lifecycle"]["player-1"], "active")
	assert_eq(state["session"]["player_lifecycle"]["player-2"], "active")
	assert_false(state.has("server_players"))
	assert_false(state.has("player_sessions"))
	assert_false(state.has("server_asteroids"))
	assert_false(state.has("server_bullets"))
	assert_false(state.has("server_pickups"))
	assert_false(state.has("player_lifecycle"))
