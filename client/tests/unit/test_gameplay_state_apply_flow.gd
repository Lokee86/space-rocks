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


class FakeRuntimeContext:
	var received_state: Dictionary = {}
	var has_received_state := false
	var apply_world_state_call_count := 0

	func apply_world_state(state: Dictionary, existing_has_received_state: bool) -> void:
		apply_world_state_call_count += 1
		received_state = state
		has_received_state = existing_has_received_state


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
	var runtime_context := FakeRuntimeContext.new()
	var alive_restore_flow := FakeAliveRestoreFlow.new()
	var event_lifecycle_flow := FakeEventLifecycleFlow.new()
	var flow := GameplayStateApplyFlow.new()
	flow.configure(
		input_context,
		devtools_context,
		hud_flow,
		runtime_context,
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
	assert_eq(runtime_context.apply_world_state_call_count, 1)
	assert_eq(runtime_context.received_state, state)
	assert_false(runtime_context.has_received_state)
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
	var runtime_context := FakeRuntimeContext.new()
	var alive_restore_flow := FakeAliveRestoreFlow.new()
	var event_lifecycle_flow := FakeEventLifecycleFlow.new()
	var flow := GameplayStateApplyFlow.new()
	flow.configure(
		input_context,
		devtools_context,
		hud_flow,
		runtime_context,
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
