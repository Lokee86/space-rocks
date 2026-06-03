extends RefCounted
class_name GameplayComposition

const SpectateMenuState := preload("res://scripts/gameplay/spectate/spectate_menu_state.gd")

signal gameplay_started
signal quit_to_main_menu_requested
signal return_to_lobby_requested

var connection_service
var scene_root: Node
var player
var bullets: Node2D
var asteroids: Node2D
var hud: Control
var session_context
var logger: Callable
var room_max_players_provider: Callable
var gameplay_shell_flow
var gameplay_presentation_flow
var gameplay_hud_flow
var gameplay_menu_flow
var dev_tools_session_flow
var spectate_session_flow

func configure(connection_service_ref, scene_root_ref: Node, player_ref, bullets_ref: Node2D, asteroids_ref: Node2D, hud_ref: Control, session_context_ref, logger_callable: Callable) -> void:
	connection_service = connection_service_ref
	scene_root = scene_root_ref
	player = player_ref
	bullets = bullets_ref
	asteroids = asteroids_ref
	hud = hud_ref
	session_context = session_context_ref
	logger = logger_callable
	gameplay_hud_flow = GameplayHudFlow.new()
	gameplay_hud_flow.configure(hud)
	gameplay_menu_flow = GameplayMenuFlow.new()
	gameplay_menu_flow.configure(hud, connection_service, player, session_context)
	var spectate_menu_state = SpectateMenuState.new()
	spectate_session_flow = SpectateSessionFlow.new()
	gameplay_shell_flow = GameplayShellFlow.new()
	gameplay_shell_flow.configure(
		connection_service,
		scene_root,
		player,
		bullets,
		asteroids,
		gameplay_hud_flow,
		gameplay_menu_flow,
		spectate_menu_state
	)
	spectate_session_flow.configure(gameplay_menu_flow, gameplay_shell_flow, spectate_menu_state)
	dev_tools_session_flow = DevToolsSessionFlow.new()
	dev_tools_session_flow.configure(connection_service, scene_root, gameplay_shell_flow, logger)
	dev_tools_session_flow.attach_to_gameplay_shell(gameplay_shell_flow)

	_connect_gameplay_shell_signal(&"gameplay_started", Callable(self, "_on_gameplay_started"))
	_connect_gameplay_shell_signal(&"quit_to_main_menu_requested", Callable(self, "_on_gameplay_quit_to_main_menu_requested"))
	_connect_gameplay_shell_signal(&"return_to_lobby_requested", Callable(self, "_on_gameplay_return_to_lobby_requested"))
	_configure_gameplay_presentation_flow()

func apply_gameplay_state(state: Dictionary) -> void:
	if gameplay_shell_flow != null && gameplay_shell_flow.devtools_context != null:
		gameplay_shell_flow.devtools_context.refresh_spawn_player_slots(_current_room_max_players())
	if spectate_session_flow != null:
		spectate_session_flow.apply_gameplay_state(state)
	if gameplay_shell_flow != null:
		gameplay_shell_flow.apply_gameplay_state(state)

func apply_player_pause_state_packet(packet: Dictionary) -> void:
	if gameplay_shell_flow != null:
		gameplay_shell_flow.apply_player_pause_state_packet(packet)

func process(delta: float, has_received_gameplay_state: bool) -> void:
	if gameplay_shell_flow != null:
		gameplay_shell_flow.process(delta)
	if dev_tools_session_flow != null:
		dev_tools_session_flow.process(delta)
	if gameplay_presentation_flow != null:
		gameplay_presentation_flow.process(delta, has_received_gameplay_state)

func handle_devtools_input(event: InputEvent) -> bool:
	if dev_tools_session_flow == null:
		return false
	return dev_tools_session_flow.handle_input(event)

func handle_gameplay_input(event: InputEvent) -> bool:
	if gameplay_shell_flow == null:
		return false
	return gameplay_shell_flow.handle_unhandled_input(event)

func reset() -> void:
	if dev_tools_session_flow != null:
		dev_tools_session_flow.reset()
	if gameplay_shell_flow != null:
		gameplay_shell_flow.reset()
	if gameplay_presentation_flow != null:
		gameplay_presentation_flow.reset()
	if spectate_session_flow != null:
		spectate_session_flow.reset()

func configure_room_state_provider(provider: Callable) -> void:
	if gameplay_menu_flow != null:
		gameplay_menu_flow.configure_room_state_provider(provider)

func configure_room_max_players_provider(provider: Callable) -> void:
	room_max_players_provider = provider

func refresh_game_over_menu_state() -> void:
	if gameplay_menu_flow != null:
		gameplay_menu_flow.refresh_game_over_menu_state()

func _connect_gameplay_shell_signal(signal_name: StringName, handler: Callable) -> void:
	if gameplay_shell_flow == null:
		return
	if gameplay_shell_flow.has_signal(signal_name) and not gameplay_shell_flow.is_connected(signal_name, handler):
		gameplay_shell_flow.connect(signal_name, handler)

func _current_room_max_players() -> int:
	if room_max_players_provider.is_valid():
		return room_max_players_provider.call()
	return 0

func _configure_gameplay_presentation_flow() -> void:
	gameplay_presentation_flow = GameplayPresentationFlow.new()
	gameplay_presentation_flow.configure(
		hud,
		player,
		Callable(gameplay_shell_flow.runtime_context, "current_camera"),
		Callable(gameplay_shell_flow.runtime_context, "remote_player_visual_positions"),
		Callable(gameplay_shell_flow.runtime_context, "remote_player_hues")
	)

func _on_gameplay_started() -> void:
	gameplay_started.emit()

func _on_gameplay_quit_to_main_menu_requested() -> void:
	quit_to_main_menu_requested.emit()

func _on_gameplay_return_to_lobby_requested() -> void:
	return_to_lobby_requested.emit()
