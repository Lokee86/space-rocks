class_name WebRTCTransport
extends RefCounted

const CHANNEL_SPECS := [
	{"lane": "world", "label": "sr.world", "id": 1, "ordered": true},
	{"lane": "overlay", "label": "sr.overlay", "id": 2, "ordered": true},
	{"lane": "session", "label": "sr.session", "id": 3, "ordered": true},
	{"lane": "event", "label": "sr.event", "id": 4, "ordered": true},
	{"lane": "asteroids", "label": "sr.asteroids", "id": 5, "ordered": false, "max_retransmits": 0},
	{"lane": "bullets", "label": "sr.bullets", "id": 6, "ordered": false, "max_retransmits": 0},
	{"lane": "asteroids_lifecycle", "label": "sr.asteroids.lifecycle", "id": 7, "ordered": true},
	{"lane": "bullets_lifecycle", "label": "sr.bullets.lifecycle", "id": 8, "ordered": true},
	{"lane": "tooling", "label": "sr.tooling", "id": 9, "ordered": true},
]
const MAX_PACKETS_PER_POLL := 48
const MAX_PACKETS_PER_LANE_PER_POLL := 12
const SMOKE_ORIGIN_CLIENT := "client"
const PacketCodec := preload("res://scripts/networking/packets/packet_codec.gd")

const ClientConstants := preload("res://scripts/generated/constants/constants.gd")

signal offer_created(description_type: String, sdp: String)
signal ice_candidate_created(media: String, index: int, name: String)
signal ready(channels: Array)
signal packet_received(packet: Dictionary, lane: String)
signal smoke_received(packet: Dictionary)
signal channel_closed(lane: String)
signal failed(error_code: String, message: String)

var peer_factory: Callable
var ice_servers: Array = []
var bullet_delta_received_count := 0
var bullet_delta_max_age_msec := -1
var bullet_delta_last_age_msec := -1
var bullet_delta_missing_server_time_count := 0
var bullet_delta_unsynchronized_server_time_count := 0
var bullet_delta_clock_skew_count := 0
var server_clock_offset_ms := -1
var _peer: Variant
var _channels: Dictionary = {}
var _ready_channels: Dictionary = {}
var _ready_emitted := false
var _channel_close_reported: Dictionary = {}
var _lifecycle_start_cursor := 0
var _general_start_cursor := 0
var _runtime_metrics = NetworkRuntimeMetrics.new()
var _lane_metrics: Dictionary = {}

const LIFECYCLE_LANES := ["asteroids_lifecycle", "bullets_lifecycle"]
const GENERAL_LANES := ["world", "overlay", "session", "event", "asteroids", "bullets", "tooling"]


func set_peer_for_tests(peer: Variant, channels: Variant) -> void:
	_peer = peer
	_channels = {}
	if channels is Dictionary:
		_channels = channels.duplicate(true)
	elif channels != null:
		_channels["world"] = channels
	_ready_channels = {}
	_ready_emitted = false
	_channel_close_reported = {}
	_lifecycle_start_cursor = 0
	_general_start_cursor = 0
	_reset_runtime_metrics()


func configure_ice_servers(servers: Array) -> void:
	ice_servers = servers.duplicate(true)


func set_server_clock_offset_ms(offset_ms: int) -> void:
	server_clock_offset_ms = offset_ms


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
	_channel_close_reported = {}
	_reset_runtime_metrics()
	for spec in CHANNEL_SPECS:
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
	for spec in CHANNEL_SPECS:
		var lane: String = str(spec.get("lane", ""))
		var channel: Variant = _channels.get(lane)
		var channel_ready: bool = channel != null and channel.get_ready_state() == WebRTCDataChannel.STATE_OPEN
		if _ready_channels.get(lane, false) and !channel_ready and !_channel_close_reported.get(lane, false):
			_channel_close_reported[lane] = true
			channel_closed.emit(lane)
		if channel == null:
			all_ready = false
			continue
		_ready_channels[lane] = channel_ready
		all_ready = all_ready and channel_ready
	if !_ready_emitted and all_ready:
		_ready_emitted = true
		ready.emit(_channel_ready_payload())
	var total_drained := 0
	var drained_by_lane := {}
	while total_drained < MAX_PACKETS_PER_POLL:
		var drained_this_pass := 0
		for lanes_data in [[LIFECYCLE_LANES, _lifecycle_start_cursor], [GENERAL_LANES, _general_start_cursor]]:
			var lanes: Array = lanes_data[0]
			var start_cursor: int = lanes_data[1]
			for offset in lanes.size():
				if total_drained >= MAX_PACKETS_PER_POLL:
					break
				var lane: String = str(lanes[(start_cursor + offset) % lanes.size()])
				var channel: Variant = _channels.get(lane)
				if channel == null or channel.get_ready_state() != WebRTCDataChannel.STATE_OPEN:
					continue
				var lane_drained: int = int(drained_by_lane.get(lane, 0))
				if lane_drained >= MAX_PACKETS_PER_LANE_PER_POLL or channel.get_available_packet_count() <= 0:
					continue
				var raw_packet: PackedByteArray = channel.get_packet()
				if raw_packet is PackedByteArray:
					_handle_channel_packet(raw_packet, lane)
					drained_by_lane[lane] = lane_drained + 1
					total_drained += 1
					drained_this_pass += 1
		if drained_this_pass == 0:
			break
	_lifecycle_start_cursor = (_lifecycle_start_cursor + 1) % LIFECYCLE_LANES.size()
	_general_start_cursor = (_general_start_cursor + 1) % GENERAL_LANES.size()

func send_json(packet: Dictionary) -> void:
	_send_json_to_lane("world", packet)


func send_tooling_json(packet: Dictionary) -> void:
	_send_json_to_lane("tooling", packet)


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
	_channel_close_reported = {}
	_lifecycle_start_cursor = 0
	_general_start_cursor = 0
	_reset_runtime_metrics()


func _on_session_description_created(description_type: String, sdp: String) -> void:
	if _peer != null:
		_peer.set_local_description(description_type, sdp)
	offer_created.emit(description_type, sdp)


func _on_ice_candidate_created(media: String, index: int, name: String) -> void:
	ice_candidate_created.emit(media, index, name)


func _handle_channel_packet(packet: PackedByteArray, lane: String = "") -> void:
	var raw_bytes := packet.size()
	var text: String = packet.get_string_from_utf8()
	var decode_result: Variant = PacketCodec.decode(text)
	if !decode_result.ok:
		_runtime_metrics.observe_decode_failure(raw_bytes)
		_metrics_for_lane(lane).observe_decode_failure(raw_bytes)
		failed.emit("invalid_json", decode_result.error)
		return
	var packet_type := _packet_type(decode_result.packet)
	_runtime_metrics.observe_inbound(raw_bytes, packet_type)
	_metrics_for_lane(lane).observe_inbound(raw_bytes, packet_type)
	_handle_decoded_packet(decode_result.packet, lane)


func _handle_decoded_packet(data: Dictionary, lane: String = "") -> void:
	if str(data.get("type", "")) == "bullet_delta":
		bullet_delta_received_count += 1
		var server_sent_msec = data.get("server_sent_msec", null)
		if server_sent_msec == null or int(server_sent_msec) <= 0:
			bullet_delta_missing_server_time_count += 1
		elif server_clock_offset_ms == -1:
			bullet_delta_unsynchronized_server_time_count += 1
		else:
			var client_sent_msec := int(server_sent_msec) - server_clock_offset_ms
			var age_msec := Time.get_ticks_msec() - client_sent_msec
			if age_msec < 0:
				bullet_delta_clock_skew_count += 1
				age_msec = 0
			bullet_delta_last_age_msec = age_msec
			bullet_delta_max_age_msec = maxi(bullet_delta_max_age_msec, age_msec)
	if str(data.get("type", "")) == "webrtc_smoke":
		smoke_received.emit(data)
		return
	packet_received.emit(data, lane)


func _send_json_to_lane(lane: String, packet: Dictionary) -> void:
	var channel: Variant = _channels.get(lane)
	if channel == null or channel.get_ready_state() != WebRTCDataChannel.STATE_OPEN:
		return
	var encoded := JSON.stringify(packet).to_utf8_buffer()
	var packet_type := _packet_type(packet)
	var result: int = int(channel.put_packet(encoded))
	if result != OK:
		_runtime_metrics.observe_send_failure(encoded.size(), packet_type)
		_metrics_for_lane(lane).observe_send_failure(encoded.size(), packet_type)
		return
	_runtime_metrics.observe_outbound(encoded.size(), packet_type)
	_metrics_for_lane(lane).observe_outbound(encoded.size(), packet_type)


func network_metrics_snapshot() -> Dictionary:
	var snapshot: Dictionary = _runtime_metrics.snapshot()
	var lanes := {}
	for lane in _lane_metrics.keys():
		lanes[lane] = _lane_metrics[lane].snapshot()
	snapshot["lanes"] = lanes
	snapshot["bullet_delta_received_count"] = bullet_delta_received_count
	snapshot["bullet_delta_last_age_msec"] = bullet_delta_last_age_msec
	snapshot["bullet_delta_max_age_msec"] = bullet_delta_max_age_msec
	snapshot["bullet_delta_missing_server_time_count"] = bullet_delta_missing_server_time_count
	snapshot["bullet_delta_unsynchronized_server_time_count"] = bullet_delta_unsynchronized_server_time_count
	snapshot["bullet_delta_clock_skew_count"] = bullet_delta_clock_skew_count
	return snapshot


func receive_metrics_snapshot() -> Dictionary:
	return network_metrics_snapshot()


func _reset_runtime_metrics() -> void:
	_runtime_metrics = NetworkRuntimeMetrics.new()
	_runtime_metrics.transport = "webrtc"
	_lane_metrics.clear()


func _metrics_for_lane(lane: String):
	if not _lane_metrics.has(lane):
		var metrics = NetworkRuntimeMetrics.new()
		metrics.transport = "webrtc:%s" % lane
		_lane_metrics[lane] = metrics
	return _lane_metrics[lane]


func _packet_type(packet: Dictionary) -> String:
	var packet_type := str(packet.get("type", ""))
	if not packet_type.is_empty():
		return packet_type
	return str(packet.get("t", ""))


func _channel_spec_for_lane(lane: String) -> Variant:
	for spec in CHANNEL_SPECS:
		if str(spec.get("lane", "")) == lane:
			return spec
	return null

func _channel_ready_payload() -> Array:
	var channels: Array = []
	for spec in CHANNEL_SPECS:
		channels.append({
			"lane": str(spec.get("lane", "")),
			"channel_label": str(spec.get("label", "")),
			"channel_id": int(spec.get("id", 0)),
		})
	return channels
