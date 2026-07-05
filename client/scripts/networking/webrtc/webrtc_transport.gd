class_name WebRTCTransport
extends RefCounted

const CHANNEL_LABEL := "sr.reliable"
const CHANNEL_ID := 1
const SMOKE_ORIGIN_CLIENT := "client"
const PacketCodec := preload("res://scripts/networking/packets/packet_codec.gd")
const ClientConstants := preload("res://scripts/generated/constants/constants.gd")

signal offer_created(description_type: String, sdp: String)
signal ice_candidate_created(media: String, index: int, name: String)
signal ready(channel_label: String, channel_id: int)
signal packet_received(packet: Dictionary)
signal smoke_received(packet: Dictionary)
signal failed(error_code: String, message: String)

var peer_factory: Callable
var ice_servers: Array = []
var _peer: Variant
var _channel: Variant
var _ready_emitted := false


func set_peer_for_tests(peer: Variant, channel: Variant) -> void:
	_peer = peer
	_channel = channel
	_ready_emitted = false


func configure_ice_servers(servers: Array) -> void:
	ice_servers = servers.duplicate(true)


func set_ice_servers_for_tests(servers: Array) -> void:
	configure_ice_servers(servers)


func start() -> void:
	_peer = _create_peer()
	var init_config := _build_initialize_config()
	var init_result: int = int(_peer.initialize(init_config))
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


func _build_initialize_config() -> Dictionary:
	var config := {}
	var effective_ice_servers := ice_servers
	if effective_ice_servers.is_empty():
		effective_ice_servers = ClientConstants.WEBRTC_ICE_SERVERS
	if !effective_ice_servers.is_empty():
		config["iceServers"] = effective_ice_servers
	return config

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


func send_json(packet: Dictionary) -> void:
	if _channel == null or _channel.get_ready_state() != WebRTCDataChannel.STATE_OPEN:
		return
	_channel.put_packet(JSON.stringify(packet).to_utf8_buffer())


func send_smoke(smoke_id: String, message: String) -> void:
	send_json({
		"type": "webrtc_smoke",
		"smoke_id": smoke_id,
		"origin": SMOKE_ORIGIN_CLIENT,
		"message": message,
	})


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
	var decode_result = PacketCodec.decode(text)
	if !decode_result.ok:
		failed.emit("invalid_json", decode_result.error)
		return
	var data: Dictionary = decode_result.packet
	packet_received.emit(data)
	if str(data.get("type", "")) == "webrtc_smoke":
		smoke_received.emit(data)
