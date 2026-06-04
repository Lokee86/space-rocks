extends GutTest

const DevtoolsPlacementContext := preload("res://scripts/devtools/context/devtools_placement_context.gd")


class FakeDevConnectionService:
	var configured := true
	var sent_spawn_results: Array = []

	func is_configured() -> bool:
		return configured

	func send_spawn_from_placement_result(result: Dictionary) -> void:
		sent_spawn_results.append(result)


class FakeRoute:
	var calls: Array = []

	func record_call(action_name: StringName, placement_context: Dictionary = {}) -> void:
		calls.append({
			"action_name": action_name,
			"placement_context": placement_context,
		})


class FakeStateContext:
	var gameplay_state := false

	func has_gameplay_state() -> bool:
		return gameplay_state


func test_request_placement_action_does_nothing_before_gameplay_state() -> void:
	var state_context := FakeStateContext.new()
	var dev_connection_service := FakeDevConnectionService.new()
	var context := DevtoolsPlacementContext.new()
	context.configure(state_context, dev_connection_service)
	var route := FakeRoute.new()
	context.configure_placement_request_route(Callable(route, "record_call"))

	context.request_placement_action(&"spawn_player", {})

	assert_eq(route.calls.size(), 0)


func test_request_placement_action_does_nothing_with_null_route() -> void:
	var state_context := FakeStateContext.new()
	state_context.gameplay_state = true
	var dev_connection_service := FakeDevConnectionService.new()
	var context := DevtoolsPlacementContext.new()
	context.configure(state_context, dev_connection_service)

	context.request_placement_action(&"spawn_player", {})

	assert_true(true)


func test_handle_placement_result_ignores_empty_result() -> void:
	var state_context := FakeStateContext.new()
	var dev_connection_service := FakeDevConnectionService.new()
	var context := DevtoolsPlacementContext.new()
	context.configure(state_context, dev_connection_service)

	context.handle_placement_result({})

	assert_eq(dev_connection_service.sent_spawn_results.size(), 0)
