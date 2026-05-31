extends Node


const SpectateMenuState := preload("res://scripts/gameplay/spectate/spectate_menu_state.gd")
const GameplayStatePacketReader := preload("res://scripts/gameplay/state/gameplay_state_packet_reader.gd")
const DebugKillInputFlow := preload("res://scripts/gameplay/devtools/debug_kill_input_flow.gd")
const DebugMouseWorldPosition := preload("res://scripts/gameplay/devtools/debug_mouse_world_position.gd")
const DebugClickPlacementFlow := preload("res://scripts/gameplay/devtools/debug_click_placement_flow.gd")

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
var debug_kill_input_flow
var debug_mouse_world_position
var debug_click_placement_flow
var spectate_menu_state
var has_received_gameplay_state := false


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
	debug_kill_input_flow = DebugKillInputFlow.new()
	debug_kill_input_flow.configure(connection_service)
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
	if scene_root is Node2D:
		debug_mouse_world_position = DebugMouseWorldPosition.new()
		debug_mouse_world_position.configure(
			scene_root,
			Callable(gameplay_shell_flow, "server_position_for_visual_position")
		)
		debug_click_placement_flow = DebugClickPlacementFlow.new()
		debug_click_placement_flow.configure(debug_mouse_world_position)
		debug_click_placement_flow.placement_completed.connect(
			Callable(self, "_on_debug_click_placement_completed")
		)
		debug_click_placement_flow.placement_cancelled.connect(
			Callable(self, "_on_debug_click_placement_cancelled")
		)
		if gameplay_shell_flow != null && gameplay_shell_flow.has_method("configure_debug_placement_route"):
			gameplay_shell_flow.configure_debug_placement_route(
				Callable(self, "begin_debug_click_placement")
			)
	if gameplay_shell_flow.has_method("configure_spectate_menu_state"):
		gameplay_shell_flow.configure_spectate_menu_state(spectate_menu_state)
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
	has_received_gameplay_state = true
	var state := GameplayStatePacketReader.read(packet)
	if gameplay_shell_flow != null && gameplay_shell_flow.has_method("refresh_debug_spawn_player_slots"):
		gameplay_shell_flow.refresh_debug_spawn_player_slots(current_room_max_players())
	if spectate_menu_state != null:
		spectate_menu_state.apply_gameplay_state(state)
	if gameplay_shell_flow != null:
		gameplay_shell_flow.apply_gameplay_state_data(state)


func handle_player_pause_state(packet: Dictionary) -> void:
	if gameplay_shell_flow != null:
		gameplay_shell_flow.apply_player_pause_state_packet(packet)


func _process(delta: float) -> void:
	if gameplay_shell_flow != null:
		gameplay_shell_flow.process(delta)
	if debug_kill_input_flow != null:
		debug_kill_input_flow.process()
	if gameplay_presentation_flow != null:
		gameplay_presentation_flow.process(delta, has_received_gameplay_state)


func _input(event: InputEvent) -> void:
	if debug_click_placement_flow == null:
		return
	if !debug_click_placement_flow.is_active():
		return
	if debug_click_placement_flow.handle_unhandled_input(event):
		get_viewport().set_input_as_handled()


func begin_debug_click_placement(action_name: StringName, placement_context: Dictionary = {}) -> void:
	if debug_click_placement_flow == null:
		return
	debug_click_placement_flow.begin(action_name, placement_context)


func _configure_gameplay_presentation_flow() -> void:
	gameplay_presentation_flow = GameplayPresentationFlow.new()
	gameplay_presentation_flow.configure(
		hud,
		player,
		Callable(gameplay_shell_flow, "current_camera"),
		Callable(gameplay_shell_flow, "remote_player_visual_positions")
	)


func reset() -> void:
	has_received_gameplay_state = false
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


func _on_debug_click_placement_completed(result: Dictionary) -> void:
	_log(
		"Debug click placement completed: %s at %s has_direction=%s direction=%s"
		% [
			String(result.get("action_name", StringName())),
			str(result.get("server_position", Vector2.ZERO)),
			str(result.get("has_direction", false)),
			str(result.get("direction", Vector2.ZERO))
		]
	)
	if gameplay_shell_flow != null && gameplay_shell_flow.has_method("handle_debug_placement_result"):
		gameplay_shell_flow.handle_debug_placement_result(result)


func _on_debug_click_placement_cancelled(action_name: StringName) -> void:
	_log("Debug click placement cancelled: %s" % String(action_name))


func _log(message: String) -> void:
	if !logger.is_null():
		logger.call(message)
