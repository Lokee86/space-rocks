class_name ClientViewportConfigFlow
extends RefCounted

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")

var connection_service
var viewport: Viewport
var visible_world_size_provider := Callable()


func configure(connection_service_ref, viewport_ref: Viewport) -> void:
	connection_service = connection_service_ref
	viewport = viewport_ref
	if viewport != null && !viewport.size_changed.is_connected(send_client_config):
		viewport.size_changed.connect(send_client_config)


func configure_visible_world_size_provider(provider: Callable) -> void:
	visible_world_size_provider = provider


func send_client_config() -> void:
	if connection_service == null || !connection_service.is_server_connected():
		return

	var visible_world_size := _visible_world_size()
	var packet := Packets.client_config_packet(visible_world_size.x, visible_world_size.y)
	connection_service.send_packet(packet)


func _visible_world_size() -> Vector2:
	if visible_world_size_provider.is_valid():
		var provided_size = visible_world_size_provider.call()
		if provided_size is Vector2 && provided_size.x > 0 && provided_size.y > 0:
			return provided_size
	if viewport != null:
		return viewport.get_visible_rect().size
	return Vector2(1280.0, 720.0)

