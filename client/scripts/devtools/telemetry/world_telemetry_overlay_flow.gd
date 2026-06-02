extends RefCounted

const WorldTelemetryOverlayScene = preload("res://scenes/devtools/world_telemetry_overlay.tscn")
const WorldTelemetryMetrics = preload("res://scripts/devtools/telemetry/world_telemetry_metrics.gd")
const REFRESH_INTERVAL_MSEC := 250

var overlay = null
var metrics = null
var connection_service = null
var last_refresh_msec: int = -1
var network_metrics: Dictionary = {}


func configure(connection_service_ref = null) -> void:
	connection_service = connection_service_ref
	metrics = WorldTelemetryMetrics.new()


func reset() -> void:
	if metrics != null:
		metrics.reset()
	if is_instance_valid(overlay):
		overlay.queue_free()
	overlay = null
	last_refresh_msec = -1


func ensure_overlay() -> void:
	if is_instance_valid(overlay):
		return

	overlay = WorldTelemetryOverlayScene.instantiate()
	overlay.visible = false
	var main_loop := Engine.get_main_loop()
	if main_loop != null and main_loop.has_method("get_root"):
		main_loop.root.add_child(overlay)


func show_overlay() -> void:
	ensure_overlay()
	if is_instance_valid(overlay):
		overlay.visible = true
		_refresh_overlay()


func hide_overlay() -> void:
	if is_instance_valid(overlay):
		overlay.visible = false


func toggle_overlay() -> void:
	ensure_overlay()
	if is_instance_valid(overlay):
		if overlay.visible:
			hide_overlay()
		else:
			show_overlay()


func is_visible() -> bool:
	return is_instance_valid(overlay) and overlay.visible


func apply_gameplay_state(state: Dictionary) -> void:
	if metrics == null:
		metrics = WorldTelemetryMetrics.new()
	metrics.apply_gameplay_state(state)


func set_network_metrics(metrics_data: Dictionary) -> void:
	network_metrics = metrics_data
	if metrics != null:
		metrics.set_network_metrics(metrics_data)


func world_packet_metrics_snapshot() -> Dictionary:
	if metrics == null:
		return {}
	return metrics.snapshot()


func process(_has_received_state: bool, _delta: float = 0.0) -> void:
	if not is_instance_valid(overlay) or not overlay.visible:
		return

	var now_msec := Time.get_ticks_msec()
	if last_refresh_msec >= 0 and now_msec - last_refresh_msec < REFRESH_INTERVAL_MSEC:
		return
	_refresh_overlay()


func _refresh_overlay() -> void:
	if not is_instance_valid(overlay):
		return
	if metrics == null:
		metrics = WorldTelemetryMetrics.new()

	var merged_metrics: Dictionary = metrics.snapshot()
	for key in network_metrics.keys():
		merged_metrics[key] = network_metrics[key]
	var fps: float = Engine.get_frames_per_second()
	merged_metrics["fps"] = fps
	merged_metrics["frame_ms"] = 1000.0 / max(fps, 1.0)
	merged_metrics["rtt_ms"] = -1
	if connection_service != null and connection_service.has_method("latest_rtt_ms"):
		merged_metrics["rtt_ms"] = connection_service.latest_rtt_ms()
	if network_metrics.has("rtt_ms"):
		merged_metrics["rtt_ms"] = network_metrics["rtt_ms"]

	overlay.refresh_metrics(merged_metrics)
	last_refresh_msec = Time.get_ticks_msec()
