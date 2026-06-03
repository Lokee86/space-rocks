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


func test_apply_state_does_nothing_when_restore_is_rejected() -> void:
	var world_sync := FakeWorldSync.new()
	var respawn_flow := FakeRespawnFlow.new()
	respawn_flow.should_restore_result = false
	var hud_flow := FakeHudFlow.new()
	var menu_flow := FakeMenuFlow.new()
	var player := FakePlayer.new()

	var flow := _make_flow(world_sync, respawn_flow, hud_flow, menu_flow, player)
	flow.apply_state(_state())

	assert_eq(respawn_flow.should_restore_calls, 1)
	assert_eq(world_sync.clear_view_target_player_calls, 0)
	assert_eq(hud_flow.set_alive_calls, 0)
	assert_eq(menu_flow.set_alive_calls, 0)
	assert_eq(respawn_flow.clear_awaiting_confirmation_calls, 0)


func test_apply_state_restores_world_hud_menu_and_respawn() -> void:
	var world_sync := FakeWorldSync.new()
	var respawn_flow := FakeRespawnFlow.new()
	respawn_flow.should_restore_result = true
	var hud_flow := FakeHudFlow.new()
	var menu_flow := FakeMenuFlow.new()
	var player := FakePlayer.new()

	var flow := _make_flow(world_sync, respawn_flow, hud_flow, menu_flow, player)
	flow.apply_state(_state())

	assert_eq(world_sync.clear_view_target_player_calls, 1)
	assert_eq(hud_flow.set_alive_calls, 1)
	assert_eq(menu_flow.set_alive_calls, 1)
	assert_eq(respawn_flow.clear_awaiting_confirmation_calls, 1)


func test_apply_state_restores_without_menu_flow() -> void:
	var world_sync := FakeWorldSync.new()
	var respawn_flow := FakeRespawnFlow.new()
	respawn_flow.should_restore_result = true
	var hud_flow := FakeHudFlow.new()
	var player := FakePlayer.new()

	var flow := _make_flow(world_sync, respawn_flow, hud_flow, null, player)
	flow.apply_state(_state())

	assert_eq(world_sync.clear_view_target_player_calls, 1)
	assert_eq(hud_flow.set_alive_calls, 1)
	assert_eq(respawn_flow.clear_awaiting_confirmation_calls, 1)


func test_apply_state_marks_stale_dead_presentation_from_hud_and_menu() -> void:
	var world_sync := FakeWorldSync.new()
	var respawn_flow := FakeRespawnFlow.new()
	respawn_flow.should_restore_result = true
	var hud_flow := FakeHudFlow.new()
	hud_flow.is_dead = true
	var menu_flow := FakeMenuFlow.new()
	var player := FakePlayer.new()

	var flow := _make_flow(world_sync, respawn_flow, hud_flow, menu_flow, player)
	flow.apply_state(_state())

	assert_true(respawn_flow.last_has_stale_dead_presentation)

	hud_flow.is_dead = false
	hud_flow.is_game_over = true
	flow.apply_state(_state())
	assert_true(respawn_flow.last_has_stale_dead_presentation)

	hud_flow.is_game_over = false
	menu_flow.is_game_over = true
	flow.apply_state(_state())
	assert_true(respawn_flow.last_has_stale_dead_presentation)
