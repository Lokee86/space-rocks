extends Node
class_name GameplaySessionController

const DevtoolsDisplayRefreshFlow := preload("res://scripts/devtools/devtools_display_refresh_flow.gd")



var connection_service: ClientConnectionService
var hud: Control
var gameplay_user_interface: Control
var main_menu: Control
var session_context
var shell_boot_flow
var logger: Callable

var gameplay_composition: GameplayComposition
var gameplay_state_flow
var gameplay_presentation_adapter: PresentationAdapter
var presentation_bridge: PresentationBridge
var realtime_packet_pipeline: RealtimePacketPipeline

var accepts_gameplay_packets := false

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
		var measurement_context = gameplay_composition.get_client_measurement_context()
		if measurement_context != null and realtime_packet_pipeline != null and realtime_packet_pipeline.has_method("set_measurement_observer"):
			realtime_packet_pipeline.set_measurement_observer(Callable(measurement_context, "record_lane_application"))
		if measurement_context != null and presentation_bridge.has_method("set_measurement_observer"):
			presentation_bridge.set_measurement_observer(Callable(measurement_context, "record_presentation"))
		var gameplay_packet_callable := Callable(presentation_bridge, "handle_gameplay_packet")
		if realtime_packet_pipeline != null and not realtime_packet_pipeline.gameplay_packet_applied.is_connected(gameplay_packet_callable):
			realtime_packet_pipeline.gameplay_packet_applied.connect(gameplay_packet_callable)

func is_gameplay_active() -> bool:
	return accepts_gameplay_packets


func is_camera_zoom_active() -> bool:
	return accepts_gameplay_packets \
		&& gameplay_composition != null \
		&& !gameplay_composition.is_game_over()


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
	if presentation_bridge != null:
		presentation_bridge.activate()

func _process(delta: float) -> void:
	var required_lane_baselines_synced := false
	if realtime_packet_pipeline != null:
		required_lane_baselines_synced = realtime_packet_pipeline.is_gameplay_ready()
	if gameplay_composition != null:
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


func start_measurement(scenario_label: String = "", metadata: Dictionary = {}) -> String:
	if gameplay_composition == null:
		return ""
	return gameplay_composition.start_measurement(scenario_label, metadata)


func stop_measurement() -> String:
	if gameplay_composition == null:
		return ""
	return gameplay_composition.stop_measurement()


func get_measurement_state() -> Dictionary:
	if gameplay_composition == null:
		return {}
	return gameplay_composition.get_measurement_state()


func get_latest_measurement_export_result() -> Dictionary:
	if gameplay_composition == null:
		return {}
	return gameplay_composition.get_latest_measurement_export_result()


func configure_measurement_report_directory(path: String) -> void:
	if gameplay_composition != null:
		gameplay_composition.configure_measurement_report_directory(path)


func request_spectate_target(player_id: String) -> void:
	if gameplay_composition != null:
		gameplay_composition.request_spectate_target(player_id)

func _on_gameplay_started() -> void:
	if main_menu != null:
		main_menu.hide()

func _on_gameplay_quit_to_main_menu_requested() -> void:
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
	if connection_service != null:
		connection_service.send_return_to_lobby_request()
	reset()

func _on_gameplay_return_to_pregame_requested(session_mode: String) -> void:
	if connection_service != null:
		connection_service.begin_graceful_close()
	reset()
	if session_context != null:
		session_context.clear()
	if shell_boot_flow != null:
		shell_boot_flow.clear()
	return_to_pregame_requested.emit(session_mode)

func _on_gameplay_replay_requested() -> void:
	if connection_service != null:
		await connection_service.close_gracefully()
	reset()
	if session_context != null:
		session_context.clear()
	if shell_boot_flow != null:
		shell_boot_flow.clear()
	replay_requested.emit()
