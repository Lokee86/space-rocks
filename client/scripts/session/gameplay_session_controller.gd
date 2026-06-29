extends Node

const DevtoolsDisplayRefreshFlow := preload("res://scripts/devtools/devtools_display_refresh_flow.gd")
const DevtoolsLaneStateAdapter := preload("res://scripts/protocol/realtime/devtools_lane_state_adapter.gd")

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
var devtools_lane_state_adapter
var gameplay_realtime_router

var accepts_gameplay_packets := false
var _lane_presentation_fanned_out := false
var _gameplay_readiness
var _logged_gameplay_ready := false
var _logged_first_fanout := false
var _logged_event_lifecycle_flow_ready := false
var _logged_stale_dead_hud_clear := false
var _logged_debug_shape_catalog_received := false

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

	gameplay_presentation_adapter = preload("res://scripts/protocol/realtime/presentation_adapter.gd").new()
	devtools_lane_state_adapter = DevtoolsLaneStateAdapter.new()
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
	_bind_realtime_protocol_dependencies()


func _bind_realtime_protocol_dependencies() -> void:
	_gameplay_readiness = null
	if connection_service != null and connection_service.has_method("get_gameplay_readiness"):
		_gameplay_readiness = connection_service.get_gameplay_readiness()
	if gameplay_presentation_adapter != null and _gameplay_readiness != null:
		gameplay_presentation_adapter.bind_gameplay_readiness(_gameplay_readiness)
	if gameplay_state_flow != null:
		gameplay_state_flow.set_gameplay_readiness(_gameplay_readiness)
	gameplay_realtime_router = null
	if connection_service != null and connection_service.has_method("get_realtime_router"):
		gameplay_realtime_router = connection_service.get_realtime_router()


func handle_gameplay_packet(packet: Dictionary) -> void:
	if !accepts_gameplay_packets:
		return
	if _gameplay_readiness == null or !bool(_gameplay_readiness.is_gameplay_ready()) or gameplay_presentation_adapter == null:
		return

	if !_logged_gameplay_ready:
		_log("Gameplay lane baselines ready")
		_logged_gameplay_ready = true

	var event_lifecycle_flow = null
	if packet.get("type") == "event_batch" and gameplay_composition != null and gameplay_composition.has_method("get_event_lifecycle_flow"):
		event_lifecycle_flow = gameplay_composition.get_event_lifecycle_flow()
		if !_logged_event_lifecycle_flow_ready and event_lifecycle_flow != null:
			_log("Gameplay event fanout target ready: event_lifecycle_flow_null=%s" % str(event_lifecycle_flow == null))
			_logged_event_lifecycle_flow_ready = true
		var events = packet.get("events", [])
		var event_types = []
		for event in events:
			event_types.append(str(event.get("type", "")))
		_log(
			"Gameplay event batch diagnostics: batch_id=%s events_size=%d event_types=%s event_lifecycle_flow_null=%s" % [
				str(packet.get("batch_id", "")),
				events.size(),
				str(event_types),
				str(event_lifecycle_flow == null)
			]
		)

	if gameplay_realtime_router != null and gameplay_presentation_adapter.can_fanout():
		if !_logged_first_fanout:
			_log("Gameplay presentation fanout started")
			_logged_first_fanout = true
		var world_sync = null
		if gameplay_composition != null and gameplay_composition.gameplay_shell_flow != null and gameplay_composition.gameplay_shell_flow.runtime_context != null:
			world_sync = gameplay_composition.gameplay_shell_flow.runtime_context.world_sync
		var gameplay_hud_flow = null
		if gameplay_composition != null:
			gameplay_hud_flow = gameplay_composition.gameplay_hud_flow
		gameplay_presentation_adapter.fanout_lane_states(gameplay_realtime_router, world_sync, gameplay_hud_flow, event_lifecycle_flow)
		if gameplay_composition != null and devtools_lane_state_adapter != null:
			var devtools_state: Dictionary = devtools_lane_state_adapter.build_state(gameplay_realtime_router)
			gameplay_composition.apply_devtools_gameplay_state(devtools_state)
		_confirm_respawn_restored_alive_hud(gameplay_hud_flow)
		_clear_stale_dead_presentation_from_lane_state(gameplay_hud_flow)
		if !_lane_presentation_fanned_out:
			gameplay_presentation_adapter.mark_fanned_out()
			_lane_presentation_fanned_out = true


func _confirm_respawn_restored_alive_hud(gameplay_hud_flow) -> void:
	if gameplay_hud_flow == null or gameplay_realtime_router == null:
		return
	if gameplay_composition == null or gameplay_composition.gameplay_shell_flow == null:
		return
	var runtime_context = gameplay_composition.gameplay_shell_flow.runtime_context
	if runtime_context == null:
		return
	var respawn_flow = runtime_context.respawn_flow
	if respawn_flow == null or !respawn_flow.has_method("is_awaiting_confirmation"):
		return
	if !respawn_flow.is_awaiting_confirmation():
		return
	var self_id := ""
	if gameplay_realtime_router.overlay_lane_state != null and gameplay_realtime_router.overlay_lane_state.self_id != null:
		self_id = str(gameplay_realtime_router.overlay_lane_state.self_id)
	if self_id == "":
		return
	var lifecycle = null
	if gameplay_realtime_router.session_lane_state != null and gameplay_realtime_router.session_lane_state.player_lifecycle != null:
		lifecycle = gameplay_realtime_router.session_lane_state.player_lifecycle.get(self_id)
	var lifecycle_status := ""
	if lifecycle is Dictionary:
		lifecycle_status = str(lifecycle.get("status", ""))
	else:
		lifecycle_status = str(lifecycle)
	if lifecycle_status != "active":
		return
	if gameplay_realtime_router.world_lane_state == null or gameplay_realtime_router.world_lane_state.ships == null:
		return
	if !gameplay_realtime_router.world_lane_state.ships.has(self_id):
		return
	gameplay_hud_flow.clear_dead_presentation()
	respawn_flow.clear_awaiting_confirmation()
	_log("respawn confirmation restored alive HUD")


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
	_bind_realtime_protocol_dependencies()
	_log(
		"accepting gameplay packets: gameplay_readiness_null=%s realtime_router_null=%s" % [
			str(_gameplay_readiness == null),
			str(gameplay_realtime_router == null)
		]
	)
	accepts_gameplay_packets = true


func _process(delta: float) -> void:
	if gameplay_composition != null:
		var required_lane_baselines_synced := false
		if gameplay_state_flow != null:
			required_lane_baselines_synced = gameplay_state_flow.is_gameplay_ready()
		if gameplay_composition.has_method("set_required_lane_baselines_synced"):
			gameplay_composition.set_required_lane_baselines_synced(required_lane_baselines_synced)
		gameplay_composition.process(delta, required_lane_baselines_synced)


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
	_lane_presentation_fanned_out = false
	_logged_gameplay_ready = false
	_logged_first_fanout = false
	_logged_event_lifecycle_flow_ready = false
	_logged_debug_shape_catalog_received = false
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

func _clear_stale_dead_presentation_from_lane_state(gameplay_hud_flow) -> void:
	if gameplay_hud_flow == null or gameplay_realtime_router == null:
		return
	if gameplay_hud_flow.hidden_for_match_over or gameplay_hud_flow.is_game_over:
		return
	if !gameplay_hud_flow._has_dead_presentation():
		return
	if gameplay_composition == null or gameplay_composition.gameplay_shell_flow == null:
		return
	var runtime_context = gameplay_composition.gameplay_shell_flow.runtime_context
	if runtime_context == null:
		return
	var self_id := ""
	if gameplay_realtime_router.overlay_lane_state != null and gameplay_realtime_router.overlay_lane_state.self_id != null:
		self_id = str(gameplay_realtime_router.overlay_lane_state.self_id)
	if self_id == "":
		return
	var lifecycle = null
	if gameplay_realtime_router.session_lane_state != null and gameplay_realtime_router.session_lane_state.player_lifecycle != null:
		lifecycle = gameplay_realtime_router.session_lane_state.player_lifecycle.get(self_id)
	var lifecycle_status := ""
	if lifecycle is Dictionary:
		lifecycle_status = str(lifecycle.get("status", ""))
	else:
		lifecycle_status = str(lifecycle)
	if lifecycle_status != "active":
		return
	if gameplay_realtime_router.world_lane_state == null or gameplay_realtime_router.world_lane_state.ships == null:
		return
	if !gameplay_realtime_router.world_lane_state.ships.has(self_id):
		return
	var respawn_flow = runtime_context.respawn_flow
	if respawn_flow != null and respawn_flow.has_method("clear_awaiting_confirmation"):
		respawn_flow.clear_awaiting_confirmation()
	if !_logged_stale_dead_hud_clear:
		_logged_stale_dead_hud_clear = true
		_log("stale dead HUD cleared from confirmed alive lane state")
	gameplay_hud_flow.clear_dead_presentation()
