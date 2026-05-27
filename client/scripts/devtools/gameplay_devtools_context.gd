extends RefCounted
class_name GameplayDevtoolsContext


var debug_flow


func configure(connection_service_ref) -> void:
	debug_flow = GameplayDebugFlow.new()
	debug_flow.configure(connection_service_ref)


func reset() -> void:
	if debug_flow != null:
		debug_flow.reset()


func process(has_received_state: bool) -> void:
	if debug_flow != null:
		debug_flow.process(has_received_state)
