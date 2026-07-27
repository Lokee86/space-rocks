class_name ClientInboundCoordinator
extends RefCounted


signal realtime_transport_ready
signal authenticate_result_received(packet: Dictionary)
signal room_snapshot_received(packet: Dictionary)
signal room_state_changed(packet: Dictionary)
signal room_error_received(packet: Dictionary)
signal player_pause_state_received(packet: Dictionary)
signal unknown_packet_received(packet: Dictionary)

var server_packet_dispatcher
var realtime_packet_pipeline
var realtime_transport_session


func configure(dispatcher, pipeline, transport_session = null) -> void:
	server_packet_dispatcher = dispatcher
	realtime_packet_pipeline = pipeline
	set_realtime_transport_session(transport_session)
	_connect_dispatcher_signal("authenticate_result_received", Callable(self, "_on_authenticate_result_received"))
	_connect_dispatcher_signal("room_snapshot_received", Callable(self, "_on_room_snapshot_received"))
	_connect_dispatcher_signal("room_state_changed", Callable(self, "_on_room_state_changed"))
	_connect_dispatcher_signal("room_error_received", Callable(self, "_on_room_error_received"))
	_connect_dispatcher_signal("player_pause_state_received", Callable(self, "_on_player_pause_state_received"))
	_connect_dispatcher_signal("unknown_packet_received", Callable(self, "_on_unknown_packet_received"))
	_connect_dispatcher_signal("world_full_received", Callable(realtime_packet_pipeline, "apply_world_full"))
	_connect_dispatcher_signal("world_delta_received", Callable(realtime_packet_pipeline, "apply_world_delta"))
	_connect_dispatcher_signal("ship_delta_received", Callable(realtime_packet_pipeline, "apply_ship_delta"))
	_connect_dispatcher_signal("player_locator_received", Callable(realtime_packet_pipeline, "apply_player_locator"))
	_connect_dispatcher_signal("ships_lifecycle_received", Callable(realtime_packet_pipeline, "apply_ships_lifecycle"))
	_connect_dispatcher_signal("asteroid_delta_received", Callable(realtime_packet_pipeline, "apply_asteroid_delta"))
	_connect_dispatcher_signal("bullet_delta_received", Callable(realtime_packet_pipeline, "apply_bullet_delta"))
	_connect_dispatcher_signal("asteroids_lifecycle_received", Callable(realtime_packet_pipeline, "apply_asteroids_lifecycle"))
	_connect_dispatcher_signal("bullets_lifecycle_received", Callable(realtime_packet_pipeline, "apply_bullets_lifecycle"))
	_connect_dispatcher_signal("overlay_full_received", Callable(realtime_packet_pipeline, "apply_overlay_full"))
	_connect_dispatcher_signal("overlay_delta_received", Callable(realtime_packet_pipeline, "apply_overlay_delta"))
	_connect_dispatcher_signal("session_full_received", Callable(realtime_packet_pipeline, "apply_session_full"))
	_connect_dispatcher_signal("session_delta_received", Callable(realtime_packet_pipeline, "apply_session_delta"))
	_connect_dispatcher_signal("event_batch_received", Callable(realtime_packet_pipeline, "apply_event_batch"))
	_connect_dispatcher_signal("resync_request_received", Callable(realtime_packet_pipeline, "apply_resync_request"))
	_connect_dispatcher_signal("resync_required_received", Callable(realtime_packet_pipeline, "apply_resync_required"))
	_connect_dispatcher_signal("webrtc_answer_received", Callable(self, "handle_webrtc_answer"))
	_connect_dispatcher_signal("webrtc_ice_candidate_received", Callable(self, "handle_webrtc_ice_candidate"))
	_connect_dispatcher_signal("webrtc_ready_received", Callable(self, "handle_webrtc_ready"))
	_connect_dispatcher_signal("webrtc_smoke_received", Callable(self, "handle_webrtc_smoke"))
	_connect_dispatcher_signal("webrtc_failed_received", Callable(self, "handle_webrtc_failed"))


func set_realtime_transport_session(session) -> void:
	realtime_transport_session = session


func _on_authenticate_result_received(packet: Dictionary) -> void:
	authenticate_result_received.emit(packet)


func _on_room_snapshot_received(packet: Dictionary) -> void:
	room_snapshot_received.emit(packet)


func _on_room_state_changed(packet: Dictionary) -> void:
	room_state_changed.emit(packet)


func _on_room_error_received(packet: Dictionary) -> void:
	room_error_received.emit(packet)



func _on_player_pause_state_received(packet: Dictionary) -> void:
	player_pause_state_received.emit(packet)


func _on_unknown_packet_received(packet: Dictionary) -> void:
	unknown_packet_received.emit(packet)


func handle_webrtc_answer(packet: Dictionary) -> void:
	if realtime_transport_session == null:
		return
	realtime_transport_session.handle_answer(
		str(packet.get("description_type", "")),
		str(packet.get("sdp", ""))
	)


func handle_webrtc_ice_candidate(packet: Dictionary) -> void:
	if realtime_transport_session == null:
		return
	realtime_transport_session.handle_remote_ice(
		str(packet.get("media", "")),
		int(packet.get("index", 0)),
		str(packet.get("name", ""))
	)


func handle_webrtc_ready(_packet: Dictionary) -> void:
	realtime_transport_ready.emit()


func handle_webrtc_smoke(_packet: Dictionary) -> void:
	pass


func handle_webrtc_failed(_packet: Dictionary) -> void:
	if realtime_transport_session == null:
		return
	realtime_transport_session.handle_remote_failure()


func _connect_dispatcher_signal(signal_name: StringName, handler: Callable) -> void:
	if server_packet_dispatcher == null:
		return
	if !server_packet_dispatcher.has_signal(signal_name):
		return
	if server_packet_dispatcher.is_connected(signal_name, handler):
		return
	server_packet_dispatcher.connect(signal_name, handler)
