extends RefCounted
class_name GameplayDevtoolsContext


var debug_flow
var devtools_window_controller


func configure(connection_service_ref) -> void:
	debug_flow = GameplayDebugFlow.new()
	debug_flow.configure(connection_service_ref)
	devtools_window_controller = DevtoolsWindowController.new()


func reset() -> void:
	if debug_flow != null:
		debug_flow.reset()


func process(has_received_state: bool) -> void:
	if Input.is_action_just_pressed("DevToggle0"):
		toggle_devtools_window()
	if debug_flow != null:
		debug_flow.process(has_received_state)


func toggle_devtools_window() -> void:
	if devtools_window_controller != null:
		devtools_window_controller.toggle_window()


func apply_debug_status(status: Dictionary) -> void:
	if devtools_window_controller != null:
		devtools_window_controller.apply_debug_status(status)
