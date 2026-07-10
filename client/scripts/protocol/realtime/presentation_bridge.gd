extends RefCounted

const DevtoolsLaneStateAdapter = preload("res://scripts/protocol/realtime/devtools_lane_state_adapter.gd")
const ClientLogger = preload("res://scripts/logging/logger.gd")

var realtime_packet_pipeline
var presentation_adapter
var gameplay_composition
var logger: Callable
var _active := false
var _presentation_pending := false
var _lane_presentation_fanned_out := false
var _logged_gameplay_ready := false
var _logged_first_fanout := false
var _logged_event_lifecycle_flow_ready := false
var _logged_debug_shape_catalog_received := false
var _pending_event_lifecycle_flow = null


func configure(pipeline_ref, presentation_adapter_ref, gameplay_composition_ref, logger_callable) -> void:
	realtime_packet_pipeline = pipeline_ref
	presentation_adapter = presentation_adapter_ref
	gameplay_composition = gameplay_composition_ref
	logger = logger_callable


func activate() -> void:
	_active = true


func deactivate() -> void:
	_active = false
	_presentation_pending = false


func mark_pending() -> void:
	if !_active:
		return
	_presentation_pending = true


func has_pending_presentation() -> bool:
	return _presentation_pending


func reset() -> void:
	_active = false
	_presentation_pending = false
	_lane_presentation_fanned_out = false
	_logged_gameplay_ready = false
	_logged_first_fanout = false
	_logged_event_lifecycle_flow_ready = false
	_logged_debug_shape_catalog_received = false
	_pending_event_lifecycle_flow = null


func handle_gameplay_packet(packet: Dictionary) -> void:
	if !_active:
		return
	if realtime_packet_pipeline == null or presentation_adapter == null:
		return
	if !_logged_gameplay_ready and realtime_packet_pipeline.is_gameplay_ready():
		_log("Gameplay lane baselines ready")
		_logged_gameplay_ready = true
	if packet.get("type") == "event_batch" and gameplay_composition != null and gameplay_composition.has_method("get_event_lifecycle_flow"):
		_pending_event_lifecycle_flow = gameplay_composition.get_event_lifecycle_flow()
		if !_logged_event_lifecycle_flow_ready and _pending_event_lifecycle_flow != null:
			_log("Gameplay event fanout target ready: event_lifecycle_flow_null=%s" % str(_pending_event_lifecycle_flow == null))
			_logged_event_lifecycle_flow_ready = true
		var event_lifecycle_flow = _pending_event_lifecycle_flow
		var events = packet.get("events", [])
		var event_types = []
		for event in events:
			event_types.append(str(event.get("type", "")))
		ClientLogger.event(ClientLogger.CATEGORY_GAME, ClientLogger.LEVEL_DEBUG, "gameplay_event_batch_diagnostics", "Gameplay event batch diagnostics", {
			"batch_id": str(packet.get("batch_id", "")),
			"events_size": events.size(),
			"event_types": event_types,
			"event_lifecycle_flow_null": event_lifecycle_flow == null,
		})
	mark_pending()


func flush_pending() -> bool:
	if !_presentation_pending:
		return false
	if realtime_packet_pipeline == null or !realtime_packet_pipeline.is_gameplay_ready() or presentation_adapter == null or gameplay_composition == null:
		return false
	if !_logged_first_fanout:
		_log("Gameplay presentation fanout started")
		_logged_first_fanout = true
	var event_lifecycle_flow = _pending_event_lifecycle_flow
	_pending_event_lifecycle_flow = null
	var world_sync = null
	if gameplay_composition.gameplay_shell_flow != null and gameplay_composition.gameplay_shell_flow.runtime_context != null:
		world_sync = gameplay_composition.gameplay_shell_flow.runtime_context.world_sync
	var gameplay_hud_flow = gameplay_composition.gameplay_hud_flow
	var presentation_state = realtime_packet_pipeline.get_presentation_state()
	presentation_adapter.fanout_lane_states(presentation_state, world_sync, gameplay_hud_flow, event_lifecycle_flow)
	var devtools_state: Dictionary = DevtoolsLaneStateAdapter.new().build_state(presentation_state)
	gameplay_composition.apply_devtools_gameplay_state(devtools_state)
	if gameplay_composition.has_method("restore_alive_presentation_from_realtime_state"):
		gameplay_composition.restore_alive_presentation_from_realtime_state(presentation_state)
	_presentation_pending = false
	_lane_presentation_fanned_out = true
	return true


func _log(message: String) -> void:
	if logger is Callable and logger.is_valid():
		logger.call(message)
