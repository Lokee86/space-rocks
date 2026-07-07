class_name WebRTCTransport
extends RefCounted

const GAMEPLAY_CHANNEL_SPECS := [
	{"lane": "world", "label": "sr.world", "id": 1, "ordered": true},
	{"lane": "overlay", "label": "sr.overlay", "id": 2, "ordered": true},
	{"lane": "session", "label": "sr.session", "id": 3, "ordered": true},
	{"lane": "event", "label": "sr.event", "id": 4, "ordered": true},
	{"lane": "asteroids", "label": "sr.asteroids", "id": 5, "ordered": false, "max_retransmits": 0},
	{"lane": "bullets", "label": "sr.bullets", "id": 6, "ordered": false, "max_retransmits": 0},
	{"lane": "asteroids_lifecycle", "label": "sr.asteroids.lifecycle", "id": 7, "ordered": true},
	{"lane": "bullets_lifecycle", "label": "sr.bullets.lifecycle", "id": 8, "ordered": true},
]
const MAX_PACKETS_PER_POLL := 48
const MAX_PACKETS_PER_LANE_PER_POLL := 12
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
var bullet_delta_received_count := 0
var bullet_delta_max_age_msec := 0
var bullet_delta_last_age_msec := 0
var bullet_delta_missing_server_time_count := 0
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
		var channel_options: Dictionary = {
			"id": channel_id,
			"negotiated": true,
			"ordered": bool(spec.get("ordered", true)),
		}
		if spec.has("max_retransmits"):
			channel_options["maxRetransmits"] = int(spec.get("max_retransmits"))
		var channel = _peer.create_data_channel(label, channel_options)
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
	var total_drained: int = 0
	for spec in GAMEPLAY_CHANNEL_SPECS:
		if total_drained >= MAX_PACKETS_PER_POLL:
			break
		var lane: String = str(spec.get("lane", ""))
		var channel: Variant = _channels.get(lane)
		if channel == null:
			continue
		if channel.get_ready_state() != WebRTCDataChannel.STATE_OPEN:
			continue
		var drained_for_lane: int = _drain_channel_packets(channel, MAX_PACKETS_PER_POLL - total_drained)
		total_drained += drained_for_lane

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
	if str(data.get("type", "")) == "bullet_delta":
		bullet_delta_received_count += 1
		var server_sent_msec = data.get("server_sent_msec", null)
		if server_sent_msec != null:
			var age_msec := Time.get_ticks_msec() - int(server_sent_msec)
			bullet_delta_last_age_msec = age_msec
			bullet_delta_max_age_msec = maxi(bullet_delta_max_age_msec, age_msec)
		else:
			bullet_delta_missing_server_time_count += 1
	packet_received.emit(data)
	if str(data.get("type", "")) == "webrtc_smoke":
		smoke_received.emit(data)


func _drain_channel_packets(channel: Variant, remaining_budget: int) -> int:
	var drained: int = 0
	var lane_budget: int = int(min(MAX_PACKETS_PER_LANE_PER_POLL, remaining_budget))
	while drained < lane_budget and channel.get_available_packet_count() > 0:
		var raw_packet: PackedByteArray = channel.get_packet()
		if raw_packet is PackedByteArray:
			_handle_channel_packet(raw_packet)
			drained += 1
	return drained

func _send_json_to_lane(lane: String, packet: Dictionary) -> void:
	var channel: Variant = _channels.get(lane)
	if channel == null or channel.get_ready_state() != WebRTCDataChannel.STATE_OPEN:
		return
	channel.put_packet(JSON.stringify(packet).to_utf8_buffer())


func receive_metrics_snapshot() -> Dictionary:
	return {
		"bullet_delta_received_count": bullet_delta_received_count,
		"bullet_delta_last_age_msec": bullet_delta_last_age_msec,
		"bullet_delta_max_age_msec": bullet_delta_max_age_msec,
		"bullet_delta_missing_server_time_count": bullet_delta_missing_server_time_count,
	}


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