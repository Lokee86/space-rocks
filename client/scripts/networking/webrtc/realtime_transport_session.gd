class_name RealtimeTransportSession
extends RefCounted

const WebRTCTransport := preload("res://scripts/networking/webrtc/webrtc_transport.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")

var transport: WebRTCTransport
var transport_factory: Callable
var dispatch_packet: Callable
var send_offer: Callable
var send_ice_candidate: Callable
var send_failed: Callable
var _smoke_sequence := 0

func start() -> void:
	if transport != null:
		return
	transport = _create_transport()
	if transport == null:
		return
	if !transport.offer_created.is_connected(_on_offer_created):
		transport.offer_created.connect(_on_offer_created)
	if !transport.ice_candidate_created.is_connected(_on_ice_candidate_created):
		transport.ice_candidate_created.connect(_on_ice_candidate_created)
	if !transport.failed.is_connected(_on_failed):
		transport.failed.connect(_on_failed)
	if !transport.ready.is_connected(_on_ready):
		transport.ready.connect(_on_ready)
	if !transport.packet_received.is_connected(_on_packet_received):
		transport.packet_received.connect(_on_packet_received)
	transport.start()

func _on_ready(channels: Array) -> void:
	var smoke_id := _next_smoke_id()
	ClientLogger.network_event(
		ClientLogger.LEVEL_INFO,
		"webrtc_data_channel_ready",
		"WebRTC smoke data channels ready",
		{
			"channels": channels,
			"smoke_id": smoke_id,
		}
	)
	if transport != null:
		transport.send_smoke(smoke_id, "client smoke peer ready")

func _on_packet_received(packet: Dictionary) -> void:
	var packet_type := str(packet.get("type", ""))
	if packet_type == "webrtc_smoke":
		ClientLogger.network_event(
			ClientLogger.LEVEL_INFO,
			"webrtc_transport_smoke_received",
			"WebRTC smoke packet received by client",
			packet
		)
		return
	ClientLogger.network_event(
		ClientLogger.LEVEL_DEBUG,
		"webrtc_transport_packet_dispatch",
		"WebRTC transport packet dispatched",
		{
			"transport": "webrtc",
			"packet_type": packet_type,
		}
	)
	if dispatch_packet.is_valid():
		dispatch_packet.call(packet)
func handle_remote_ice(media: String, index: int, name: String) -> void:
	if transport == null:
		return
	transport.handle_remote_ice(media, index, name)

func handle_remote_failure() -> void:
	close()

func handle_answer(description_type: String, sdp: String) -> void:
	if transport == null:
		return
	transport.handle_answer(description_type, sdp)

func _on_offer_created(description_type: String, sdp: String) -> void:
	if send_offer.is_valid():
		send_offer.call(description_type, sdp)

func _on_ice_candidate_created(media: String, index: int, name: String) -> void:
	if send_ice_candidate.is_valid():
		send_ice_candidate.call(media, index, name)

func _on_failed(error_code: String, message: String) -> void:
	if send_failed.is_valid():
		send_failed.call(error_code, message)

func poll() -> void:
	if transport == null:
		return
	transport.poll()

func close() -> void:
	if transport != null:
		transport.close()
		transport = null

func _next_smoke_id() -> String:
	_smoke_sequence += 1
	return "client-smoke-%d" % _smoke_sequence

func _create_transport() -> WebRTCTransport:
	if transport_factory.is_valid():
		return transport_factory.call()
	return WebRTCTransport.new()
