class_name WebRTCSmokePeer
extends RefCounted

const CHANNEL_LABEL := "sr.reliable"
const CHANNEL_ID := 1
const SMOKE_ORIGIN_CLIENT := "client"

signal offer_created(description_type: String, sdp: String)
signal ice_candidate_created(media: String, index: int, name: String)
signal ready(channel_label: String, channel_id: int)
signal smoke_received(packet: Dictionary)
signal failed(error_code: String, message: String)

var peer_factory: Callable
var _peer: Variant
var _channel: Variant
var _ready_emitted := false


func set_peer_for_tests(peer: Variant, channel: Variant) -> void:
	_peer = peer
	_channel = channel
	_ready_emitted = false


func start() -> void:
	_peer = _create_peer()
	var init_result: int = int(_peer.initialize())
	if init_result != OK:
		failed.emit("peer_init_failed", "WebRTC peer initialization failed")
		return

	_peer.session_description_created.connect(_on_session_description_created)
	_peer.ice_candidate_created.connect(_on_ice_candidate_created)
	_channel = _peer.create_data_channel(CHANNEL_LABEL, {
		"id": CHANNEL_ID,
		"negotiated": true,
		"ordered": true,
	})
	if _channel == null:
		failed.emit("channel_create_failed", "WebRTC data channel creation failed")
		return
	_channel.write_mode = WebRTCDataChannel.WRITE_MODE_TEXT
	_peer.create_offer()



func _create_peer() -> Variant:
	if peer_factory.is_valid():
		return peer_factory.call()
	return WebRTCPeerConnection.new()

func handle_answer(description_type: String, sdp: String) -> void:
	if _peer == null:
		return
	_peer.set_remote_description(description_type, sdp)


func handle_remote_ice(media: String, index: int, name: String) -> void:
	if _peer == null:
		return
	_peer.add_ice_candidate(media, index, name)


func poll() -> void:
	if _peer == null:
		return
	_peer.poll()
	if _channel == null:
		return
	if !_ready_emitted and _channel.get_ready_state() == WebRTCDataChannel.STATE_OPEN:
		_ready_emitted = true
		ready.emit(CHANNEL_LABEL, CHANNEL_ID)
	while _channel.get_available_packet_count() > 0:
		var raw_packet = _channel.get_packet()
		if raw_packet is PackedByteArray:
			_handle_channel_packet(raw_packet)


func send_smoke(smoke_id: String, message: String) -> void:
	if _channel == null or _channel.get_ready_state() != WebRTCDataChannel.STATE_OPEN:
		return
	_channel.put_packet(JSON.stringify({
		"type": "webrtc_smoke",
		"smoke_id": smoke_id,
		"origin": SMOKE_ORIGIN_CLIENT,
		"message": message,
	}).to_utf8_buffer())


func close() -> void:
	if _channel != null:
		_channel.close()
	if _peer != null:
		_peer.close()
	_peer = null
	_channel = null
	_ready_emitted = false


func _on_session_description_created(description_type: String, sdp: String) -> void:
	if _peer != null:
		_peer.set_local_description(description_type, sdp)
	offer_created.emit(description_type, sdp)


func _on_ice_candidate_created(media: String, index: int, name: String) -> void:
	ice_candidate_created.emit(media, index, name)


func _handle_channel_packet(packet: PackedByteArray) -> void:
	var text := packet.get_string_from_utf8()
	var json := JSON.new()
	var parse_error := json.parse(text)
	if parse_error != OK:
		failed.emit("invalid_json", "WebRTC smoke packet was not valid JSON")
		return
	var data: Dictionary = json.get_data()
	if str(data.get("type", "")) != "webrtc_smoke":
		return
	smoke_received.emit(data)
