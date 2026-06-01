extends GutTest

const DevtoolsDisplayRefreshFlow = preload("res://scripts/devtools/devtools_display_refresh_flow.gd")

class FakeWindowController:
	var received_rows: Array = []
	var received_target_kind := ""
	var received_target_id := ""

	func refresh_game_target_options(rows: Array, current_target_kind: String, current_target_id: String) -> void:
		received_rows = rows
		received_target_kind = current_target_kind
		received_target_id = current_target_id

	func refresh_debug_player_targets(_rows: Array, _invincible_rows: Array, _infinite_lives_rows: Array, _frozen_rows: Array) -> void:
		pass

	func refresh_counter_player_targets(_rows: Array) -> void:
		pass

	func apply_debug_status(_status: Dictionary) -> void:
		pass

	func refresh_spawn_player_slots(_max_players: int) -> void:
		pass

func test_refresh_gameplay_state_forwards_game_target_state_to_window_controller() -> void:
	var controller := FakeWindowController.new()
	var flow := DevtoolsDisplayRefreshFlow.new()
	flow.configure(controller)

	flow.refresh_gameplay_state({
		"self_id": "Player-1",
		"server_players": {
			"Player-1": {
				"target_kind": "player",
				"target_id": "Player-2",
			},
			"Player-2": {},
		},
		"player_lifecycle": {
			"Player-1": "active",
			"Player-2": "active",
		},
		"debug_statuses": {},
	})

	assert_eq(controller.received_target_kind, "player")
	assert_eq(controller.received_target_id, "Player-2")
