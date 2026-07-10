extends GutTest

const WebRTCTransportScript := preload("res://scripts/networking/webrtc/webrtc_transport.gd")


class FakeChannel:
	extends RefCounted

	var write_mode = null
	var ready_state := WebRTCDataChannel.STATE_CLOSED
	var packets: Array[PackedByteArray] = []
	var sent_packets: Array[PackedByteArray] = []
	var closed := false

	func get_ready_state() -> int:
		return ready_state

	func get_available_packet_count() -> int:
		return packets.size()

	func get_packet() -> PackedByteArray:
		return packets.pop_front()

	func put_packet(packet: PackedByteArray) -> int:
		sent_packets.append(packet)
		return OK

	func close() -> void:
		closed = true

class FakePeer:
	extends RefCounted

	signal session_description_created(description_type: String, sdp: String)
	signal ice_candidate_created(media: String, index: int, name: String)

	var initialize_result := OK
	var initialize_args: Array = []
	var channels: Dictionary = {}
	var create_data_channel_args: Array = []
	var create_offer_called := 0
	var set_local_description_args: Array = []
	var set_remote_description_args: Array = []
	var add_ice_candidate_args: Array = []
	var poll_called := 0
	var closed := false

	func initialize(config: Dictionary = {}) -> int:
		initialize_args = [config]
		return initialize_result

	func create_data_channel(label: String, options: Dictionary) -> FakeChannel:
		create_data_channel_args.append({"label": label, "options": options})
		if !channels.has(label):
			channels[label] = FakeChannel.new()
		return channels[label]

	func create_offer() -> int:
		create_offer_called += 1
		session_description_created.emit("offer", "sdp-text")
		return OK

	func set_local_description(description_type: String, sdp: String) -> int:
		set_local_description_args = [description_type, sdp]
		return OK

	func set_remote_description(description_type: String, sdp: String) -> int:
		set_remote_description_args = [description_type, sdp]
		return OK

	func add_ice_candidate(media: String, index: int, name: String) -> int:
		add_ice_candidate_args = [media, index, name]
		return OK

	func poll() -> int:
		poll_called += 1
		return OK

	func close() -> void:
		closed = true

func test_start_configures_channels_and_emits_offer() -> void:
	var peer := WebRTCTransportScript.new()
	var fake_peer := FakePeer.new()
	var ready_values: Array = []
	var offer_values: Array = []
	peer.ready.connect(func(channels: Array) -> void:
		ready_values.clear()
		ready_values.append_array(channels)
	)
	peer.offer_created.connect(func(description_type: String, sdp: String) -> void:
		offer_values.clear()
		offer_values.append(description_type)
		offer_values.append(sdp)
	)

	peer.peer_factory = func():
		return fake_peer
	peer.start()
	for channel in fake_peer.channels.values():
		channel.ready_state = WebRTCDataChannel.STATE_OPEN
	peer.poll()

	assert_eq(fake_peer.initialize_args.size(), 1)
	assert_true(fake_peer.initialize_args[0].is_empty())

	assert_eq(fake_peer.create_data_channel_args.size(), 8)
	assert_eq(fake_peer.create_data_channel_args[0]["label"], "sr.world")
	assert_eq(fake_peer.create_data_channel_args[0]["options"]["id"], 1)
	assert_eq(fake_peer.create_data_channel_args[0]["options"]["negotiated"], true)
	assert_eq(fake_peer.create_data_channel_args[0]["options"]["ordered"], true)
	assert_false(fake_peer.create_data_channel_args[0]["options"].has("maxRetransmits"))
	assert_eq(fake_peer.create_data_channel_args[1]["label"], "sr.overlay")
	assert_eq(fake_peer.create_data_channel_args[1]["options"]["id"], 2)
	assert_eq(fake_peer.create_data_channel_args[1]["options"]["negotiated"], true)
	assert_eq(fake_peer.create_data_channel_args[1]["options"]["ordered"], true)
	assert_false(fake_peer.create_data_channel_args[1]["options"].has("maxRetransmits"))
	assert_eq(fake_peer.create_data_channel_args[2]["label"], "sr.session")
	assert_eq(fake_peer.create_data_channel_args[2]["options"]["id"], 3)
	assert_eq(fake_peer.create_data_channel_args[2]["options"]["negotiated"], true)
	assert_eq(fake_peer.create_data_channel_args[2]["options"]["ordered"], true)
	assert_false(fake_peer.create_data_channel_args[2]["options"].has("maxRetransmits"))
	assert_eq(fake_peer.create_data_channel_args[3]["label"], "sr.event")
	assert_eq(fake_peer.create_data_channel_args[3]["options"]["id"], 4)
	assert_eq(fake_peer.create_data_channel_args[3]["options"]["negotiated"], true)
	assert_eq(fake_peer.create_data_channel_args[3]["options"]["ordered"], true)
	assert_false(fake_peer.create_data_channel_args[3]["options"].has("maxRetransmits"))
	assert_eq(fake_peer.create_data_channel_args[4]["label"], "sr.asteroids")
	assert_eq(fake_peer.create_data_channel_args[4]["options"]["id"], 5)
	assert_eq(fake_peer.create_data_channel_args[4]["options"]["negotiated"], true)
	assert_eq(fake_peer.create_data_channel_args[4]["options"]["ordered"], false)
	assert_eq(fake_peer.create_data_channel_args[4]["options"]["maxRetransmits"], 0)
	assert_eq(fake_peer.create_data_channel_args[5]["label"], "sr.bullets")
	assert_eq(fake_peer.create_data_channel_args[5]["options"]["id"], 6)
	assert_eq(fake_peer.create_data_channel_args[5]["options"]["negotiated"], true)
	assert_eq(fake_peer.create_data_channel_args[5]["options"]["ordered"], false)
	assert_eq(fake_peer.create_data_channel_args[5]["options"]["maxRetransmits"], 0)
	assert_eq(fake_peer.create_data_channel_args[6]["label"], "sr.asteroids.lifecycle")
	assert_eq(fake_peer.create_data_channel_args[6]["options"]["id"], 7)
	assert_eq(fake_peer.create_data_channel_args[6]["options"]["negotiated"], true)
	assert_eq(fake_peer.create_data_channel_args[6]["options"]["ordered"], true)
	assert_false(fake_peer.create_data_channel_args[6]["options"].has("maxRetransmits"))
	assert_eq(fake_peer.create_data_channel_args[7]["label"], "sr.bullets.lifecycle")
	assert_eq(fake_peer.create_data_channel_args[7]["options"]["id"], 8)
	assert_eq(fake_peer.create_data_channel_args[7]["options"]["negotiated"], true)
	assert_eq(fake_peer.create_data_channel_args[7]["options"]["ordered"], true)
	assert_false(fake_peer.create_data_channel_args[7]["options"].has("maxRetransmits"))
	assert_eq(fake_peer.create_offer_called, 1)
	assert_eq(offer_values.size(), 2)
	assert_eq(offer_values[0], "offer")
	assert_eq(offer_values[1], "sdp-text")
	assert_eq(ready_values.size(), 8)
	assert_eq(ready_values[0]["lane"], "world")
	assert_eq(ready_values[0]["channel_label"], "sr.world")
	assert_eq(ready_values[0]["channel_id"], 1)
	assert_eq(ready_values[1]["lane"], "overlay")
	assert_eq(ready_values[1]["channel_label"], "sr.overlay")
	assert_eq(ready_values[1]["channel_id"], 2)
	assert_eq(ready_values[2]["lane"], "session")
	assert_eq(ready_values[2]["channel_label"], "sr.session")
	assert_eq(ready_values[2]["channel_id"], 3)
	assert_eq(ready_values[3]["lane"], "event")
	assert_eq(ready_values[3]["channel_label"], "sr.event")
	assert_eq(ready_values[3]["channel_id"], 4)
	assert_eq(ready_values[4]["lane"], "asteroids")
	assert_eq(ready_values[4]["channel_label"], "sr.asteroids")
	assert_eq(ready_values[4]["channel_id"], 5)
	assert_eq(ready_values[5]["lane"], "bullets")
	assert_eq(ready_values[5]["channel_label"], "sr.bullets")
	assert_eq(ready_values[5]["channel_id"], 6)
	assert_eq(ready_values[6]["lane"], "asteroids_lifecycle")
	assert_eq(ready_values[6]["channel_label"], "sr.asteroids.lifecycle")
	assert_eq(ready_values[6]["channel_id"], 7)
	assert_eq(ready_values[7]["lane"], "bullets_lifecycle")
	assert_eq(ready_values[7]["channel_label"], "sr.bullets.lifecycle")
	assert_eq(ready_values[7]["channel_id"], 8)

func test_start_passes_configured_ice_servers_to_initialize() -> void:
	var peer := WebRTCTransportScript.new()
	var fake_peer := FakePeer.new()
	peer.peer_factory = func():
		return fake_peer
	peer.set_ice_servers_for_tests([
		{"urls": ["stun:stun.example.invalid:3478"]},
	])

	peer.start()

	assert_eq(fake_peer.initialize_args.size(), 1)
	assert_eq(fake_peer.initialize_args[0]["iceServers"].size(), 1)
	assert_eq(fake_peer.initialize_args[0]["iceServers"][0]["urls"][0], "stun:stun.example.invalid:3478")

func test_handle_answer_and_remote_ice_forward_to_peer() -> void:
	var peer := WebRTCTransportScript.new()
	var fake_peer := FakePeer.new()
	peer.set_peer_for_tests(fake_peer, {
		"world": FakeChannel.new(),
		"overlay": FakeChannel.new(),
		"session": FakeChannel.new(),
		"event": FakeChannel.new(),
		"asteroids": FakeChannel.new(),
		"bullets": FakeChannel.new(),
		"asteroids_lifecycle": FakeChannel.new(),
		"bullets_lifecycle": FakeChannel.new(),
	})

	peer.handle_answer("answer", "remote-sdp")
	peer.handle_remote_ice("audio", 2, "candidate-name")

	assert_eq(fake_peer.set_remote_description_args[0], "answer")
	assert_eq(fake_peer.set_remote_description_args[1], "remote-sdp")
	assert_eq(fake_peer.add_ice_candidate_args[0], "audio")
	assert_eq(fake_peer.add_ice_candidate_args[1], 2)
	assert_eq(fake_peer.add_ice_candidate_args[2], "candidate-name")

func test_poll_emits_ready_only_after_all_channels_open() -> void:
	var peer := WebRTCTransportScript.new()
	var fake_peer := FakePeer.new()
	var channels := {
		"world": FakeChannel.new(),
		"overlay": FakeChannel.new(),
		"session": FakeChannel.new(),
		"event": FakeChannel.new(),
		"asteroids": FakeChannel.new(),
		"bullets": FakeChannel.new(),
		"asteroids_lifecycle": FakeChannel.new(),
		"bullets_lifecycle": FakeChannel.new(),
	}
	peer.set_peer_for_tests(fake_peer, channels)
	channels["world"].ready_state = WebRTCDataChannel.STATE_OPEN
	channels["overlay"].ready_state = WebRTCDataChannel.STATE_OPEN

	var ready_values: Array = []
	peer.ready.connect(func(payload: Array) -> void:
		ready_values.clear()
		ready_values.append_array(payload)
	)

	peer.poll()
	assert_eq(fake_peer.poll_called, 1)
	assert_true(ready_values.is_empty())

	channels["session"].ready_state = WebRTCDataChannel.STATE_OPEN
	channels["event"].ready_state = WebRTCDataChannel.STATE_OPEN
	channels["asteroids"].ready_state = WebRTCDataChannel.STATE_OPEN
	channels["bullets"].ready_state = WebRTCDataChannel.STATE_OPEN
	peer.poll()

	assert_eq(fake_peer.poll_called, 2)
	assert_eq(ready_values.size(), 8)
	assert_eq(ready_values[0]["lane"], "world")
	assert_eq(ready_values[0]["channel_label"], "sr.world")
	assert_eq(ready_values[0]["channel_id"], 1)
	assert_eq(ready_values[1]["lane"], "overlay")
	assert_eq(ready_values[1]["channel_label"], "sr.overlay")
	assert_eq(ready_values[1]["channel_id"], 2)
	assert_eq(ready_values[2]["lane"], "session")
	assert_eq(ready_values[2]["channel_label"], "sr.session")
	assert_eq(ready_values[2]["channel_id"], 3)
	assert_eq(ready_values[3]["lane"], "event")
	assert_eq(ready_values[3]["channel_label"], "sr.event")
	assert_eq(ready_values[3]["channel_id"], 4)
	assert_eq(ready_values[4]["lane"], "asteroids")
	assert_eq(ready_values[4]["channel_label"], "sr.asteroids")
	assert_eq(ready_values[4]["channel_id"], 5)
	assert_eq(ready_values[5]["lane"], "bullets")
	assert_eq(ready_values[5]["channel_label"], "sr.bullets")
	assert_eq(ready_values[5]["channel_id"], 6)
	assert_eq(ready_values[6]["lane"], "asteroids_lifecycle")
	assert_eq(ready_values[6]["channel_label"], "sr.asteroids.lifecycle")
	assert_eq(ready_values[6]["channel_id"], 7)
	assert_eq(ready_values[7]["lane"], "bullets_lifecycle")
	assert_eq(ready_values[7]["channel_label"], "sr.bullets.lifecycle")
	assert_eq(ready_values[7]["channel_id"], 8)

func test_poll_emits_ready_packet_received_and_smoke_received() -> void:
	var peer := WebRTCTransportScript.new()
	var fake_peer := FakePeer.new()
	var channels := {
		"world": FakeChannel.new(),
		"overlay": FakeChannel.new(),
		"session": FakeChannel.new(),
		"event": FakeChannel.new(),
		"asteroids": FakeChannel.new(),
		"bullets": FakeChannel.new(),
		"asteroids_lifecycle": FakeChannel.new(),
		"bullets_lifecycle": FakeChannel.new(),
	}
	peer.set_peer_for_tests(fake_peer, channels)
	channels["world"].ready_state = WebRTCDataChannel.STATE_OPEN
	channels["overlay"].ready_state = WebRTCDataChannel.STATE_OPEN
	channels["session"].ready_state = WebRTCDataChannel.STATE_OPEN
	channels["event"].ready_state = WebRTCDataChannel.STATE_OPEN
	channels["asteroids"].ready_state = WebRTCDataChannel.STATE_OPEN
	channels["bullets"].ready_state = WebRTCDataChannel.STATE_OPEN
	channels["asteroids_lifecycle"].ready_state = WebRTCDataChannel.STATE_OPEN
	channels["bullets_lifecycle"].ready_state = WebRTCDataChannel.STATE_OPEN
	channels["world"].packets.append(JSON.stringify({
		"type": "webrtc_smoke",
		"smoke_id": "smoke-1",
		"origin": "client",
		"message": "hello",
	}).to_utf8_buffer())
	channels["overlay"].packets.append(JSON.stringify({
		"type": "overlay_delta",
		"lane": "overlay",
	}).to_utf8_buffer())
	channels["session"].packets.append(JSON.stringify({
		"type": "session_delta",
		"lane": "session",
	}).to_utf8_buffer())
	channels["event"].packets.append(JSON.stringify({
		"type": "event_batch",
		"lane": "event",
	}).to_utf8_buffer())
	channels["asteroids"].packets.append(JSON.stringify({
		"type": "asteroid_delta",
		"lane": "asteroids",
	}).to_utf8_buffer())
	channels["bullets"].packets.append(JSON.stringify({
		"type": "bullet_delta",
		"lane": "bullets",
		"bullet_updates": [
			{"id": "bullet-1", "x": 10, "y": 20},
		],
	}).to_utf8_buffer())

	var ready_values: Array = []
	var received_packets: Array = []
	var smoke_packets: Array = []
	var failed_packets: Array = []
	peer.ready.connect(func(payload: Array) -> void:
		ready_values.clear()
		ready_values.append_array(payload)
	)
	peer.packet_received.connect(func(packet: Dictionary) -> void:
		received_packets.append(packet)
	)
	peer.smoke_received.connect(func(packet: Dictionary) -> void:
		smoke_packets.append(packet)
	)
	peer.failed.connect(func(error_code: String, message: String) -> void:
		failed_packets.clear()
		failed_packets.append(error_code)
		failed_packets.append(message)
	)

	peer.poll()

	assert_eq(fake_peer.poll_called, 1)
	assert_eq(ready_values.size(), 8)
	# Assert all expected lanes are present in ready metadata
	var lanes := {}
	for channel in ready_values:
		lanes[str(channel["lane"])] = true
	assert_true(lanes.has("world"))
	assert_true(lanes.has("overlay"))
	assert_true(lanes.has("session"))
	assert_true(lanes.has("event"))
	assert_true(lanes.has("asteroids"))
	assert_true(lanes.has("bullets"))
	assert_true(lanes.has("asteroids_lifecycle"))
	assert_true(lanes.has("bullets_lifecycle"))
	# Positional checks for specific channels
	assert_eq(ready_values[0]["lane"], "world")
	assert_eq(ready_values[0]["channel_label"], "sr.world")
	assert_eq(ready_values[0]["channel_id"], 1)
	assert_eq(received_packets.size(), 5)
	assert_eq(received_packets[0]["type"], "overlay_delta")
	assert_eq(received_packets[0]["lane"], "overlay")
	assert_eq(received_packets[1]["type"], "session_delta")
	assert_eq(received_packets[1]["lane"], "session")
	assert_eq(received_packets[2]["type"], "event_batch")
	assert_eq(received_packets[2]["lane"], "event")
	assert_eq(received_packets[3]["type"], "asteroid_delta")
	assert_eq(received_packets[3]["lane"], "asteroids")
	assert_eq(received_packets[4]["type"], "bullet_delta")
	assert_eq(received_packets[4]["lane"], "bullets")
	assert_eq(smoke_packets.size(), 1)
	assert_true(failed_packets.is_empty())


func test_poll_bounded_receive_drain_per_lane() -> void:
	var peer := WebRTCTransportScript.new()
	var fake_peer := FakePeer.new()
	var channels = {
		"world": FakeChannel.new(),
		"overlay": FakeChannel.new(),
		"session": FakeChannel.new(),
		"event": FakeChannel.new(),
		"asteroids": FakeChannel.new(),
		"bullets": FakeChannel.new(),
		"asteroids_lifecycle": FakeChannel.new(),
		"bullets_lifecycle": FakeChannel.new(),
	}
	peer.set_peer_for_tests(fake_peer, channels)
	for channel in channels.values():
		channel.ready_state = WebRTCDataChannel.STATE_OPEN
	for i in range(WebRTCTransportScript.MAX_PACKETS_PER_LANE_PER_POLL + 3):
		channels["world"].packets.append(JSON.stringify({
			"type": "world_delta",
			"sequence": i + 1,
			"baseline_id": "b%d" % (i + 1),
		}).to_utf8_buffer())

	var received_packets: Array = []
	peer.packet_received.connect(func(packet: Dictionary) -> void:
		received_packets.append(packet)
	)

	peer.poll()

	assert_eq(received_packets.size(), WebRTCTransportScript.MAX_PACKETS_PER_LANE_PER_POLL)
	assert_gt(channels["world"].packets.size(), 0)
	peer.poll()
	assert_gt(received_packets.size(), WebRTCTransportScript.MAX_PACKETS_PER_LANE_PER_POLL)
	assert_eq(fake_peer.poll_called, 2)

func test_poll_tracks_bullet_delta_receive_metrics() -> void:
	var peer := WebRTCTransportScript.new()
	var fake_peer := FakePeer.new()
	var channels = {
		"world": FakeChannel.new(),
		"overlay": FakeChannel.new(),
		"session": FakeChannel.new(),
		"event": FakeChannel.new(),
		"asteroids": FakeChannel.new(),
		"bullets": FakeChannel.new(),
		"asteroids_lifecycle": FakeChannel.new(),
		"bullets_lifecycle": FakeChannel.new(),
	}
	peer.set_peer_for_tests(fake_peer, channels)
	for channel in channels.values():
		channel.ready_state = WebRTCDataChannel.STATE_OPEN

	var server_sent_msec := Time.get_ticks_msec()
	channels["bullets"].packets.append(JSON.stringify({
		"type": "bullet_delta",
		"lane": "bullets",
		"server_sent_msec": server_sent_msec,
	}).to_utf8_buffer())

	peer.poll()

	var metrics = peer.receive_metrics_snapshot()
	assert_eq(metrics["bullet_delta_received_count"], 1)
	assert_eq(metrics["bullet_delta_missing_server_time_count"], 0)
	assert_gte(metrics["bullet_delta_last_age_msec"], 0)
	assert_gte(metrics["bullet_delta_max_age_msec"], 0)

func test_poll_tracks_bullet_delta_missing_server_time() -> void:
	var peer := WebRTCTransportScript.new()
	var fake_peer := FakePeer.new()
	var channels = {
		"world": FakeChannel.new(),
		"overlay": FakeChannel.new(),
		"session": FakeChannel.new(),
		"event": FakeChannel.new(),
		"asteroids": FakeChannel.new(),
		"bullets": FakeChannel.new(),
		"asteroids_lifecycle": FakeChannel.new(),
		"bullets_lifecycle": FakeChannel.new(),
	}
	peer.set_peer_for_tests(fake_peer, channels)
	for channel in channels.values():
		channel.ready_state = WebRTCDataChannel.STATE_OPEN

	channels["bullets"].packets.append(JSON.stringify({
		"type": "bullet_delta",
		"lane": "bullets",
	}).to_utf8_buffer())

	peer.poll()

	var metrics = peer.receive_metrics_snapshot()
	assert_eq(metrics["bullet_delta_missing_server_time_count"], 1)

func test_send_json_writes_text_packet_when_open() -> void:
	var peer := WebRTCTransportScript.new()
	var fake_peer := FakePeer.new()
	var channels := {
		"world": FakeChannel.new(),
	}
	peer.set_peer_for_tests(fake_peer, channels)
	channels["world"].ready_state = WebRTCDataChannel.STATE_OPEN

	peer.send_json({
		"type": "custom_packet",
		"value": "hello",
	})

	var world_channel: FakeChannel = channels["world"]
	assert_eq(world_channel.sent_packets.size(), 1)
	var text: String = world_channel.sent_packets[0].get_string_from_utf8()
	var parsed = JSON.parse_string(text)
	assert_eq(parsed["type"], "custom_packet")
	assert_eq(parsed["value"], "hello")

func test_send_smoke_writes_smoke_packet_when_open() -> void:
	var peer := WebRTCTransportScript.new()
	var fake_peer := FakePeer.new()
	var channels := {
		"world": FakeChannel.new(),
	}
	peer.set_peer_for_tests(fake_peer, channels)
	channels["world"].ready_state = WebRTCDataChannel.STATE_OPEN

	peer.send_smoke("smoke-2", "payload")

	var world_channel: FakeChannel = channels["world"]
	assert_eq(world_channel.sent_packets.size(), 1)
	var text: String = world_channel.sent_packets[0].get_string_from_utf8()
	var parsed = JSON.parse_string(text)
	assert_eq(parsed["type"], "webrtc_smoke")
	assert_eq(parsed["smoke_id"], "smoke-2")
	assert_eq(parsed["origin"], "client")
	assert_eq(parsed["message"], "payload")