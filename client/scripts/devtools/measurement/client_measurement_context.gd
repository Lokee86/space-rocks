extends RefCounted
class_name ClientMeasurementContext

const ClientMeasurementSessionScript := preload("res://scripts/devtools/measurement/client_measurement_session.gd")
const LongMatchStoreMetrics := preload("res://scripts/devtools/telemetry/long_match_store_metrics.gd")

var session: ClientMeasurementSessionScript
var connection_service
var realtime_packet_pipeline
var world_sync
var telemetry_context

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
	session.start(metadata)


func stop(reason: String = "") -> Dictionary:
	return session.stop(reason)


func reset() -> Dictionary:
	return session.reset()


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
		var connection_network_metrics: Dictionary = connection_service.network_metrics_snapshot()
		for key in connection_network_metrics.keys():
			network_metrics[key] = connection_network_metrics[key]
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
