extends RefCounted
class_name ClientMeasurementContext

const ClientMeasurementSessionScript := preload("res://scripts/devtools/measurement/client_measurement_session.gd")
const LongMatchStoreMetrics := preload("res://scripts/devtools/telemetry/long_match_store_metrics.gd")
const CUMULATIVE_NETWORK_KEYS := [
	"packets_in",
	"packets_out",
	"bytes_in",
	"bytes_out",
	"decode_failures",
	"encode_failures",
	"send_failures",
	"bullet_delta_received_count",
	"bullet_delta_missing_server_time_count",
	"bullet_delta_unsynchronized_server_time_count",
	"bullet_delta_clock_skew_count",
]

var session: ClientMeasurementSessionScript
var connection_service
var realtime_packet_pipeline
var world_sync
var telemetry_context
var _network_metrics_baseline: Dictionary = {}

signal finalized(result: Dictionary)


func _init() -> void:
	session = ClientMeasurementSessionScript.new()
	session.set_periodic_sample_provider(Callable(self, "_build_periodic_sample"))
	session.finalized.connect(Callable(self, "_on_session_finalized"))


func configure(connection_service_ref, realtime_packet_pipeline_ref = null, world_sync_ref = null, telemetry_context_ref = null) -> void:
	connection_service = connection_service_ref
	realtime_packet_pipeline = realtime_packet_pipeline_ref
	world_sync = world_sync_ref
	telemetry_context = telemetry_context_ref


func start(metadata: Dictionary = {}) -> void:
	_network_metrics_baseline = _connection_network_metrics()
	session.start(metadata)
	session.record_periodic_sample(_build_periodic_sample())


func stop(reason: String = "") -> Dictionary:
	if session.is_recording():
		session.record_periodic_sample(_build_periodic_sample())
	var result := session.stop(reason)
	_network_metrics_baseline.clear()
	return result


func reset() -> Dictionary:
	var result := session.reset()
	_network_metrics_baseline.clear()
	return result


func is_recording() -> bool:
	return session.is_recording()


func snapshot() -> Dictionary:
	return session.snapshot()


func process_frame(delta: float) -> void:
	session.process_frame(delta)


func record_lane_application(lane: String, duration_ms: float) -> void:
	session.record_lane_application(lane, duration_ms)


func record_presentation(duration_ms: float) -> void:
	session.record_presentation(duration_ms)


func record_lifecycle(entity_kind: String, operation: String, count: int = 1) -> void:
	session.record_lifecycle(entity_kind, operation, count)


func _on_session_finalized(result: Dictionary) -> void:
	finalized.emit(result.duplicate(true))


func _connection_network_metrics() -> Dictionary:
	if connection_service != null and connection_service.has_method("network_metrics_snapshot"):
		var result = connection_service.network_metrics_snapshot()
		if result is Dictionary:
			return result.duplicate(true)
	return {}


func _run_relative_network_metrics(current: Dictionary, baseline: Dictionary) -> Dictionary:
	var relative := current.duplicate(true)
	for key in CUMULATIVE_NETWORK_KEYS:
		var current_value = current.get(key, null)
		if !(current_value is int or current_value is float):
			continue
		var baseline_value = baseline.get(key, 0)
		if !(baseline_value is int or baseline_value is float):
			baseline_value = 0
		relative[key] = maxi(int(current_value) - int(baseline_value), 0)
	for key in current.keys():
		var current_child = current[key]
		if !(current_child is Dictionary):
			continue
		var baseline_child = baseline.get(key, {})
		relative[key] = _run_relative_network_metrics(
			current_child,
			baseline_child if baseline_child is Dictionary else {}
		)
	return relative


func _build_periodic_sample() -> Dictionary:
	var missing_metrics: Array = []
	var entity_counts := {}
	if world_sync != null and world_sync.has_method("entity_counts"):
		entity_counts = world_sync.entity_counts()
	else:
		missing_metrics.append("entity_counts")

	var network_metrics := {}
	if telemetry_context != null and telemetry_context.has_method("telemetry_snapshot"):
		network_metrics = telemetry_context.telemetry_snapshot()
	else:
		missing_metrics.append("telemetry")
	if connection_service != null and connection_service.has_method("network_metrics_snapshot"):
		var connection_network_metrics := _connection_network_metrics()
		for key in connection_network_metrics.keys():
			network_metrics[key] = connection_network_metrics[key]
		network_metrics = _run_relative_network_metrics(network_metrics, _network_metrics_baseline)
	else:
		missing_metrics.append("network_metrics")

	var long_match_store_metrics := LongMatchStoreMetrics.snapshot(realtime_packet_pipeline, world_sync)
	var node_count := -1
	if world_sync != null and world_sync.has_method("scene_node_count"):
		node_count = world_sync.scene_node_count()
	else:
		missing_metrics.append("node_count")

	var memory_bytes := -1
	if OS.has_method("get_static_memory_usage"):
		memory_bytes = int(OS.get_static_memory_usage())
	else:
		missing_metrics.append("memory")

	return {
		"timestamp": Time.get_ticks_msec(),
		"fps": Engine.get_frames_per_second(),
		"memory_bytes": memory_bytes,
		"node_count": node_count,
		"entity_counts": entity_counts,
		"network_metrics": network_metrics,
		"long_match_store_metrics": long_match_store_metrics,
		"missing_metrics": missing_metrics,
	}
