extends RefCounted
class_name GameplayComposition

const SpectateMenuState := preload("res://scripts/gameplay/spectate/spectate_menu_state.gd")
const MatchResultsFlowScript := preload("res://scripts/ui/match_results/match_results_flow.gd")
const GameplayShellFlowScript := preload("res://scripts/shell/gameplay_shell_flow.gd")

signal gameplay_started
signal replay_requested
signal quit_to_main_menu_requested
signal return_to_pregame_requested(session_mode: String)
signal return_to_lobby_requested

var connection_service
var scene_root: Node
var player
var view_anchor
var bullets: Node2D
var asteroids: Node2D
var hud: Control
var gameplay_user_interface: Control
var session_context
var logger: Callable
var room_max_players_provider: Callable
var gameplay_shell_flow
var gameplay_presentation_flow
var gameplay_hud_flow
var gameplay_menu_flow
var match_end_flow
var match_results_flow
var dev_tools_session_flow
var spectate_session_flow

func configure(connection_service_ref, scene_root_ref: Node, player_ref, view_anchor_ref, bullets_ref: Node2D, asteroids_ref: Node2D, pickups_ref: Node2D, hud_ref: Control, gameplay_user_interface_ref: Control, session_context_ref, logger_callable: Callable) -> void:
	connection_service = connection_service_ref
	scene_root = scene_root_ref
	player = player_ref
	view_anchor = view_anchor_ref
	bullets = bullets_ref
	asteroids = asteroids_ref
	hud = hud_ref
	gameplay_user_interface = gameplay_user_interface_ref
	session_context = session_context_ref
	logger = logger_callable
	gameplay_hud_flow = GameplayHudFlow.new()
	gameplay_hud_flow.configure(hud)
	gameplay_menu_flow = GameplayMenuFlow.new()
	gameplay_menu_flow.configure(hud, connection_service, player, session_context)
	var overlay_parent := gameplay_user_interface
	if overlay_parent == null:
		overlay_parent = hud
	gameplay_menu_flow.configure_overlay_parent(overlay_parent)
	match_end_flow = MatchEndFlow.new()
	match_end_flow.configure(gameplay_hud_flow, gameplay_menu_flow, session_context)
	match_results_flow = MatchResultsFlowScript.new()
	var match_results_mount_parent := gameplay_user_interface
	if match_results_mount_parent == null:
		match_results_mount_parent = hud
	match_results_flow.configure(match_results_mount_parent)
	match_end_flow.configure_match_results_flow(match_results_flow)
	var spectate_menu_state = SpectateMenuState.new()
	spectate_session_flow = SpectateSessionFlow.new()
	gameplay_shell_flow = GameplayShellFlow.new()
	gameplay_shell_flow.configure(
		connection_service,
		scene_root,
		player,
		view_anchor_ref,
		bullets,
		asteroids,
		pickups_ref,
		gameplay_hud_flow,
		gameplay_menu_flow,
		spectate_menu_state,
		match_end_flow
	)
	spectate_session_flow.configure(gameplay_menu_flow, gameplay_shell_flow, spectate_menu_state)
	dev_tools_session_flow = DevToolsSessionFlow.new()
	dev_tools_session_flow.configure(connection_service, scene_root, gameplay_shell_flow, logger)
	dev_tools_session_flow.attach_to_gameplay_shell(gameplay_shell_flow)

	_connect_gameplay_shell_signal(&"gameplay_started", Callable(self, "_on_gameplay_started"))
	_connect_gameplay_shell_signal(&"quit_to_main_menu_requested", Callable(self, "_on_gameplay_quit_to_main_menu_requested"))
	_connect_gameplay_shell_signal(&"return_to_pregame_requested", Callable(self, "_on_gameplay_return_to_pregame_requested"))
	_connect_gameplay_shell_signal(&"return_to_lobby_requested", Callable(self, "_on_gameplay_return_to_lobby_requested"))
	_connect_match_end_signal(&"replay_requested", Callable(self, "_on_match_end_replay_requested"))
	_connect_match_end_signal(&"return_to_lobby_requested", Callable(self, "_on_match_end_return_to_lobby_requested"))
	_connect_match_end_signal(&"return_to_pregame_requested", Callable(self, "_on_match_end_return_to_pregame_requested"))
	_connect_match_end_signal(&"quit_to_main_menu_requested", Callable(self, "_on_match_end_quit_to_main_menu_requested"))
	_configure_gameplay_presentation_flow()

func configure_gameplay_readiness(gameplay_readiness) -> void:
	if gameplay_shell_flow == null or gameplay_shell_flow.gameplay_state_flow == null:
		return
	gameplay_shell_flow.gameplay_state_flow.gameplay_readiness = gameplay_readiness

func set_required_lane_baselines_synced(value: bool) -> void:
	if gameplay_shell_flow != null:
		gameplay_shell_flow.set_required_lane_baselines_synced(value)

func get_event_lifecycle_flow():
	if gameplay_shell_flow == null:
		return null
	return gameplay_shell_flow.get_event_lifecycle_flow()

func apply_gameplay_state(state: Dictionary) -> void:
	if gameplay_shell_flow != null:
		gameplay_shell_flow.apply_devtools_gameplay_state(state)
		gameplay_shell_flow.refresh_devtools_spawn_player_slots(_current_room_max_players())
	if spectate_session_flow != null:
		spectate_session_flow.apply_gameplay_state(state)
	if gameplay_shell_flow != null:
		gameplay_shell_flow.apply_gameplay_state(state)

func apply_player_pause_state_packet(packet: Dictionary) -> void:
	if gameplay_shell_flow != null:
		gameplay_shell_flow.apply_player_pause_state_packet(packet)


func apply_devtools_debug_status_packet(packet: Dictionary) -> void:
	if gameplay_shell_flow == null:
		return
	gameplay_shell_flow.apply_devtools_debug_status_packet(packet)


func apply_debug_shape_catalog_packet(packet: Dictionary) -> void:
	if gameplay_shell_flow == null:
		return
	gameplay_shell_flow.apply_debug_shape_catalog_packet(packet)

func process(delta: float, required_lane_baselines_synced: bool) -> void:
	if gameplay_shell_flow != null:
		gameplay_shell_flow.process(delta)
	if dev_tools_session_flow != null:
		dev_tools_session_flow.process(delta)
	if gameplay_presentation_flow != null:
		gameplay_presentation_flow.process(delta, required_lane_baselines_synced)

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
	if match_end_flow != null:
		match_end_flow.reset()
	if match_results_flow != null:
		match_results_flow.clear()
	if spectate_session_flow != null:
		spectate_session_flow.reset()

func configure_room_state_provider(provider: Callable) -> void:
	if gameplay_menu_flow != null:
		gameplay_menu_flow.configure_room_state_provider(provider)
	if match_end_flow != null:
		match_end_flow.configure_room_state_provider(provider)


func configure_match_result_provider(provider: Callable) -> void:
	if match_end_flow != null:
		match_end_flow.configure_match_result_provider(provider)

func configure_room_max_players_provider(provider: Callable) -> void:
	room_max_players_provider = provider

func refresh_match_end_state() -> void:
	if match_end_flow != null:
		match_end_flow.refresh_match_end_state()


func refresh_game_over_menu_state() -> void:
	refresh_match_end_state()

func _connect_gameplay_shell_signal(signal_name: StringName, handler: Callable) -> void:
	if gameplay_shell_flow == null:
		return
	if gameplay_shell_flow.has_signal(signal_name) and not gameplay_shell_flow.is_connected(signal_name, handler):
		gameplay_shell_flow.connect(signal_name, handler)


func _connect_match_end_signal(signal_name: StringName, handler: Callable) -> void:
	if match_end_flow == null:
		return
	if match_end_flow.has_signal(signal_name) and not match_end_flow.is_connected(signal_name, handler):
		match_end_flow.connect(signal_name, handler)

func _current_room_max_players() -> int:
	if room_max_players_provider.is_valid():
		return room_max_players_provider.call()
	return 0

func _configure_gameplay_presentation_flow() -> void:
	gameplay_presentation_flow = GameplayPresentationFlow.new()
	gameplay_presentation_flow.configure(
		hud,
		player,
		Callable(self, "_active_camera"),
		Callable(gameplay_shell_flow.runtime_context.world_sync, "get_remote_player_visual_positions"),
		Callable(gameplay_shell_flow.runtime_context.world_sync, "get_remote_player_hues")
	)


func _active_camera():
	if view_anchor == null:
		return null
	return view_anchor.get_node_or_null("Camera2D")

func _on_gameplay_started() -> void:
	gameplay_started.emit()

func _on_gameplay_quit_to_main_menu_requested() -> void:
	quit_to_main_menu_requested.emit()

func _on_gameplay_return_to_pregame_requested(session_mode: String) -> void:
	return_to_pregame_requested.emit(session_mode)

func _on_gameplay_return_to_lobby_requested() -> void:
	return_to_lobby_requested.emit()


func _on_match_end_replay_requested() -> void:
	replay_requested.emit()


func _on_match_end_return_to_lobby_requested() -> void:
	return_to_lobby_requested.emit()


func _on_match_end_return_to_pregame_requested() -> void:
	return_to_pregame_requested.emit(session_context.active_mode if session_context != null else "")


func _on_match_end_quit_to_main_menu_requested() -> void:
	quit_to_main_menu_requested.emit()
