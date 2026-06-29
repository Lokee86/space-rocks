extends GutTest

const GameplayAliveRestoreFlow = preload("res://scripts/gameplay/respawn/gameplay_alive_restore_flow.gd")


class FakeWorldSync:
	var clear_view_target_player_calls := 0

	func clear_view_target_player() -> void:
		clear_view_target_player_calls += 1


class FakeRespawnFlow:
	var should_restore_result := false
	var should_restore_calls := 0
	var last_state
	var last_player
	var last_has_stale_dead_presentation := false
	var clear_awaiting_confirmation_calls := 0

	func should_restore_alive_hud(state: Dictionary, player, has_stale_dead_presentation := false) -> bool:
		should_restore_calls += 1
		last_state = state
		last_player = player
		last_has_stale_dead_presentation = has_stale_dead_presentation
		return should_restore_result

	func clear_awaiting_confirmation() -> void:
		clear_awaiting_confirmation_calls += 1


class FakeHudFlow:
	var is_dead := false
	var is_game_over := false
	var set_alive_calls := 0

	func set_alive() -> void:
		set_alive_calls += 1


class FakeMenuFlow:
	var is_game_over := false
	var set_alive_calls := 0

	func set_alive() -> void:
		set_alive_calls += 1


class FakePlayer:
	pass


func _make_flow(
	world_sync,
	respawn_flow,
	hud_flow,
	menu_flow,
	player
) -> GameplayAliveRestoreFlow:
	var flow := GameplayAliveRestoreFlow.new()
	flow.configure(world_sync, respawn_flow, hud_flow, menu_flow, player)
	return flow


func _state() -> Dictionary:
	return {
		"self_id": "player-1",
		"server_players": {
			"player-1": {}
		},
	}


func test_should_restore_alive_hud_requires_awaiting_confirmation_and_confirmed_lane_state() -> void:
	var respawn_flow := GameplayAliveRestoreFlow.new()
	var player := FakePlayer.new()
	var state := _state()

	assert_false(respawn_flow.should_restore_alive_hud(state, player))

	var awaiting_respawn_flow := FakeRespawnFlow.new()
	awaiting_respawn_flow.awaiting_confirmation = true
	assert_true(awaiting_respawn_flow.should_restore_alive_hud(state, player))

	var inactive_state := _state()
	inactive_state["player_lifecycle"] = {"player-1": "pending_respawn"}
	assert_false(awaiting_respawn_flow.should_restore_alive_hud(inactive_state, player))

	var missing_ship_state := _state()
	missing_ship_state["server_players"] = {}
	assert_false(awaiting_respawn_flow.should_restore_alive_hud(missing_ship_state, player))

	var stale_flow := FakeRespawnFlow.new()
	assert_true(stale_flow.should_restore_alive_hud(state, player, true))
