extends RefCounted
class_name DevtoolsPlacementContext

var state_context
var dev_connection_service
var placement_request_route: Callable


func configure(state_context_ref, dev_connection_service_ref) -> void:
	state_context = state_context_ref
	dev_connection_service = dev_connection_service_ref


func configure_placement_request_route(route: Callable) -> void:
	placement_request_route = route


func request_placement_action(action_name: StringName, placement_context: Dictionary = {}) -> void:
	if state_context == null or !state_context.has_gameplay_state():
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
