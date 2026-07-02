extends GutTest

const GameplayAliveRestoreFlow = preload("res://scripts/gameplay/respawn/gameplay_alive_restore_flow.gd")
const GameplayRespawnFlow = preload("res://scripts/gameplay/respawn/gameplay_respawn_flow.gd")


class FakeWorldSync:
	var clear_view_target_player_calls := 0

	func clear_view_target_player() -> void:
		clear_view_target_player_calls += 1


class FakeRespawnFlow:
	var should_restore_result := false
	var should_restore_calls := 0
	var last_world_ships: Dictionary = {}
	var last_player_lifecycle: Dictionary = {}
	var last_self_id := ""
	var last_player
	var last_has_stale_dead_presentation := false
	var awaiting_respawn_confirmation := false
	var clear_awaiting_confirmation_calls := 0

	func should_restore_alive_hud(world_ships: Dictionary, player_lifecycle: Dictionary, self_id: String, player, has_stale_dead_presentation := false) -> bool:
		should_restore_calls += 1
		last_world_ships = world_ships
		last_player_lifecycle = player_lifecycle
		last_self_id = self_id
		last_player = player
		last_has_stale_dead_presentation = has_stale_dead_presentation
		return should_restore_result

	func clear_awaiting_confirmation() -> void:
		clear_awaiting_confirmation_calls += 1


class FakeHudFlow:
	var is_dead := false
	var can_respawn := false
	var hidden_for_match_over := false
	var is_game_over := false
	var set_alive_calls := 0
	var clear_dead_presentation_calls := 0

	func set_alive() -> void:
		set_alive_calls += 1

	func clear_dead_presentation() -> void:
		clear_dead_presentation_calls += 1
		is_dead = false
		can_respawn = false

	func has_dead_presentation() -> bool:
		return is_dead or can_respawn


class FakeMatchEndFlow:
	var handle_alive_restored_calls := 0

	func has_stale_dead_presentation() -> bool:
		return false

	func handle_alive_restored() -> void:
		handle_alive_restored_calls += 1


class FakePlayer:
	pass


class FakeWorldLaneState:
	var ships := {}


class FakeSessionLaneState:
	var player_lifecycle := {}


func _make_flow(
	world_sync,
	respawn_flow,
	hud_flow,
	match_end_flow,
	player
) -> GameplayAliveRestoreFlow:
	var flow := GameplayAliveRestoreFlow.new()
	flow.configure(world_sync, respawn_flow, hud_flow, match_end_flow, player)
	return flow


func _state() -> Dictionary:
	return {
		"self_id": "player-1",
		"world": {
			"ships": {
				"player-1": {}
			}
		},
		"session": {
			"player_lifecycle": {
				"player-1": "active",
			},
		},
	}


func test_apply_state_rejects_when_respawn_flow_says_not_ready() -> void:
	var world_sync := FakeWorldSync.new()
	var respawn_flow := FakeRespawnFlow.new()
	var hud_flow := FakeHudFlow.new()
	var match_end_flow := FakeMatchEndFlow.new()
	var player := FakePlayer.new()
	var flow := _make_flow(world_sync, respawn_flow, hud_flow, match_end_flow, player)

	flow.apply_state(_state())

	assert_eq(respawn_flow.should_restore_calls, 1)
	assert_eq(world_sync.clear_view_target_player_calls, 0)
	assert_eq(hud_flow.set_alive_calls, 0)
	assert_eq(match_end_flow.handle_alive_restored_calls, 0)
	assert_eq(respawn_flow.clear_awaiting_confirmation_calls, 0)


func test_apply_state_restores_alive_when_respawn_flow_allows_it() -> void:
	var world_sync := FakeWorldSync.new()
	var respawn_flow := FakeRespawnFlow.new()
	respawn_flow.should_restore_result = true
	var hud_flow := FakeHudFlow.new()
	var match_end_flow := FakeMatchEndFlow.new()
	var player := FakePlayer.new()
	var flow := _make_flow(world_sync, respawn_flow, hud_flow, match_end_flow, player)

	flow.apply_state(_state())

	assert_eq(respawn_flow.should_restore_calls, 1)
	assert_eq(world_sync.clear_view_target_player_calls, 1)
	assert_eq(hud_flow.set_alive_calls, 1)
	assert_eq(match_end_flow.handle_alive_restored_calls, 1)
	assert_eq(respawn_flow.clear_awaiting_confirmation_calls, 1)


func test_apply_state_without_match_end_flow_still_restores_alive() -> void:
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


func test_apply_lane_state_restores_dead_hud_when_confirmation_awaiting_and_self_active() -> void:
	var world_sync := FakeWorldSync.new()
	var respawn_flow := FakeRespawnFlow.new()
	respawn_flow.should_restore_result = true
	respawn_flow.awaiting_respawn_confirmation = true
	var hud_flow := FakeHudFlow.new()
	hud_flow.is_dead = true
	hud_flow.can_respawn = true
	var match_end_flow := FakeMatchEndFlow.new()
	var player := FakePlayer.new()
	var flow := _make_flow(world_sync, respawn_flow, hud_flow, match_end_flow, player)

	var world_lane_state := FakeWorldLaneState.new()
	world_lane_state.ships = {"player-1": {}}
	var session_lane_state := FakeSessionLaneState.new()
	session_lane_state.player_lifecycle = {"player-1": "active"}
	flow.apply_lane_state(world_lane_state, session_lane_state, "player-1")

	assert_eq(respawn_flow.should_restore_calls, 1)
	assert_eq(world_sync.clear_view_target_player_calls, 1)
	assert_eq(hud_flow.clear_dead_presentation_calls, 1)
	assert_eq(hud_flow.set_alive_calls, 0)
	assert_eq(match_end_flow.handle_alive_restored_calls, 1)
	assert_eq(respawn_flow.clear_awaiting_confirmation_calls, 1)


func test_apply_lane_state_clears_stale_dead_presentation_without_confirmation_when_lane_alive() -> void:
	var world_sync := FakeWorldSync.new()
	var respawn_flow := FakeRespawnFlow.new()
	respawn_flow.should_restore_result = true
	var hud_flow := FakeHudFlow.new()
	hud_flow.is_dead = true
	var match_end_flow := FakeMatchEndFlow.new()
	var player := FakePlayer.new()
	var flow := _make_flow(world_sync, respawn_flow, hud_flow, match_end_flow, player)

	var world_lane_state := FakeWorldLaneState.new()
	world_lane_state.ships = {"player-1": {}}
	var session_lane_state := FakeSessionLaneState.new()
	session_lane_state.player_lifecycle = {"player-1": "active"}
	flow.apply_lane_state(world_lane_state, session_lane_state, "player-1")

	assert_eq(respawn_flow.should_restore_calls, 1)
	assert_eq(world_sync.clear_view_target_player_calls, 1)
	assert_eq(hud_flow.clear_dead_presentation_calls, 1)
	assert_eq(match_end_flow.handle_alive_restored_calls, 1)
	assert_eq(respawn_flow.clear_awaiting_confirmation_calls, 1)


func test_apply_lane_state_does_not_clear_dead_presentation_when_hidden_for_match_over_or_game_over() -> void:
	var world_sync := FakeWorldSync.new()
	var respawn_flow := FakeRespawnFlow.new()
	respawn_flow.should_restore_result = true
	var hud_flow := FakeHudFlow.new()
	hud_flow.is_dead = true
	hud_flow.hidden_for_match_over = true
	hud_flow.is_game_over = true
	var match_end_flow := FakeMatchEndFlow.new()
	var player := FakePlayer.new()
	var flow := _make_flow(world_sync, respawn_flow, hud_flow, match_end_flow, player)

	var world_lane_state := FakeWorldLaneState.new()
	world_lane_state.ships = {"player-1": {}}
	var session_lane_state := FakeSessionLaneState.new()
	session_lane_state.player_lifecycle = {"player-1": "active"}
	flow.apply_lane_state(world_lane_state, session_lane_state, "player-1")

	assert_eq(respawn_flow.should_restore_calls, 0)
	assert_eq(world_sync.clear_view_target_player_calls, 0)
	assert_eq(hud_flow.clear_dead_presentation_calls, 0)
	assert_eq(match_end_flow.handle_alive_restored_calls, 0)
	assert_eq(respawn_flow.clear_awaiting_confirmation_calls, 0)


func test_apply_lane_state_with_real_respawn_flow_supports_state_records() -> void:
	var world_sync := FakeWorldSync.new()
	var hud_flow := FakeHudFlow.new()
	hud_flow.is_dead = true
	hud_flow.can_respawn = true
	var respawn_flow := GameplayRespawnFlow.new()
	respawn_flow.configure(null, hud_flow)
	respawn_flow.mark_awaiting_confirmation()
	var flow := _make_flow(world_sync, respawn_flow, hud_flow, null, FakePlayer.new())

	var world_lane_state := FakeWorldLaneState.new()
	world_lane_state.ships = {"player-1": {"id": "player-1"}}
	var session_lane_state := FakeSessionLaneState.new()
	session_lane_state.player_lifecycle = {"player-1": {"state": "active"}}
	flow.apply_lane_state(world_lane_state, session_lane_state, "player-1")

	assert_eq(hud_flow.clear_dead_presentation_calls, 1)
	assert_eq(world_sync.clear_view_target_player_calls, 1)
	assert_eq(respawn_flow.is_awaiting_confirmation(), false)


func test_apply_lane_state_with_real_respawn_flow_supports_status_records() -> void:
	var world_sync := FakeWorldSync.new()
	var hud_flow := FakeHudFlow.new()
	hud_flow.is_dead = true
	hud_flow.can_respawn = true
	var respawn_flow := GameplayRespawnFlow.new()
	respawn_flow.configure(null, hud_flow)
	respawn_flow.mark_awaiting_confirmation()
	var flow := _make_flow(world_sync, respawn_flow, hud_flow, null, FakePlayer.new())

	var world_lane_state := FakeWorldLaneState.new()
	world_lane_state.ships = {"player-1": {"id": "player-1"}}
	var session_lane_state := FakeSessionLaneState.new()
	session_lane_state.player_lifecycle = {"player-1": {"status": "active"}}
	flow.apply_lane_state(world_lane_state, session_lane_state, "player-1")

	assert_eq(hud_flow.clear_dead_presentation_calls, 1)
	assert_eq(world_sync.clear_view_target_player_calls, 1)
	assert_eq(respawn_flow.is_awaiting_confirmation(), false)
