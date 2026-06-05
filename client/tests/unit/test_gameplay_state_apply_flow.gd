extends GutTest

const GameplayStateApplyFlow = preload("res://scripts/gameplay/state/gameplay_state_apply_flow.gd")


class FakeInputContext:
	var mark_gameplay_state_received_call_count := 0

	func mark_gameplay_state_received() -> void:
		mark_gameplay_state_received_call_count += 1


class FakeDevtoolsContext:
	var received_state: Dictionary = {}
	var apply_gameplay_state_call_count := 0

	func apply_gameplay_state(state: Dictionary) -> void:
		apply_gameplay_state_call_count += 1
		received_state = state


class FakeHudFlow:
	var received_state: Dictionary = {}
	var apply_gameplay_state_summary_call_count := 0

	func apply_gameplay_state_summary(state: Dictionary) -> void:
		apply_gameplay_state_summary_call_count += 1
		received_state = state


class FakeWorldSync:
	var apply_state_call_count := 0
	var received_self_id := ""
	var received_players: Dictionary = {}
	var received_bullets: Dictionary = {}
	var received_asteroids: Dictionary = {}
	var received_pickups: Dictionary = {}

	func apply_state(
		self_id: String,
		server_players: Dictionary,
		server_bullets: Dictionary,
		server_asteroids: Dictionary,
		server_pickups: Dictionary
	) -> void:
		apply_state_call_count += 1
		received_self_id = self_id
		received_players = server_players
		received_bullets = server_bullets
		received_asteroids = server_asteroids
		received_pickups = server_pickups


class FakeAliveRestoreFlow:
	var received_state: Dictionary = {}
	var apply_state_call_count := 0

	func apply_state(state: Dictionary) -> void:
		apply_state_call_count += 1
		received_state = state


class FakeEventLifecycleFlow:
	var received_state: Dictionary = {}
	var apply_server_events_call_count := 0

	func apply_server_events(state: Dictionary) -> void:
		apply_server_events_call_count += 1
		received_state = state


func test_apply_state_delegates_to_new_seams_on_first_state() -> void:
	var input_context := FakeInputContext.new()
	var devtools_context := FakeDevtoolsContext.new()
	var hud_flow := FakeHudFlow.new()
	var world_sync := FakeWorldSync.new()
	var alive_restore_flow := FakeAliveRestoreFlow.new()
	var event_lifecycle_flow := FakeEventLifecycleFlow.new()
	var flow := GameplayStateApplyFlow.new()
	flow.configure(
		input_context,
		devtools_context,
		hud_flow,
		world_sync,
		event_lifecycle_flow,
		alive_restore_flow
	)
	var state := {
		"phase": 8,
		"self_id": "player-1",
		"server_events": [{"type": "test_event"}],
	}

	var result := flow.apply_state(state, false)

	assert_eq(devtools_context.apply_gameplay_state_call_count, 1)
	assert_eq(devtools_context.received_state, state)
	assert_eq(input_context.mark_gameplay_state_received_call_count, 1)
	assert_eq(hud_flow.apply_gameplay_state_summary_call_count, 1)
	assert_eq(hud_flow.received_state, state)
	assert_eq(world_sync.apply_state_call_count, 1)
	assert_eq(world_sync.received_self_id, "player-1")
	assert_eq(world_sync.received_players, {})
	assert_eq(world_sync.received_bullets, {})
	assert_eq(world_sync.received_asteroids, {})
	assert_eq(world_sync.received_pickups, {})
	assert_eq(alive_restore_flow.apply_state_call_count, 1)
	assert_eq(alive_restore_flow.received_state, state)
	assert_eq(event_lifecycle_flow.apply_server_events_call_count, 1)
	assert_eq(event_lifecycle_flow.received_state, state)
	assert_true(result.has_received_state)
	assert_true(result.started_gameplay)


func test_apply_state_reports_not_first_gameplay_state_after_initial_state() -> void:
	var input_context := FakeInputContext.new()
	var devtools_context := FakeDevtoolsContext.new()
	var hud_flow := FakeHudFlow.new()
	var world_sync := FakeWorldSync.new()
	var alive_restore_flow := FakeAliveRestoreFlow.new()
	var event_lifecycle_flow := FakeEventLifecycleFlow.new()
	var flow := GameplayStateApplyFlow.new()
	flow.configure(
		input_context,
		devtools_context,
		hud_flow,
		world_sync,
		event_lifecycle_flow,
		alive_restore_flow
	)
	var state := {
		"phase": 8,
		"self_id": "player-1",
		"server_events": [],
	}

	var result := flow.apply_state(state, true)

	assert_false(result.started_gameplay)
