extends Node


const SpectateMenuState := preload("res://scripts/gameplay/spectate/spectate_menu_state.gd")
const GameplayStatePacketReader := preload("res://scripts/gameplay/state/gameplay_state_packet_reader.gd")
const DevToolsSessionFlow := preload("res://scripts/devtools/dev_tools_session_flow.gd")

var connection_service
var scene_root: Node
var player
var bullets: Node2D
var asteroids: Node2D
var hud: Control
var main_menu: Control
var session_context
var shell_boot_flow
var logger: Callable
var room_max_players_provider: Callable

var gameplay_shell_flow
var gameplay_presentation_flow
var gameplay_hud_flow
var gameplay_menu_flow
var dev_tools_session_flow
var spectate_menu_state
var has_received_gameplay_state := false
var accepts_gameplay_packets := false


func configure(
	connection_service_ref,
	scene_root_ref: Node,
	player_ref,
	bullets_ref: Node2D,
	asteroids_ref: Node2D,
	hud_ref: Control,
	main_menu_ref: Control,
	session_context_ref,
	shell_boot_flow_ref,
	logger_callable: Callable
) -> void:
	connection_service = connection_service_ref
	scene_root = scene_root_ref
	player = player_ref
	bullets = bullets_ref
	asteroids = asteroids_ref
	hud = hud_ref
	main_menu = main_menu_ref
	session_context = session_context_ref
	shell_boot_flow = shell_boot_flow_ref
	logger = logger_callable

	gameplay_hud_flow = GameplayHudFlow.new()
	gameplay_hud_flow.configure(hud)
	gameplay_menu_flow = GameplayMenuFlow.new()
	gameplay_menu_flow.configure(hud, connection_service, player, session_context)
	spectate_menu_state = SpectateMenuState.new()
	gameplay_menu_flow.configure_spectate_menu_state(spectate_menu_state)
	gameplay_shell_flow = GameplayShellFlow.new()
	gameplay_shell_flow.configure(
		connection_service,
		scene_root,
		player,
		bullets,
		asteroids,
		gameplay_hud_flow,
		gameplay_menu_flow
	)
	if gameplay_shell_flow.has_method("configure_spectate_menu_state"):
		gameplay_shell_flow.configure_spectate_menu_state(spectate_menu_state)
	dev_tools_session_flow = DevToolsSessionFlow.new()
	dev_tools_session_flow.configure(connection_service, scene_root, gameplay_shell_flow, logger)
	if gameplay_shell_flow != null && gameplay_shell_flow.has_method("configure_debug_placement_route"):
		gameplay_shell_flow.configure_debug_placement_route(
			Callable(dev_tools_session_flow, "begin_debug_click_placement")
		)
	_configure_gameplay_presentation_flow()
	_connect_gameplay_shell_signal("gameplay_started", Callable(self, "_on_gameplay_started"))
	_connect_gameplay_shell_signal(
		"quit_to_main_menu_requested",
		Callable(self, "_on_gameplay_quit_to_main_menu_requested")
	)
	_connect_gameplay_shell_signal(
		"return_to_lobby_requested",
		Callable(self, "_on_gameplay_return_to_lobby_requested")
	)


func handle_gameplay_state(packet: Dictionary) -> void:
	if !accepts_gameplay_packets:
		return
	has_received_gameplay_state = true
	var state := GameplayStatePacketReader.read(packet)
	if gameplay_shell_flow != null && gameplay_shell_flow.has_method("refresh_debug_spawn_player_slots"):
		gameplay_shell_flow.refresh_debug_spawn_player_slots(current_room_max_players())
	if spectate_menu_state != null:
		spectate_menu_state.apply_gameplay_state(state)
	if gameplay_shell_flow != null:
		gameplay_shell_flow.apply_gameplay_state_data(state)


func handle_player_pause_state(packet: Dictionary) -> void:
	if !accepts_gameplay_packets:
		return
	if gameplay_shell_flow != null:
		gameplay_shell_flow.apply_player_pause_state_packet(packet)


func begin_accepting_gameplay_packets() -> void:
	accepts_gameplay_packets = true


func _process(delta: float) -> void:
	if gameplay_shell_flow != null:
		gameplay_shell_flow.process(delta)
	if dev_tools_session_flow != null:
		dev_tools_session_flow.process(delta)
	if gameplay_presentation_flow != null:
		gameplay_presentation_flow.process(delta, has_received_gameplay_state)


func _input(event: InputEvent) -> void:
	if dev_tools_session_flow != null and dev_tools_session_flow.handle_input(event):
		get_viewport().set_input_as_handled()
		return

	if _hud_should_receive_mouse_event(event):
		return

	if gameplay_shell_flow == null:
		return
	if gameplay_shell_flow.handle_unhandled_input(event):
		get_viewport().set_input_as_handled()


func _hud_should_receive_mouse_event(event: InputEvent) -> bool:
	if !(event is InputEventMouseButton):
		return false
	if !event.pressed:
		return false
	if hud == null or !hud.visible:
		return false

	var hovered_control = get_viewport().gui_get_hovered_control()
	if hovered_control == null:
		return false
	if hovered_control == hud:
		return true
	if hud.is_ancestor_of(hovered_control):
		return true
	return false

func _unhandled_input(event: InputEvent) -> void:
	if gameplay_shell_flow == null:
		return
	if gameplay_shell_flow.handle_unhandled_input(event):
		get_viewport().set_input_as_handled()


func _configure_gameplay_presentation_flow() -> void:
	gameplay_presentation_flow = GameplayPresentationFlow.new()
	gameplay_presentation_flow.configure(
		hud,
		player,
		Callable(gameplay_shell_flow, "current_camera"),
		Callable(gameplay_shell_flow, "remote_player_visual_positions"),
		Callable(gameplay_shell_flow, "remote_player_hues")
	)


func reset() -> void:
	accepts_gameplay_packets = false
	has_received_gameplay_state = false
	if dev_tools_session_flow != null:
		dev_tools_session_flow.reset()
	if gameplay_shell_flow != null:
		gameplay_shell_flow.reset()
	if gameplay_presentation_flow != null:
		gameplay_presentation_flow.reset()
	if spectate_menu_state != null:
		spectate_menu_state.reset()


func configure_room_state_provider(provider: Callable) -> void:
	if gameplay_menu_flow != null:
		gameplay_menu_flow.configure_room_state_provider(provider)


func configure_room_max_players_provider(provider: Callable) -> void:
	room_max_players_provider = provider


func current_room_max_players() -> int:
	if room_max_players_provider.is_null():
		return 0
	return int(room_max_players_provider.call())


func refresh_game_over_menu_state() -> void:
	if gameplay_menu_flow != null && gameplay_menu_flow.has_method("refresh_game_over_menu_state"):
		gameplay_menu_flow.refresh_game_over_menu_state()


func _connect_gameplay_shell_signal(signal_name: StringName, handler: Callable) -> void:
	if gameplay_shell_flow.has_signal(signal_name) && !gameplay_shell_flow.is_connected(signal_name, handler):
		gameplay_shell_flow.connect(signal_name, handler)


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


func _log(message: String) -> void:
	if !logger.is_null():
		logger.call(message)
