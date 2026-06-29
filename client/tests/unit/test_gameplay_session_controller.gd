extends GutTest

const GameplaySessionController := preload("res://scripts/session/gameplay_session_controller.gd")


class FakeConnectionService:
	extends Node

	var close_calls := 0
	var events: Array[String] = []

	func close_gracefully() -> void:
		close_calls += 1
		events.append("close_gracefully_started")
		await get_tree().process_frame
		events.append("close_gracefully_finished")


class FakeSessionContext:
	extends RefCounted

	var clear_calls := 0

	func clear() -> void:
		clear_calls += 1


class FakeShellBootFlow:
	extends RefCounted

	var clear_calls := 0

	func clear() -> void:
		clear_calls += 1


class ReplayProbe:
	extends RefCounted

	var events: Array[String] = []

	func mark_replay_requested() -> void:
		events.append("replay_requested")


class FakeHudFlow:
	var hidden_for_match_over := false
	var is_game_over := false
	var dead_presentation := false
	var clear_calls := 0

	func _has_dead_presentation() -> bool:
		return dead_presentation

	func clear_dead_presentation() -> void:
		clear_calls += 1
		dead_presentation = false


class FakeRespawnFlow:
	var awaiting_confirmation := false
	var clear_calls := 0

	func clear_awaiting_confirmation() -> void:
		clear_calls += 1
		awaiting_confirmation = false


class FakeRuntimeContext:
	var respawn_flow


class FakeGameplayShellFlow:
	var runtime_context


class FakeGameplayComposition:
	var gameplay_shell_flow


func test_replay_waits_for_graceful_close_before_emitting_replay_requested() -> void:
	var controller := GameplaySessionController.new()
	var connection_service := FakeConnectionService.new()
	var session_context := FakeSessionContext.new()
	var shell_boot_flow := FakeShellBootFlow.new()
	var replay_probe := ReplayProbe.new()

	add_child_autofree(connection_service)
	add_child_autofree(controller)
	controller.connection_service = connection_service
	controller.session_context = session_context
	controller.shell_boot_flow = shell_boot_flow
	controller.logger = Callable()
	connection_service.events = replay_probe.events
	controller.replay_requested.connect(Callable(replay_probe, "mark_replay_requested"))

	await controller._on_gameplay_replay_requested()

	assert_eq(connection_service.close_calls, 1)
	assert_eq(session_context.clear_calls, 1)
	assert_eq(shell_boot_flow.clear_calls, 1)
	assert_eq(
		replay_probe.events,
		["close_gracefully_started", "close_gracefully_finished", "replay_requested"]
	)


func test_clear_stale_dead_presentation_from_lane_state_clears_dead_overlay_when_local_player_is_active_and_present() -> void:
	var controller := GameplaySessionController.new()
	var hud_flow := FakeHudFlow.new()
	hud_flow.dead_presentation = true
	var respawn_flow := FakeRespawnFlow.new()
	var runtime_context := FakeRuntimeContext.new()
	runtime_context.respawn_flow = respawn_flow
	var gameplay_shell_flow := FakeGameplayShellFlow.new()
	gameplay_shell_flow.runtime_context = runtime_context
	var gameplay_composition := FakeGameplayComposition.new()
	gameplay_composition.gameplay_shell_flow = gameplay_shell_flow
	controller.gameplay_composition = gameplay_composition
	controller.gameplay_realtime_router = {
		"overlay_lane_state": {"self_id": "player-1"},
		"session_lane_state": {
			"player_lifecycle": {
				"player-1": {"player_id": "player-1", "status": "active"},
			},
		},
		"world_lane_state": {
			"ships": {
				"player-1": {"id": "player-1"},
			},
		},
	}

	controller.call("_clear_stale_dead_presentation_from_lane_state", hud_flow)

	assert_eq(hud_flow.clear_calls, 1)
	assert_eq(respawn_flow.clear_calls, 1)
	assert_false(hud_flow.dead_presentation)


func test_clear_stale_dead_presentation_from_lane_state_does_not_clear_without_active_lifecycle_and_ship() -> void:
	var controller := GameplaySessionController.new()
	var hud_flow := FakeHudFlow.new()
	hud_flow.dead_presentation = true
	var respawn_flow := FakeRespawnFlow.new()
	var runtime_context := FakeRuntimeContext.new()
	runtime_context.respawn_flow = respawn_flow
	var gameplay_shell_flow := FakeGameplayShellFlow.new()
	gameplay_shell_flow.runtime_context = runtime_context
	var gameplay_composition := FakeGameplayComposition.new()
	gameplay_composition.gameplay_shell_flow = gameplay_shell_flow
	controller.gameplay_composition = gameplay_composition
	controller.gameplay_realtime_router = {
		"overlay_lane_state": {"self_id": "player-1"},
		"session_lane_state": {
			"player_lifecycle": {
				"player-1": {"player_id": "player-1", "status": "pending_respawn"},
			},
		},
		"world_lane_state": {
			"ships": {},
		},
	}

	controller.call("_clear_stale_dead_presentation_from_lane_state", hud_flow)

	assert_eq(hud_flow.clear_calls, 0)
	assert_eq(respawn_flow.clear_calls, 0)
	assert_true(hud_flow.dead_presentation)

func test_clear_stale_dead_presentation_from_lane_state_does_not_clear_when_lifecycle_is_active_but_self_ship_is_missing() -> void:
	var controller := GameplaySessionController.new()
	var hud_flow := FakeHudFlow.new()
	hud_flow.dead_presentation = true
	var respawn_flow := FakeRespawnFlow.new()
	var runtime_context := FakeRuntimeContext.new()
	runtime_context.respawn_flow = respawn_flow
	var gameplay_shell_flow := FakeGameplayShellFlow.new()
	gameplay_shell_flow.runtime_context = runtime_context
	var gameplay_composition := FakeGameplayComposition.new()
	gameplay_composition.gameplay_shell_flow = gameplay_shell_flow
	controller.gameplay_composition = gameplay_composition
	controller.gameplay_realtime_router = {
		"overlay_lane_state": {"self_id": "player-1"},
		"session_lane_state": {
			"player_lifecycle": {
				"player-1": {"player_id": "player-1", "status": "active"},
			},
		},
		"world_lane_state": {
			"ships": {},
		},
	}

	controller.call("_clear_stale_dead_presentation_from_lane_state", hud_flow)

	assert_eq(hud_flow.clear_calls, 0)
	assert_eq(respawn_flow.clear_calls, 0)
	assert_true(hud_flow.dead_presentation)
