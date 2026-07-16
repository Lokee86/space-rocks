extends RefCounted
class_name DevtoolsPlacementContext

const ClientOperationTrace := preload("res://scripts/observability/client_operation_trace.gd")

var state_context
var dev_connection_service
var placement_request_route: Callable
var operation_trace_factory: Callable


func configure(
	state_context_ref,
	dev_connection_service_ref,
	operation_trace_factory_ref: Callable = Callable()
) -> void:
	state_context = state_context_ref
	dev_connection_service = dev_connection_service_ref
	operation_trace_factory = operation_trace_factory_ref


func create_operation_trace(action_name: String) -> ClientOperationTrace:
	return ClientOperationTrace.create(action_name, operation_trace_factory)


func configure_placement_request_route(route: Callable) -> void:
	placement_request_route = route


func request_placement_action(action_name: StringName, placement_context: Dictionary = {}) -> void:
	if state_context == null or !state_context.has_lane_baseline_sync():
		return
	if placement_request_route.is_null():
		return
	placement_request_route.call(action_name, placement_context)


func handle_placement_result(result: Dictionary) -> void:
	if result.is_empty():
		return
	var action_name := StringName(result.get("action_name", StringName()))
	if action_name.is_empty():
		return
	if dev_connection_service == null || !dev_connection_service.is_configured():
		return
	dev_connection_service.send_spawn_from_placement_result(result)
