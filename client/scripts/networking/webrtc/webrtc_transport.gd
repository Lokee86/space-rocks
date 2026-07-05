class_name WebRTCTransport
extends RefCounted

const GAMEPLAY_CHANNEL_SPECS := [
	{"lane": "world", "label": "sr.world", "id": 1},
	{"lane": "overlay", "label": "sr.overlay", "id": 2},
	{"lane": "session", "label": "sr.session", "id": 3},
	{"lane": "event", "label": "sr.event", "id": 4},
]
const SMOKE_ORIGIN_CLIENT := "client"
const PacketCodec := preload("res://scripts/networking/packets/packet_codec.gd")
const ClientConstants := preload("res://scripts/generated/constants/constants.gd")

signal offer_created(description_type: String, sdp: String)
signal ice_candidate_created(media: String, index: int, name: String)
signal ready(channels: Array)
signal packet_received(packet: Dictionary)
signal smoke_received(packet: Dictionary)
signal failed(error_code: String, message: String)

var peer_factory: Callable
var ice_servers: Array = []
var _peer: Variant
var _channels: Dictionary = {}
var _ready_channels: Dictionary = {}
var _ready_emitted := false


func set_peer_for_tests(peer: Variant, channels: Variant) -> void:
	_peer = peer
	_channels = {}
	if channels is Dictionary:
		_channels = channels.duplicate(true)
	elif channels != null:
		_channels["world"] = channels
	_ready_channels = {}
	_ready_emitted = false


func configure_ice_servers(servers: Array) -> void:
	ice_servers = servers.duplicate(true)


func set_ice_servers_for_tests(servers: Array) -> void:
	configure_ice_servers(servers)


func start() -> void:
	_peer = _create_peer()
	var init_config: Dictionary = _build_initialize_config()
	var init_result: int = int(_peer.initialize(init_config))
	if init_result != OK:
		failed.emit("peer_init_failed", "WebRTC peer initialization failed")
		return

	_peer.session_description_created.connect(_on_session_description_created)
	_peer.ice_candidate_created.connect(_on_ice_candidate_created)
	_channels = {}
	_ready_channels = {}
	for spec in GAMEPLAY_CHANNEL_SPECS:
		var lane: String = str(spec.get("lane", ""))
		var label: String = str(spec.get("label", ""))
		var channel_id: int = int(spec.get("id", 0))
		var channel = _peer.create_data_channel(label, {
			"id": channel_id,
			"negotiated": true,
			"ordered": true,
		})
		if channel == null:
			failed.emit("channel_create_failed", "WebRTC data channel creation failed")
			return
		channel.write_mode = WebRTCDataChannel.WRITE_MODE_TEXT
		_channels[lane] = channel
		_ready_channels[lane] = false
	_peer.create_offer()



func _create_peer() -> Variant:
	if peer_factory.is_valid():
		return peer_factory.call()
	return WebRTCPeerConnection.new()


func _build_initialize_config() -> Dictionary:
	var config: Dictionary = {}
	var effective_ice_servers: Array = ice_servers
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
	if _channels.is_empty():
		return
	var all_ready: bool = true
	for spec in GAMEPLAY_CHANNEL_SPECS:
		var lane: String = str(spec.get("lane", ""))
		var channel: Variant = _channels.get(lane)
		if channel == null:
			all_ready = false
			continue
		var channel_ready: bool = channel.get_ready_state() == WebRTCDataChannel.STATE_OPEN
		_ready_channels[lane] = channel_ready
		all_ready = all_ready and channel_ready
	if !_ready_emitted and all_ready:
		_ready_emitted = true
		ready.emit(_gameplay_channel_ready_payload())
	for spec in GAMEPLAY_CHANNEL_SPECS:
		var lane: String = str(spec.get("lane", ""))
		var channel: Variant = _channels.get(lane)
		if channel == null:
			continue
		if channel.get_ready_state() != WebRTCDataChannel.STATE_OPEN:
			continue
		while channel.get_available_packet_count() > 0:
			var raw_packet: PackedByteArray = channel.get_packet()
			if raw_packet is PackedByteArray:
				_handle_channel_packet(raw_packet)


func send_json(packet: Dictionary) -> void:
	_send_json_to_lane("world", packet)


func send_smoke(smoke_id: String, message: String) -> void:
	send_json({
		"type": "webrtc_smoke",
		"smoke_id": smoke_id,
		"origin": SMOKE_ORIGIN_CLIENT,
		"message": message,
	})


func close() -> void:
	for lane in _channels.keys():
		var channel: Variant = _channels.get(lane)
		if channel != null:
			channel.close()
	if _peer != null:
		_peer.close()
	_peer = null
	_channels = {}
	_ready_channels = {}
	_ready_emitted = false


func _on_session_description_created(description_type: String, sdp: String) -> void:
	if _peer != null:
		_peer.set_local_description(description_type, sdp)
	offer_created.emit(description_type, sdp)


func _on_ice_candidate_created(media: String, index: int, name: String) -> void:
	ice_candidate_created.emit(media, index, name)


func _handle_channel_packet(packet: PackedByteArray) -> void:
	var text: String = packet.get_string_from_utf8()
	var decode_result: Variant = PacketCodec.decode(text)
	if !decode_result.ok:
		failed.emit("invalid_json", decode_result.error)
		return
	var data: Dictionary = decode_result.packet
	packet_received.emit(data)
	if str(data.get("type", "")) == "webrtc_smoke":
		smoke_received.emit(data)


func _send_json_to_lane(lane: String, packet: Dictionary) -> void:
	var channel: Variant = _channels.get(lane)
	if channel == null or channel.get_ready_state() != WebRTCDataChannel.STATE_OPEN:
		return
	channel.put_packet(JSON.stringify(packet).to_utf8_buffer())


func _gameplay_channel_spec_for_lane(lane: String) -> Variant:
	for spec in GAMEPLAY_CHANNEL_SPECS:
		if str(spec.get("lane", "")) == lane:
			return spec
	return null


func _gameplay_channel_ready_payload() -> Array:
	var channels: Array = []
	for spec in GAMEPLAY_CHANNEL_SPECS:
		channels.append({
			"lane": str(spec.get("lane", "")),
			"channel_label": str(spec.get("label", "")),
			"channel_id": int(spec.get("id", 0)),
		})
	return channels
