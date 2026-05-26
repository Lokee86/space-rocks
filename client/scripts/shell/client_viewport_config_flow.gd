class_name ClientViewportConfigFlow
extends RefCounted

const Packets := preload("res://scripts/networking/packets/packets.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")

var connection_service
var viewport: Viewport


func configure(connection_service_ref, viewport_ref: Viewport) -> void:
	connection_service = connection_service_ref
	viewport = viewport_ref
	if viewport != null && !viewport.size_changed.is_connected(send_client_config):
		viewport.size_changed.connect(send_client_config)


func send_client_config() -> void:
	if connection_service == null || !connection_service.is_server_connected():
		return

	var viewport_size := viewport.get_visible_rect().size
	var packet := Packets.client_config_packet(viewport_size.x, viewport_size.y)
	connection_service.send_packet(packet)
	ClientLogger.shell_debug("V2 sent client viewport config: size=%s" % viewport_size)
