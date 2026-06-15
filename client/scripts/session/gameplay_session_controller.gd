extends Node

var connection_service
var hud: Control
var gameplay_user_interface: Control
var main_menu: Control
var session_context
var shell_boot_flow
var logger: Callable

var gameplay_composition
var gameplay_state_flow
var accepts_gameplay_packets := false

signal return_to_pregame_requested(session_mode: String)
signal replay_requested


func configure(
	connection_service_ref,
	scene_root_ref: Node,
	player_ref,
	view_anchor_ref,
	bullets_ref: Node2D,
	asteroids_ref: Node2D,
	pickups_ref: Node2D,
	hud_ref: Control,
	gameplay_user_interface_ref: Control,
	main_menu_ref: Control,
	session_context_ref,
	shell_boot_flow_ref,
	logger_callable: Callable
) -> void:
	connection_service = connection_service_ref
	hud = hud_ref
	gameplay_user_interface = gameplay_user_interface_ref
	main_menu = main_menu_ref
	session_context = session_context_ref
	shell_boot_flow = shell_boot_flow_ref
	logger = logger_callable

	gameplay_composition = GameplayComposition.new()
	gameplay_composition.configure(
		connection_service,
		scene_root_ref,
		player_ref,
		view_anchor_ref,
		bullets_ref,
		asteroids_ref,
		pickups_ref,
		hud,
		gameplay_user_interface,
		session_context,
		logger
	)
	gameplay_composition.gameplay_started.connect(_on_gameplay_started)
	gameplay_composition.quit_to_main_menu_requested.connect(_on_gameplay_quit_to_main_menu_requested)
	gameplay_composition.replay_requested.connect(_on_gameplay_replay_requested)
	gameplay_composition.return_to_pregame_requested.connect(_on_gameplay_return_to_pregame_requested)
	gameplay_composition.return_to_lobby_requested.connect(_on_gameplay_return_to_lobby_requested)
	gameplay_state_flow = GameplayStateFlow.new()
	gameplay_state_flow.configure(gameplay_composition)


func handle_gameplay_state(packet: Dictionary) -> void:
	if !accepts_gameplay_packets:
		return
	if gameplay_state_flow != null:
		gameplay_state_flow.handle_gameplay_state_packet(packet)


func handle_player_pause_state(packet: Dictionary) -> void:
	if !accepts_gameplay_packets:
		return
	if gameplay_composition != null:
		gameplay_composition.apply_player_pause_state_packet(packet)


func handle_debug_status_packet(packet: Dictionary) -> void:
	if gameplay_composition != null:
		gameplay_composition.apply_devtools_debug_status_packet(packet)


func handle_debug_shape_catalog_packet(packet: Dictionary) -> void:
	if gameplay_composition != null:
		gameplay_composition.apply_debug_shape_catalog_packet(packet)


func begin_accepting_gameplay_packets() -> void:
	accepts_gameplay_packets = true


func _process(delta: float) -> void:
	if gameplay_composition != null:
		var has_received_state := false
		if gameplay_state_flow != null:
			has_received_state = gameplay_state_flow.has_received_state()
		gameplay_composition.process(delta, has_received_state)


func _input(event: InputEvent) -> void:
	if gameplay_composition != null and gameplay_composition.handle_devtools_input(event):
		get_viewport().set_input_as_handled()
		return

	var hud_input_policy = get_node_or_null("/root/HudInputPolicy")
	if hud_input_policy != null:
		if hud_input_policy.has_method("should_gameplay_ui_receive_mouse_event"):
			if hud_input_policy.should_gameplay_ui_receive_mouse_event(event, gameplay_user_interface, get_viewport()):
				return
		elif hud_input_policy.should_hud_receive_mouse_event(event, hud, get_viewport()):
			return

	if gameplay_composition == null:
		return
	if gameplay_composition.handle_gameplay_input(event):
		get_viewport().set_input_as_handled()


func _unhandled_input(event: InputEvent) -> void:
	if gameplay_composition == null:
		return
	if gameplay_composition.handle_gameplay_input(event):
		get_viewport().set_input_as_handled()


func reset() -> void:
	accepts_gameplay_packets = false
	if gameplay_state_flow != null:
		gameplay_state_flow.reset()
	if gameplay_composition != null:
		gameplay_composition.reset()


func configure_room_state_provider(provider: Callable) -> void:
	if gameplay_composition != null:
		gameplay_composition.configure_room_state_provider(provider)


func configure_match_result_provider(provider: Callable) -> void:
	if gameplay_composition != null:
		gameplay_composition.configure_match_result_provider(provider)


func configure_room_max_players_provider(provider: Callable) -> void:
	if gameplay_composition != null:
		gameplay_composition.configure_room_max_players_provider(provider)


func refresh_match_end_state() -> void:
	if gameplay_composition != null:
		gameplay_composition.refresh_match_end_state()


func refresh_game_over_menu_state() -> void:
	refresh_match_end_state()


func _on_gameplay_started() -> void:
	if main_menu != null:
		main_menu.hide()


func _on_gameplay_quit_to_main_menu_requested() -> void:
	_log("Gameplay quit to main menu requested")
	if connection_service != null:
		connection_service.begin_graceful_close()
	reset()
	if session_context != null:
		session_context.clear()
	if shell_boot_flow != null:
		shell_boot_flow.clear()
	if main_menu != null:
		main_menu.show()


func _on_gameplay_return_to_lobby_requested() -> void:
	_log("Gameplay return to lobby requested")
	if connection_service != null:
		connection_service.send_return_to_lobby_request()
	reset()


func _on_gameplay_return_to_pregame_requested(session_mode: String) -> void:
	_log("Gameplay return to pregame requested: %s" % session_mode)
	if connection_service != null && connection_service.has_method("begin_graceful_close"):
		connection_service.begin_graceful_close()
	reset()
	if session_context != null:
		session_context.clear()
	if shell_boot_flow != null:
		shell_boot_flow.clear()
	return_to_pregame_requested.emit(session_mode)


func _on_gameplay_replay_requested() -> void:
	_log("Gameplay replay requested")
	if connection_service != null && connection_service.has_method("close_gracefully"):
		await connection_service.close_gracefully()
	reset()
	if session_context != null:
		session_context.clear()
	if shell_boot_flow != null:
		shell_boot_flow.clear()
	replay_requested.emit()


func _log(message: String) -> void:
	if !logger.is_null():
		logger.call(message)
