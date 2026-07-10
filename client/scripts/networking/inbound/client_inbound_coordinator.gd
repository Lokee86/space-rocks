class_name ClientInboundCoordinator
extends RefCounted

const ClientLogger := preload("res://scripts/logging/logger.gd")

signal realtime_transport_ready

var server_packet_dispatcher
var realtime_transport_session


func configure(dispatcher, transport_session = null) -> void:
	server_packet_dispatcher = dispatcher
	set_realtime_transport_session(transport_session)
	_connect_dispatcher_signal("webrtc_answer_received", Callable(self, "handle_webrtc_answer"))
	_connect_dispatcher_signal("webrtc_ice_candidate_received", Callable(self, "handle_webrtc_ice_candidate"))
	_connect_dispatcher_signal("webrtc_ready_received", Callable(self, "handle_webrtc_ready"))
	_connect_dispatcher_signal("webrtc_smoke_received", Callable(self, "handle_webrtc_smoke"))
	_connect_dispatcher_signal("webrtc_failed_received", Callable(self, "handle_webrtc_failed"))


func set_realtime_transport_session(session) -> void:
	realtime_transport_session = session


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


func handle_webrtc_ready(packet: Dictionary) -> void:
	ClientLogger.network_event(
		ClientLogger.LEVEL_INFO,
		"webrtc_ready_received",
		"WebRTC ready packet received",
		packet
	)
	realtime_transport_ready.emit()


func handle_webrtc_smoke(packet: Dictionary) -> void:
	ClientLogger.network_event(
		ClientLogger.LEVEL_INFO,
		"webrtc_smoke_received",
		"WebRTC smoke packet received",
		packet
	)


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
