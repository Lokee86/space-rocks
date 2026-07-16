extends GutTest

const DevtoolsPlacementContext := preload("res://scripts/devtools/context/devtools_placement_context.gd")
const ClientOperationTrace := preload("res://scripts/observability/client_operation_trace.gd")


class FakeDevConnectionService:
	var configured := true
	var sent_spawn_results: Array = []

	func is_configured() -> bool:
		return configured

	func send_spawn_from_placement_result(result: Dictionary, operation_trace = null) -> void:
		sent_spawn_results.append({"result": result, "operation_trace": operation_trace})


class FakeRoute:
	var calls: Array = []

	func record_call(action_name: StringName, placement_context: Dictionary = {}) -> void:
		calls.append({
			"action_name": action_name,
			"placement_context": placement_context,
		})


class FakeStateContext:
	var lane_baseline_sync := false

	func has_lane_baseline_sync() -> bool:
		return lane_baseline_sync


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
	state_context.lane_baseline_sync = true
	var dev_connection_service := FakeDevConnectionService.new()
	var context := DevtoolsPlacementContext.new()
	context.configure(state_context, dev_connection_service)

	context.request_placement_action(&"spawn_player", {})

	assert_true(true)


func test_request_placement_action_carries_trace_through_route_context() -> void:
	var state_context := FakeStateContext.new()
	state_context.lane_baseline_sync = true
	var dev_connection_service := FakeDevConnectionService.new()
	var context := DevtoolsPlacementContext.new()
	var trace_id := "00000000-0000-4000-8000-000000000703"
	context.configure(
		state_context,
		dev_connection_service,
		func(operation_name: String):
			return ClientOperationTrace.new(operation_name, func() -> String: return trace_id)
	)
	var route := FakeRoute.new()
	context.configure_placement_request_route(Callable(route, "record_call"))

	context.request_placement_action(&"spawn_player", {})

	assert_eq(route.calls.size(), 1)
	assert_eq(route.calls[0]["placement_context"]["_operation_trace"].trace_id(), trace_id)


func test_handle_placement_result_ignores_empty_result() -> void:
	var state_context := FakeStateContext.new()
	var dev_connection_service := FakeDevConnectionService.new()
	var context := DevtoolsPlacementContext.new()
	context.configure(state_context, dev_connection_service)

	context.handle_placement_result({})

	assert_eq(dev_connection_service.sent_spawn_results.size(), 0)
