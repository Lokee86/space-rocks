extends Node

const DevtoolsDisplayRefreshFlow := preload("res://scripts/devtools/devtools_display_refresh_flow.gd")
const PresentationBridge := preload("res://scripts/protocol/realtime/presentation_bridge.gd")

var connection_service
var hud: Control
var gameplay_user_interface: Control
var main_menu: Control
var session_context
var shell_boot_flow
var logger: Callable

var gameplay_composition
var gameplay_state_flow
var gameplay_presentation_adapter
var presentation_bridge
var realtime_packet_pipeline

var accepts_gameplay_packets := false
var _logged_debug_shape_catalog_received := false

signal return_to_pregame_requested(session_mode: String)
signal replay_requested

func configure(
	connection_service_ref,
	realtime_packet_pipeline_ref,
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
	realtime_packet_pipeline = realtime_packet_pipeline_ref
	hud = hud_ref
	gameplay_user_interface = gameplay_user_interface_ref
	main_menu = main_menu_ref
	session_context = session_context_ref
	shell_boot_flow = shell_boot_flow_ref
	logger = logger_callable

	gameplay_presentation_adapter = preload("res://scripts/protocol/realtime/presentation_adapter.gd").new()
	presentation_bridge = PresentationBridge.new()
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
	if presentation_bridge != null:
		presentation_bridge.configure(realtime_packet_pipeline, gameplay_presentation_adapter, gameplay_composition, logger)
		var gameplay_packet_callable := Callable(presentation_bridge, "handle_gameplay_packet")
		if realtime_packet_pipeline != null and not realtime_packet_pipeline.gameplay_packet_applied.is_connected(gameplay_packet_callable):
			realtime_packet_pipeline.gameplay_packet_applied.connect(gameplay_packet_callable)

func handle_player_pause_state(packet: Dictionary) -> void:
	if !accepts_gameplay_packets:
		return
	if gameplay_composition != null:
		gameplay_composition.apply_player_pause_state_packet(packet)

func handle_debug_status_packet(packet: Dictionary) -> void:
	if gameplay_composition != null:
		gameplay_composition.apply_devtools_debug_status_packet(packet)

func handle_debug_shape_catalog_packet(packet: Dictionary) -> void:
	if !_logged_debug_shape_catalog_received:
		var shape_count := 0
		var shapes = packet.get("shapes", {})
		if shapes is Dictionary:
			shape_count = shapes.size()
		_log("debug shape catalog received: shape_count=%d" % shape_count)
		_logged_debug_shape_catalog_received = true
	if gameplay_composition != null:
		gameplay_composition.apply_debug_shape_catalog_packet(packet)

func begin_accepting_gameplay_packets() -> void:
	_log("accepting gameplay packets: realtime_packet_pipeline_null=%s presentation_state_null=%s" % [str(realtime_packet_pipeline == null), str(realtime_packet_pipeline == null or realtime_packet_pipeline.get_presentation_state() == null)])
	accepts_gameplay_packets = true
	if presentation_bridge != null:
		presentation_bridge.activate()

func _process(delta: float) -> void:
	var required_lane_baselines_synced := false
	if realtime_packet_pipeline != null:
		required_lane_baselines_synced = realtime_packet_pipeline.is_gameplay_ready()
	if gameplay_composition != null and gameplay_composition.has_method("set_required_lane_baselines_synced"):
		gameplay_composition.set_required_lane_baselines_synced(required_lane_baselines_synced)
	if presentation_bridge != null:
		presentation_bridge.flush_pending()
	if gameplay_composition != null:
		gameplay_composition.process(delta, required_lane_baselines_synced)

func _input(event: InputEvent) -> void:
	if !accepts_gameplay_packets:
		return
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
	if !accepts_gameplay_packets:
		return
	if gameplay_composition == null:
		return
	if gameplay_composition.handle_gameplay_input(event):
		get_viewport().set_input_as_handled()

func reset() -> void:
	accepts_gameplay_packets = false
	if presentation_bridge != null:
		presentation_bridge.reset()
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
