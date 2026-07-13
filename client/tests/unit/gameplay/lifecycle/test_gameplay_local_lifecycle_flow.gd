extends GutTest

const GameplayLocalLifecycleFlow = preload("res://scripts/gameplay/lifecycle/gameplay_local_lifecycle_flow.gd")
const GameplayRespawnFlow = preload("res://scripts/gameplay/respawn/gameplay_respawn_flow.gd")


class FakeWorldSync extends WorldSync:
	var clear_view_target_player_calls := 0

	func clear_view_target_player() -> void:
		clear_view_target_player_calls += 1


class FakeRespawnFlow extends GameplayRespawnFlow:
	var should_restore_result := false
	var use_restore_policy := false
	var should_restore_calls := 0
	var last_world_ships: Dictionary = {}
	var last_player_lifecycle: Dictionary = {}
	var last_self_id := ""
	var last_player
	var last_has_stale_dead_presentation := false
	var clear_awaiting_confirmation_calls := 0

	func should_restore_alive_hud(world_ships: Dictionary, player_lifecycle: Dictionary, self_id: String, player: Player, has_stale_dead_presentation := false) -> bool:
		should_restore_calls += 1
		last_world_ships = world_ships
		last_player_lifecycle = player_lifecycle
		last_self_id = self_id
		last_player = player
		last_has_stale_dead_presentation = has_stale_dead_presentation
		if use_restore_policy:
			var lifecycle = player_lifecycle.get(self_id, {})
			var lifecycle_status := ""
			if lifecycle is Dictionary:
				lifecycle_status = str(lifecycle.get("status", lifecycle.get("state", "")))
			else:
				lifecycle_status = str(lifecycle)
			return has_stale_dead_presentation and lifecycle_status == "active" and world_ships.has(self_id)
		return should_restore_result

	func clear_awaiting_confirmation() -> void:
		clear_awaiting_confirmation_calls += 1


class FakeHudFlow extends GameplayHudFlow:
	var set_alive_calls := 0
	var clear_dead_presentation_calls := 0
	var set_dead_calls := 0
	var last_respawn_delay := -1.0
	var last_lives := -1

	func set_alive() -> void:
		set_alive_calls += 1

	func apply_lives(lives) -> void:
		last_lives = lives

	func clear_dead_presentation() -> void:
		clear_dead_presentation_calls += 1
		is_dead = false
		can_respawn = false

	func set_dead(respawn_delay: float) -> void:
		set_dead_calls += 1
		last_respawn_delay = respawn_delay
		is_dead = true
		can_respawn = respawn_delay <= 0.0


	func has_dead_presentation() -> bool:
		return is_dead or can_respawn


class FakeMatchEndFlow extends MatchEndFlow:
	var handle_alive_restored_calls := 0
	var local_player_eliminated_calls := 0
	var last_eliminated_lives := -1

	func has_stale_dead_presentation() -> bool:
		return false

	func handle_alive_restored() -> void:
		handle_alive_restored_calls += 1

	func handle_local_player_eliminated(lives: int) -> void:
		local_player_eliminated_calls += 1
		last_eliminated_lives = lives


class FakePlayer extends Player:
	var stop_transient_effects_calls := 0

	func stop_transient_effects() -> void:
		stop_transient_effects_calls += 1


class FakeWorldLaneState extends WorldLaneState:
	pass


class FakeSessionLaneState extends SessionLaneState:
	pass


func _make_flow(
	world_sync,
	respawn_flow,
	hud_flow,
	match_end_flow,
	player
) -> GameplayLocalLifecycleFlow:
	var flow := GameplayLocalLifecycleFlow.new()
	if player != null:
		autofree(player)
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


func test_apply_lane_state_clears_hud_and_respawn_confirmation_when_local_player_is_active_and_present() -> void:
	var world_sync := FakeWorldSync.new()
	var respawn_flow := FakeRespawnFlow.new()
	respawn_flow.use_restore_policy = true
	var hud_flow := FakeHudFlow.new()
	hud_flow.is_dead = true
	var player := FakePlayer.new()
	var flow := _make_flow(world_sync, respawn_flow, hud_flow, null, player)

	var world_lane_state := FakeWorldLaneState.new()
	world_lane_state.ships = {"player-1": {"id": "player-1"}}
	var session_lane_state := FakeSessionLaneState.new()
	session_lane_state.player_lifecycle = {"player-1": {"player_id": "player-1", "status": "active"}}

	flow.apply_lane_state(world_lane_state, session_lane_state, "player-1")

	assert_eq(hud_flow.clear_dead_presentation_calls, 1)
	assert_eq(respawn_flow.clear_awaiting_confirmation_calls, 1)
	assert_false(hud_flow.is_dead)


func test_apply_lane_state_does_not_clear_when_lifecycle_is_pending_respawn_and_ship_missing() -> void:
	var world_sync := FakeWorldSync.new()
	var respawn_flow := FakeRespawnFlow.new()
	respawn_flow.use_restore_policy = true
	var hud_flow := FakeHudFlow.new()
	hud_flow.is_dead = true
	var player := FakePlayer.new()
	var flow := _make_flow(world_sync, respawn_flow, hud_flow, null, player)

	var world_lane_state := FakeWorldLaneState.new()
	world_lane_state.ships = {}
	var session_lane_state := FakeSessionLaneState.new()
	session_lane_state.player_lifecycle = {"player-1": {"player_id": "player-1", "status": "pending_respawn"}}

	flow.apply_lane_state(world_lane_state, session_lane_state, "player-1")

	assert_eq(hud_flow.clear_dead_presentation_calls, 0)
	assert_eq(respawn_flow.clear_awaiting_confirmation_calls, 0)
	assert_true(hud_flow.is_dead)


func test_apply_lane_state_does_not_clear_when_lifecycle_is_active_but_self_ship_is_missing() -> void:
	var world_sync := FakeWorldSync.new()
	var respawn_flow := FakeRespawnFlow.new()
	respawn_flow.use_restore_policy = true
	var hud_flow := FakeHudFlow.new()
	hud_flow.is_dead = true
	var player := FakePlayer.new()
	var flow := _make_flow(world_sync, respawn_flow, hud_flow, null, player)

	var world_lane_state := FakeWorldLaneState.new()
	world_lane_state.ships = {}
	var session_lane_state := FakeSessionLaneState.new()
	session_lane_state.player_lifecycle = {"player-1": {"player_id": "player-1", "status": "active"}}

	flow.apply_lane_state(world_lane_state, session_lane_state, "player-1")

	assert_eq(hud_flow.clear_dead_presentation_calls, 0)
	assert_eq(respawn_flow.clear_awaiting_confirmation_calls, 0)
	assert_true(hud_flow.is_dead)


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

func _eliminated_state(lives := 0) -> Dictionary:
	return {"self_id": "player-1", "session": {"players": {"player-1": {"lives": lives}}, "player_lifecycle": {"player-1": "eliminated"}}}

func test_initial_eliminated_baseline_reconstructs_without_ship_death_event() -> void:
	var hud_flow := FakeHudFlow.new(); var match_end_flow := FakeMatchEndFlow.new(); var player := FakePlayer.new()
	var flow := _make_flow(null, null, hud_flow, match_end_flow, player); flow.apply_state(_eliminated_state(2))
	assert_eq(player.stop_transient_effects_calls, 1); assert_eq(hud_flow.last_lives, 2); assert_eq(match_end_flow.local_player_eliminated_calls, 1)

func test_authoritative_zero_lives_reach_match_end_flow() -> void:
	var hud_flow := FakeHudFlow.new(); var match_end_flow := FakeMatchEndFlow.new()
	var flow := _make_flow(null, null, hud_flow, match_end_flow, FakePlayer.new()); flow.apply_state(_eliminated_state(0))
	assert_eq(hud_flow.last_lives, 0); assert_eq(match_end_flow.last_eliminated_lives, 0)

func test_repeated_eliminated_fanout_does_not_repeat_lifecycle_flow_handoff() -> void:
	var hud_flow := FakeHudFlow.new(); var match_end_flow := FakeMatchEndFlow.new(); var player := FakePlayer.new()
	var flow := _make_flow(null, null, hud_flow, match_end_flow, player); flow.apply_state(_eliminated_state(0)); flow.apply_state(_eliminated_state(0))
	assert_eq(player.stop_transient_effects_calls, 1); assert_eq(match_end_flow.local_player_eliminated_calls, 1)

func test_active_then_eliminated_allows_new_transition_after_alive_restoration() -> void:
	var respawn_flow := FakeRespawnFlow.new(); respawn_flow.should_restore_result = true; var hud_flow := FakeHudFlow.new(); var match_end_flow := FakeMatchEndFlow.new()
	var flow := _make_flow(null, respawn_flow, hud_flow, match_end_flow, FakePlayer.new()); flow.apply_state(_eliminated_state(1)); flow.apply_state(_state()); flow.apply_state(_eliminated_state(0))
	assert_eq(match_end_flow.handle_alive_restored_calls, 1); assert_eq(match_end_flow.local_player_eliminated_calls, 2)


func _pending_state(cooldown: float, lives: int = 2) -> Array:
	var world_lane_state := FakeWorldLaneState.new()
	var session_lane_state := FakeSessionLaneState.new()
	session_lane_state.player_lifecycle = {"player-1": {"status": "pending_respawn"}}
	session_lane_state.player_sessions = {"player-1": {"lives": lives, "respawn_cooldown": cooldown}}
	return [world_lane_state, session_lane_state]


func test_pending_respawn_without_event_applies_lives_and_dead_presentation() -> void:
	var player := FakePlayer.new()
	var hud_flow := FakeHudFlow.new()
	var flow := _make_flow(null, FakeRespawnFlow.new(), hud_flow, null, player)
	var state := _pending_state(3.0)
	flow.apply_lane_state(state[0], state[1], "player-1")
	assert_eq(hud_flow.last_lives, 2)
	assert_eq(hud_flow.set_dead_calls, 1)
	assert_eq(hud_flow.last_respawn_delay, 3.0)
	assert_eq(player.stop_transient_effects_calls, 1)


func test_pending_respawn_zero_cooldown_is_immediately_available() -> void:
	var hud_flow := FakeHudFlow.new()
	var flow := _make_flow(null, FakeRespawnFlow.new(), hud_flow, null, FakePlayer.new())
	var state := _pending_state(0.0)
	flow.apply_lane_state(state[0], state[1], "player-1")
	assert_true(hud_flow.can_respawn)


func test_unchanged_pending_respawn_does_not_restart_countdown() -> void:
	var hud_flow := FakeHudFlow.new()
	var flow := _make_flow(null, FakeRespawnFlow.new(), hud_flow, null, FakePlayer.new())
	var state := _pending_state(3.0)
	flow.apply_lane_state(state[0], state[1], "player-1")
	flow.apply_lane_state(state[0], state[1], "player-1")
	assert_eq(hud_flow.set_dead_calls, 1)


func test_changed_pending_respawn_cooldown_refreshes_countdown() -> void:
	var hud_flow := FakeHudFlow.new()
	var flow := _make_flow(null, FakeRespawnFlow.new(), hud_flow, null, FakePlayer.new())
	var state := _pending_state(3.0)
	flow.apply_lane_state(state[0], state[1], "player-1")
	state = _pending_state(1.0)
	flow.apply_lane_state(state[0], state[1], "player-1")
	assert_eq(hud_flow.set_dead_calls, 2)
	assert_eq(hud_flow.last_respawn_delay, 1.0)


func test_reset_allows_pending_respawn_state_to_apply_again() -> void:
	var hud_flow := FakeHudFlow.new()
	var flow := _make_flow(null, FakeRespawnFlow.new(), hud_flow, null, FakePlayer.new())
	var state := _pending_state(3.0)
	flow.apply_lane_state(state[0], state[1], "player-1")
	flow.reset()
	flow.apply_lane_state(state[0], state[1], "player-1")
	assert_eq(hud_flow.set_dead_calls, 2)
