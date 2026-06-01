extends RefCounted

const WorldTelemetryOverlayFlow = preload("res://scripts/devtools/telemetry/world_telemetry_overlay_flow.gd")

var overlay_flow = null


func configure(connection_service_ref) -> void:
	overlay_flow = WorldTelemetryOverlayFlow.new()
	overlay_flow.configure(connection_service_ref)


func reset() -> void:
	if overlay_flow != null:
		overlay_flow.reset()


func apply_gameplay_state(state: Dictionary) -> void:
	if overlay_flow != null:
		overlay_flow.apply_gameplay_state(state)


func process(has_received_state: bool, delta: float = 0.0) -> void:
	if overlay_flow != null:
		overlay_flow.process(has_received_state, delta)


func toggle_overlay() -> void:
	if overlay_flow != null:
		overlay_flow.toggle_overlay()
