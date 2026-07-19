extends RefCounted
class_name DevtoolsPlacementContext

const ClientLogger := preload("res://scripts/logging/logger.gd")

const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")


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
	var operation_trace := create_operation_trace("devtools.%s" % String(action_name))
	var routed_context := placement_context.duplicate(true)
	routed_context["_operation_trace"] = operation_trace
	placement_request_route.call(action_name, routed_context)


func handle_placement_result(result: Dictionary) -> void:
	if result.is_empty():
		return
	var action_name := StringName(result.get("action_name", StringName()))
	if action_name.is_empty():
		return
	var operation_trace = result.get("_operation_trace", null)
	if not operation_trace is ClientOperationTrace:
		operation_trace = create_operation_trace("devtools.%s" % String(action_name))
	var packet: Dictionary = DevSpawnPacketBuilder.build_from_placement_result(result)
	if packet.is_empty():
		_reject(operation_trace, _command_type_for_action(action_name), "placement_result_invalid")
		return
	if dev_connection_service == null || !dev_connection_service.is_configured():
		_emit_dependency_unavailable("dev_connection_service")
		return
	dev_connection_service.send_spawn_from_placement_result(result, operation_trace)


func _reject(operation_trace: ClientOperationTrace, command_type: String, reason: String) -> void:
	ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_DEVTOOLS_COMMAND_REJECTED,
		"",
		{"trace_id": operation_trace.trace_id()},
		{"command_type": command_type, "reason": reason}
	)


func _command_type_for_action(action_name: StringName) -> String:
	match String(action_name):
		"spawn_pickup":
			return "debug_spawn_pickup"
		"continuous_spawn_bullet":
			return "debug_begin_continuous_bullet_stream"
		_:
			return "debug_spawn_entity"


func _emit_dependency_unavailable(dependency: String) -> void:
	ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_DEPENDENCY_UNAVAILABLE,
		"",
		{},
		{
			"subsystem": "devtools",
			"dependency": dependency,
			"failure_mode": "not_configured",
		}
	)
