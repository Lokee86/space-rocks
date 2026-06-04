extends RefCounted
class_name DevtoolsOverlayContext

const PlayerDevLabelsContext := preload("res://scripts/devtools/player_labels/player_dev_labels_context.gd")
const WorldTelemetryContext := preload("res://scripts/devtools/telemetry/world_telemetry_context.gd")

var player_dev_labels_context
var world_telemetry_context
var server_hitbox_overlay
var remote_player_nodes_provider: Callable


func configure(state_context_ref, connection_service_ref) -> void:
	player_dev_labels_context = PlayerDevLabelsContext.new()
	if !remote_player_nodes_provider.is_null() and remote_player_nodes_provider.is_valid():
		player_dev_labels_context.configure(remote_player_nodes_provider)
	world_telemetry_context = WorldTelemetryContext.new()
	world_telemetry_context.configure(connection_service_ref)


func configure_remote_player_nodes_provider(provider: Callable) -> void:
	remote_player_nodes_provider = provider
	if player_dev_labels_context != null:
		player_dev_labels_context.configure(remote_player_nodes_provider)


func set_player_dev_label_mode(mode: String) -> void:
	if player_dev_labels_context != null && player_dev_labels_context.has_method("set_mode"):
		player_dev_labels_context.set_mode(mode)


func configure_server_hitbox_overlay(overlay_ref) -> void:
	server_hitbox_overlay = overlay_ref
	if server_hitbox_overlay != null && server_hitbox_overlay.has_method("set_hitbox_entries"):
		server_hitbox_overlay.set_hitbox_entries([])


func get_player_dev_labels_context():
	return player_dev_labels_context


func get_world_telemetry_context():
	return world_telemetry_context


func get_server_hitbox_overlay():
	return server_hitbox_overlay


func set_server_hitboxes_enabled(enabled: bool) -> void:
	if server_hitbox_overlay == null || !is_instance_valid(server_hitbox_overlay):
		return
	if server_hitbox_overlay.has_method("set_enabled"):
		server_hitbox_overlay.set_enabled(enabled)


func toggle_world_telemetry_overlay() -> void:
	if world_telemetry_context != null:
		world_telemetry_context.toggle_overlay()


func reset() -> void:
	if server_hitbox_overlay != null && is_instance_valid(server_hitbox_overlay) and server_hitbox_overlay.has_method("set_hitbox_entries"):
		server_hitbox_overlay.set_hitbox_entries([])
	if player_dev_labels_context != null && player_dev_labels_context.has_method("clear_labels"):
		player_dev_labels_context.clear_labels()
	if world_telemetry_context != null:
		world_telemetry_context.reset()


func process(has_received_state: bool) -> void:
	if player_dev_labels_context != null and world_telemetry_context != null:
		if world_telemetry_context.has_method("telemetry_snapshot") and player_dev_labels_context.has_method("apply_network_metrics"):
			player_dev_labels_context.apply_network_metrics(world_telemetry_context.telemetry_snapshot())
	if player_dev_labels_context != null && player_dev_labels_context.has_method("sync_remote_labels"):
		player_dev_labels_context.sync_remote_labels()
	if world_telemetry_context != null:
		world_telemetry_context.process(has_received_state, 0.0)


func apply_gameplay_state(state: Dictionary) -> void:
	if player_dev_labels_context != null && player_dev_labels_context.has_method("apply_gameplay_state"):
		player_dev_labels_context.apply_gameplay_state(state)
	if world_telemetry_context != null:
		world_telemetry_context.apply_gameplay_state(state)
