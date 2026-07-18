class_name RealtimeTransportSession
extends RefCounted

const WebRTCTransport := preload("res://scripts/networking/webrtc/webrtc_transport.gd")

const RECOVERY_TIMEOUT_MSEC := 10000

signal tooling_packet_received(packet: Dictionary)
signal recovery_started(lane: String)
signal recovery_succeeded()
signal recovery_failed()

var transport: WebRTCTransport
var transport_factory: Callable
var dispatch_packet: Callable
var send_offer: Callable
var send_ice_candidate: Callable
var send_failed: Callable
var server_clock_offset_ms := -1
var _clock: Callable
var _smoke_sequence := 0
var _recovery_active := false
var _recovery_failed := false
var _recovery_deadline_msec := -1

func _init(clock: Callable = Callable(Time, "get_ticks_msec")) -> void:
	_clock = clock

func start() -> void:
	if transport != null:
		return
	if !_start_transport():
		return

func _start_transport() -> bool:
	transport = _create_transport()
	if transport == null:
		return false
	transport.set_server_clock_offset_ms(server_clock_offset_ms)
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
	if !transport.channel_closed.is_connected(_on_channel_closed):
		transport.channel_closed.connect(_on_channel_closed)
	transport.start()
	return true

func _on_ready(_channels: Array) -> void:
	var smoke_id := _next_smoke_id()
	if transport != null:
		transport.send_smoke(smoke_id, "client smoke peer ready")
	if _recovery_active:
		_recovery_active = false
		_recovery_deadline_msec = -1
		recovery_succeeded.emit()

func _on_packet_received(packet: Dictionary, lane: String = "") -> void:
	if lane == "tooling":
		tooling_packet_received.emit(packet)
		return
	var packet_type := str(packet.get("type", ""))
	if packet_type == "webrtc_smoke":
		return
	if dispatch_packet.is_valid():
		dispatch_packet.call(packet)

func _on_channel_closed(lane: String) -> void:
	if _recovery_active or _recovery_failed:
		return
	_recovery_active = true
	_recovery_deadline_msec = _now_msec() + RECOVERY_TIMEOUT_MSEC
	recovery_started.emit(lane)
	_replace_transport_for_recovery()

func _replace_transport_for_recovery() -> void:
	if transport != null:
		transport.close()
		transport = null
	if !_start_transport():
		_fail_recovery()
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
	if _recovery_active:
		_fail_recovery()

func poll() -> void:
	if transport != null:
		transport.poll()
	if _recovery_active and _now_msec() >= _recovery_deadline_msec:
		_fail_recovery()

func send_tooling_packet(packet: Dictionary) -> void:
	if transport != null:
		transport.send_tooling_json(packet)

func _fail_recovery() -> void:
	if !_recovery_active:
		return
	_recovery_active = false
	_recovery_failed = true
	_recovery_deadline_msec = -1
	if transport != null:
		transport.close()
		transport = null
	recovery_failed.emit()

func close() -> void:
	if transport != null:
		transport.close()
		transport = null
	_recovery_active = false
	_recovery_failed = false
	_recovery_deadline_msec = -1

func set_server_clock_offset_ms(offset_ms: int) -> void:
	server_clock_offset_ms = offset_ms
	if transport != null:
		transport.set_server_clock_offset_ms(offset_ms)

func _next_smoke_id() -> String:
	_smoke_sequence += 1
	return "client-smoke-%d" % _smoke_sequence

func _now_msec() -> int:
	if _clock.is_valid():
		return int(_clock.call())
	return Time.get_ticks_msec()

func _create_transport() -> WebRTCTransport:
	if transport_factory.is_valid():
		return transport_factory.call()
	return WebRTCTransport.new()
