extends RefCounted
class_name GameplayDevtoolsContext


var debug_flow
var devtools_window_controller
var has_received_gameplay_state := false


func configure(connection_service_ref) -> void:
	debug_flow = GameplayDebugFlow.new()
	debug_flow.configure(connection_service_ref)
	devtools_window_controller = DevtoolsWindowController.new()
	_connect_window_controller_signals()


func reset() -> void:
	if debug_flow != null:
		debug_flow.reset()


func process(has_received_state: bool) -> void:
	has_received_gameplay_state = has_received_state
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


func _connect_window_controller_signals() -> void:
	if !devtools_window_controller.toggle_invincible_requested.is_connected(request_toggle_invincible):
		devtools_window_controller.toggle_invincible_requested.connect(request_toggle_invincible)
	if !devtools_window_controller.toggle_infinite_lives_requested.is_connected(request_toggle_infinite_lives):
		devtools_window_controller.toggle_infinite_lives_requested.connect(request_toggle_infinite_lives)
	if !devtools_window_controller.toggle_freeze_world_requested.is_connected(request_toggle_freeze_world):
		devtools_window_controller.toggle_freeze_world_requested.connect(request_toggle_freeze_world)
	if !devtools_window_controller.toggle_freeze_player_requested.is_connected(request_toggle_freeze_player):
		devtools_window_controller.toggle_freeze_player_requested.connect(request_toggle_freeze_player)


func request_toggle_invincible() -> void:
	if !has_received_gameplay_state || debug_flow == null:
		return
	debug_flow.toggle_invincible()


func request_toggle_infinite_lives() -> void:
	if !has_received_gameplay_state || debug_flow == null:
		return
	debug_flow.toggle_infinite_lives()


func request_toggle_freeze_world() -> void:
	if !has_received_gameplay_state || debug_flow == null:
		return
	debug_flow.toggle_freeze_world()


func request_toggle_freeze_player() -> void:
	if !has_received_gameplay_state || debug_flow == null:
		return
	debug_flow.toggle_freeze_player()
