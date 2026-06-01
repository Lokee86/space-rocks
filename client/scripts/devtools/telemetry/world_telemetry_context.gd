extends RefCounted

const WorldTelemetryOverlayFlow = preload("res://scripts/devtools/telemetry/world_telemetry_overlay_flow.gd")
const NetworkTelemetryMetrics = preload("res://scripts/devtools/telemetry/network_telemetry_metrics.gd")
const PING_INTERVAL_MSEC := 1000

var overlay_flow = null
var network_metrics = null
var connection_service = null
var last_ping_msec: int = -1


func configure(connection_service_ref) -> void:
	connection_service = connection_service_ref
	last_ping_msec = -1
	network_metrics = NetworkTelemetryMetrics.new()
	overlay_flow = WorldTelemetryOverlayFlow.new()
	overlay_flow.configure(connection_service_ref)
	if connection_service_ref != null and connection_service_ref.has_signal("telemetry_pong_received"):
		var apply_pong_callable := Callable(self, "_on_telemetry_pong_received")
		if not connection_service_ref.is_connected("telemetry_pong_received", apply_pong_callable):
			connection_service_ref.connect("telemetry_pong_received", apply_pong_callable)


func reset() -> void:
	last_ping_msec = -1
	if network_metrics != null:
		network_metrics.reset()
	if overlay_flow != null:
		overlay_flow.reset()


func apply_gameplay_state(state: Dictionary) -> void:
	if overlay_flow != null:
		overlay_flow.apply_gameplay_state(state)


func process(has_received_state: bool, delta: float = 0.0) -> void:
	if overlay_flow != null:
		if network_metrics != null:
			overlay_flow.set_network_metrics(network_metrics.snapshot())
			_process_ping()
		overlay_flow.process(has_received_state, delta)


func toggle_overlay() -> void:
	if overlay_flow != null:
		overlay_flow.toggle_overlay()


func _on_telemetry_pong_received(packet: Dictionary) -> void:
	if network_metrics != null:
		network_metrics.apply_pong(packet)


func _process_ping() -> void:
	if connection_service == null:
		return
	if overlay_flow == null or not overlay_flow.is_visible():
		return
	if connection_service.has_method("is_server_connected") and not connection_service.is_server_connected():
		return

	var now_msec := Time.get_ticks_msec()
	if last_ping_msec >= 0 and now_msec - last_ping_msec < PING_INTERVAL_MSEC:
		return

	connection_service.send_packet(network_metrics.next_ping_packet())
	last_ping_msec = now_msec
