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
	var channel := FakeChannel.new()
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
		create_data_channel_args = [label, options]
		return channel

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


func test_start_configures_channel_and_emits_offer() -> void:
	var peer := WebRTCTransportScript.new()
	var fake_peer := FakePeer.new()
	var ready_values: Array = []
	var offer_values: Array = []
	peer.ready.connect(func(channel_label: String, channel_id: int) -> void:
		ready_values.clear()
		ready_values.append(channel_label)
		ready_values.append(channel_id)
	)
	peer.offer_created.connect(func(description_type: String, sdp: String) -> void:
		offer_values.clear()
		offer_values.append(description_type)
		offer_values.append(sdp)
	)

	peer.peer_factory = func():
		return fake_peer
	peer.start()
	fake_peer.channel.ready_state = WebRTCDataChannel.STATE_OPEN
	peer.poll()

	assert_eq(fake_peer.initialize_args.size(), 1)
	assert_true(fake_peer.initialize_args[0].is_empty())

	assert_eq(fake_peer.create_data_channel_args.size(), 2)
	assert_eq(fake_peer.create_data_channel_args[0], WebRTCTransportScript.CHANNEL_LABEL)
	assert_eq(fake_peer.create_data_channel_args[1]["id"], WebRTCTransportScript.CHANNEL_ID)
	assert_eq(fake_peer.create_data_channel_args[1]["negotiated"], true)
	assert_eq(fake_peer.create_data_channel_args[1]["ordered"], true)
	assert_eq(fake_peer.create_offer_called, 1)
	assert_eq(offer_values.size(), 2)
	assert_eq(offer_values[0], "offer")
	assert_eq(offer_values[1], "sdp-text")
	assert_eq(ready_values.size(), 2)
	assert_eq(ready_values[0], WebRTCTransportScript.CHANNEL_LABEL)
	assert_eq(ready_values[1], WebRTCTransportScript.CHANNEL_ID)


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
	peer.set_peer_for_tests(fake_peer, fake_peer.channel)

	peer.handle_answer("answer", "remote-sdp")
	peer.handle_remote_ice("audio", 2, "candidate-name")

	assert_eq(fake_peer.set_remote_description_args[0], "answer")
	assert_eq(fake_peer.set_remote_description_args[1], "remote-sdp")
	assert_eq(fake_peer.add_ice_candidate_args[0], "audio")
	assert_eq(fake_peer.add_ice_candidate_args[1], 2)
	assert_eq(fake_peer.add_ice_candidate_args[2], "candidate-name")


func test_poll_emits_ready_packet_received_and_smoke_received() -> void:
	var peer := WebRTCTransportScript.new()
	var fake_peer := FakePeer.new()
	peer.set_peer_for_tests(fake_peer, fake_peer.channel)
	fake_peer.channel.ready_state = WebRTCDataChannel.STATE_OPEN
	fake_peer.channel.packets.append(JSON.stringify({
		"type": "webrtc_smoke",
		"smoke_id": "smoke-1",
		"origin": "client",
		"message": "hello",
	}).to_utf8_buffer())

	var ready_values: Array = []
	var received_packets: Array = []
	var smoke_packets: Array = []
	var failed_packets: Array = []
	peer.ready.connect(func(channel_label: String, channel_id: int) -> void:
		ready_values.clear()
		ready_values.append(channel_label)
		ready_values.append(channel_id)
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
	assert_eq(ready_values.size(), 2)
	assert_eq(ready_values[0], WebRTCTransportScript.CHANNEL_LABEL)
	assert_eq(ready_values[1], WebRTCTransportScript.CHANNEL_ID)
	assert_eq(received_packets.size(), 1)
	assert_eq(received_packets[0]["type"], "webrtc_smoke")
	assert_eq(received_packets[0]["smoke_id"], "smoke-1")
	assert_eq(smoke_packets.size(), 1)
	assert_true(failed_packets.is_empty())


func test_send_json_writes_text_packet_when_open() -> void:
	var peer := WebRTCTransportScript.new()
	var fake_peer := FakePeer.new()
	peer.set_peer_for_tests(fake_peer, fake_peer.channel)
	fake_peer.channel.ready_state = WebRTCDataChannel.STATE_OPEN

	peer.send_json({
		"type": "custom_packet",
		"value": "hello",
	})

	assert_eq(fake_peer.channel.sent_packets.size(), 1)
	var text := fake_peer.channel.sent_packets[0].get_string_from_utf8()
	var parsed = JSON.parse_string(text)
	assert_eq(parsed["type"], "custom_packet")
	assert_eq(parsed["value"], "hello")


func test_send_smoke_writes_smoke_packet_when_open() -> void:
	var peer := WebRTCTransportScript.new()
	var fake_peer := FakePeer.new()
	peer.set_peer_for_tests(fake_peer, fake_peer.channel)
	fake_peer.channel.ready_state = WebRTCDataChannel.STATE_OPEN

	peer.send_smoke("smoke-2", "payload")

	assert_eq(fake_peer.channel.sent_packets.size(), 1)
	var text := fake_peer.channel.sent_packets[0].get_string_from_utf8()
	var parsed = JSON.parse_string(text)
	assert_eq(parsed["type"], "webrtc_smoke")
	assert_eq(parsed["smoke_id"], "smoke-2")
	assert_eq(parsed["origin"], WebRTCTransportScript.SMOKE_ORIGIN_CLIENT)
	assert_eq(parsed["message"], "payload")


func test_poll_emits_failed_for_invalid_json() -> void:
	var peer := WebRTCTransportScript.new()
	var fake_peer := FakePeer.new()
	peer.set_peer_for_tests(fake_peer, fake_peer.channel)
	fake_peer.channel.ready_state = WebRTCDataChannel.STATE_OPEN
	fake_peer.channel.packets.append("{invalid json".to_utf8_buffer())

	var failed_packets: Array = []
	peer.failed.connect(func(error_code: String, message: String) -> void:
		failed_packets.clear()
		failed_packets.append(error_code)
		failed_packets.append(message)
	)

	peer.poll()

	assert_eq(failed_packets.size(), 2)
	assert_eq(failed_packets[0], "invalid_json")
	assert_true(str(failed_packets[1]).contains("Invalid packet JSON"))


func test_poll_emits_packet_received_for_compact_realtime_packet() -> void:
	var peer := WebRTCTransportScript.new()
	var fake_peer := FakePeer.new()
	peer.set_peer_for_tests(fake_peer, fake_peer.channel)
	fake_peer.channel.ready_state = WebRTCDataChannel.STATE_OPEN
	fake_peer.channel.packets.append('{"t":"wd","q":1,"bq":0,"ms":123}'.to_utf8_buffer())

	var received_packets: Array = []
	peer.packet_received.connect(func(packet: Dictionary) -> void:
		received_packets.append(packet)
	)

	peer.poll()

	assert_eq(received_packets.size(), 1)
	assert_eq(received_packets[0]["type"], "world_delta")

