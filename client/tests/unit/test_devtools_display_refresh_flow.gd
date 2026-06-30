extends GutTest

const DevtoolsDisplayRefreshFlow = preload("res://scripts/devtools/devtools_display_refresh_flow.gd")

class FakeWindowController:
	var received_rows: Array = []
	var received_target_kind := ""
	var received_target_id := ""
	var received_target_state: Dictionary = {}

	func refresh_game_target_options(rows: Array, current_target_kind: String, current_target_id: String) -> void:
		received_rows = rows
		received_target_kind = current_target_kind
		received_target_id = current_target_id

	func refresh_target_state(target_kind: String, target_id: String, state: Dictionary) -> void:
		received_target_kind = target_kind
		received_target_id = target_id
		received_target_state = state

	func refresh_debug_player_targets(
		_kill_player_rows: Array,
		_target_rows: Array,
		_invincible_rows: Array,
		_infinite_lives_rows: Array,
		_player_frozen_rows: Array
	) -> void:
		pass

	func refresh_counter_player_targets(_rows: Array) -> void:
		pass

	func apply_debug_status(_status: Dictionary) -> void:
		pass

	func refresh_spawn_player_slots(_max_players: int) -> void:
		pass


class FakeWindowControllerWithoutTelemetrySources extends FakeWindowController:
	pass

func test_refresh_gameplay_state_forwards_game_target_state_to_window_controller() -> void:
	var controller := FakeWindowController.new()
	var flow := DevtoolsDisplayRefreshFlow.new()
	flow.configure(controller)

	flow.refresh_gameplay_state({
		"self_id": "Player-1",
		"world": {
			"ships": {
				"Player-1": {
					"target_kind": "player",
					"target_id": "Player-2",
				},
				"Player-2": {},
			},
		},
		"session": {
			"players": {},
			"player_lifecycle": {
				"Player-1": "active",
				"Player-2": "active",
			},
		},
		"debug_statuses": {},
	})

	assert_eq(controller.received_target_kind, "player")
	assert_eq(controller.received_target_id, "Player-2")


func test_refresh_gameplay_state_forwards_target_kind_target_id_and_raw_target_state() -> void:
	var controller := FakeWindowController.new()
	var flow := DevtoolsDisplayRefreshFlow.new()
	flow.configure(controller)

	var expected_target_state := {
		"x": 17.0,
		"y": 32.0,
		"size": 2,
	}
	flow.refresh_gameplay_state({
		"self_id": "Player-1",
		"world": {
			"ships": {
				"Player-1": {
					"target_kind": "asteroid",
					"target_id": "asteroid-3",
				},
			},
			"asteroids": {
				"asteroid-3": expected_target_state,
			},
		},
		"session": {
			"players": {},
			"player_lifecycle": {
				"Player-1": "active",
			},
		},
		"debug_statuses": {},
	})

	assert_eq(controller.received_target_kind, "asteroid")
	assert_eq(controller.received_target_id, "asteroid-3")
	assert_eq(controller.received_target_state, expected_target_state)


func test_refresh_gameplay_state_defaults_to_lane_sources_when_selectors_missing() -> void:
	var controller := FakeWindowControllerWithoutTelemetrySources.new()
	var flow := DevtoolsDisplayRefreshFlow.new()
	flow.configure(controller)

	flow.refresh_gameplay_state({
		"self_id": "Player-1",
		"world": {
			"ships": {
				"Player-1": {
					"target_kind": "player",
					"target_id": "Player-2",
				},
				"Player-2": {
					"score": 7,
					"lives": 2,
				},
			},
		},
		"session": {
			"players": {
				"Player-1": {
					"id": "Player-1",
					"score": 9,
					"lives": 3,
				},
				"Player-2": {
					"id": "Player-2",
					"score": 11,
					"lives": 1,
				},
			},
			"player_lifecycle": {
				"Player-1": "active",
				"Player-2": "active",
			},
		},
		"debug_statuses": {},
	})

	assert_eq(controller.received_target_kind, "player")
	assert_eq(controller.received_target_id, "Player-2")
	assert_eq(controller.received_target_state, {"score": 7, "lives": 2})
