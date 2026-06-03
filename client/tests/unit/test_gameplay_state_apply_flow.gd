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


func test_apply_state_sends_gameplay_state_to_input_context() -> void:
	var input_context := FakeInputContext.new()
	var devtools_context := FakeDevtoolsContext.new()
	var flow := GameplayStateApplyFlow.new()
	flow.configure(input_context, devtools_context, null, null, null)
	var state := {"phase": 8}

	var result := flow.apply_state(state, false)

	assert_eq(devtools_context.apply_gameplay_state_call_count, 1)
	assert_eq(devtools_context.received_state, state)
	assert_eq(input_context.mark_gameplay_state_received_call_count, 1)
	assert_true(result.has_received_state)
	assert_true(result.started_gameplay)
