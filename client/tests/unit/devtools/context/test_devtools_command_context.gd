extends GutTest

const DevtoolsCommandContext := preload("res://scripts/devtools/context/devtools_command_context.gd")


class FakeDebugFlow:
	var calls: Array = []

	func process(has_received_state: bool) -> void:
		calls.append(has_received_state)


class FakeStateContext:
	var gameplay_state := false

	func has_gameplay_state() -> bool:
		return gameplay_state


func test_process_delegates_to_debug_flow() -> void:
	var debug_flow := FakeDebugFlow.new()
	var state_context := FakeStateContext.new()
	var context := DevtoolsCommandContext.new()
	context.configure(debug_flow, state_context)

	context.process(true)

	assert_eq(debug_flow.calls.size(), 1)
	assert_true(debug_flow.calls[0])
