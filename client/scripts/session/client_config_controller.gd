extends RefCounted


var client_viewport_config_flow


func configure(connection_service_ref, viewport_ref: Viewport) -> void:
	client_viewport_config_flow = ClientViewportConfigFlow.new()
	client_viewport_config_flow.configure(connection_service_ref, viewport_ref)


func send_client_config() -> void:
	if client_viewport_config_flow != null:
		client_viewport_config_flow.send_client_config()
