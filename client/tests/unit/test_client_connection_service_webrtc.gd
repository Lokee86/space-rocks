extends GutTest

const ClientConnectionService := preload("res://scripts/networking/client_connection_service.gd")
const WebRTCTransportScript := preload("res://scripts/networking/webrtc/webrtc_transport.gd")


class FakeTransportPeer:
	extends WebRTCTransport

	var started := 0
	var polled := 0
	var closed := 0
	var answered: Array = []
	var remote_ice: Array = []
	var sent_smokes: Array = []

	func start() -> void:
		started += 1

	func poll() -> void:
		polled += 1

	func handle_answer(description_type: String, sdp: String) -> void:
		answered = [description_type, sdp]

	func handle_remote_ice(media: String, index: int, name: String) -> void:
		remote_ice = [media, index, name]

	func send_smoke(smoke_id: String, message: String) -> void:
		sent_smokes.append([smoke_id, message])

	func close() -> void:
		closed += 1


class FakeNetworkClient:
	extends NetworkClient

	var sent_packets: Array = []

	func send_raw_packet(packet: Dictionary) -> void:
		sent_packets.append(packet)

	func is_connected_to_server() -> bool:
		return true

	func connect_to_server(_url: String) -> Error:
		return OK

	func begin_graceful_close() -> bool:
		return true

	func close_gracefully() -> void:
		pass


func test_connection_service_starts_and_wires_webrtc_transport() -> void:
	var service := ClientConnectionService.new()
	var fake_network := FakeNetworkClient.new()
	var fake_peer := FakeTransportPeer.new()
	var fake_sender := ClientConnectionService.ClientPacketSender.new(fake_network)
	service.network_client = fake_network
	service.client_packet_sender = fake_sender
	service.server_packet_dispatcher = null
	service.webrtc_transport_factory = Callable(self, "_make_fake_transport_peer").bind(fake_peer)
	add_child_autofree(service)

	service._on_connected()
	service.realtime_transport_session.transport.offer_created.emit("answer-type", "answer-sdp")
	service.realtime_transport_session.transport.ice_candidate_created.emit("audio", 2, "candidate-name")
	service.realtime_transport_session.transport.ready.emit([
		{"lane": "world", "channel_label": "sr.world", "channel_id": 1},
		{"lane": "overlay", "channel_label": "sr.overlay", "channel_id": 2},
		{"lane": "session", "channel_label": "sr.session", "channel_id": 3},
		{"lane": "event", "channel_label": "sr.event", "channel_id": 4},
		{"lane": "asteroids", "channel_label": "sr.asteroids", "channel_id": 5},
		{"lane": "bullets", "channel_label": "sr.bullets", "channel_id": 6},
	])
	service.realtime_transport_session.transport.packet_received.emit({"type": "webrtc_smoke", "smoke_id": "server-smoke", "origin": "server"})
	service.realtime_transport_session.transport.failed.emit("peer_error", "boom")
	service._process(0.0)
	service._on_closed()

	assert_eq(fake_peer.started, 1)
	assert_true(fake_network.sent_packets.has({"type": "webrtc_offer", "description_type": "answer-type", "sdp": "answer-sdp"}))
	assert_true(fake_network.sent_packets.has({"type": "webrtc_ice_candidate", "media": "audio", "index": 2, "name": "candidate-name"}))
	assert_eq(fake_peer.sent_smokes, [["client-smoke-1", "client smoke peer ready"]])
	assert_true(fake_network.sent_packets.has({"type": "webrtc_failed", "error_code": "peer_error", "message": "boom"}))
	assert_true(fake_peer.polled >= 1)
	assert_true(fake_peer.closed >= 1)


func test_connection_service_does_not_poll_closed_webrtc_transport_after_reset() -> void:
	var service := ClientConnectionService.new()
	var fake_network := FakeNetworkClient.new()
	var fake_peer := FakeTransportPeer.new()
	var fake_sender := ClientConnectionService.ClientPacketSender.new(fake_network)
	service.network_client = fake_network
	service.client_packet_sender = fake_sender
	service.server_packet_dispatcher = null
	service.webrtc_transport_factory = Callable(self, "_make_fake_transport_peer").bind(fake_peer)
	add_child_autofree(service)

	service._on_connected()
	var old_transport: WebRTCTransport = service.realtime_transport_session.transport
	var initial_poll_count := fake_peer.polled
	service.reset_realtime_protocol_state()
	service._process(0.0)

	assert_true(old_transport != null)
	assert_eq(fake_peer.polled, initial_poll_count)


func test_webrtc_transport_replacement_packets_reach_dispatcher_and_gameplay() -> void:
	var service := ClientConnectionService.new()
	var fake_network := FakeNetworkClient.new()
	var first_peer := FakeTransportPeer.new()
	var second_peer := FakeTransportPeer.new()
	var fake_sender := ClientConnectionService.ClientPacketSender.new(fake_network)
	var dispatcher_packets: Array = []
	var gameplay_packets: Array = []
	var smoke_packets: Array = []
	var transport_peers := [first_peer, second_peer]
	service.network_client = fake_network
	service.client_packet_sender = fake_sender
	service.server_packet_dispatcher = ClientConnectionService.ServerPacketDispatcher.new()
	service.webrtc_transport_factory = Callable(self, "_make_fake_transport_peer_from_queue").bind(transport_peers)
	add_child_autofree(service)

	service.server_packet_dispatcher.resync_request_received.connect(func(packet: Dictionary) -> void:
		dispatcher_packets.append(packet)
	)
	service.gameplay_packet_received.connect(func(packet: Dictionary) -> void:
		gameplay_packets.append(packet)
	)
	service.webrtc_smoke_received.connect(func(packet: Dictionary) -> void:
		smoke_packets.append(packet)
	)

	service._on_connected()
	service._on_closed()
	service._on_connected()
	service.realtime_transport_session.transport.packet_received.emit({"type": "resync_request", "lane": "world"})
	service.realtime_transport_session.transport.packet_received.emit({"type": "webrtc_smoke", "smoke_id": "smoke-1", "origin": "server"})

	assert_eq(first_peer.started, 1)
	assert_eq(first_peer.closed, 1)
	assert_eq(second_peer.started, 1)
	assert_eq(dispatcher_packets.size(), 1)
	assert_eq(dispatcher_packets[0], {"type": "resync_request", "lane": "world"})
	assert_eq(gameplay_packets.size(), 1)
	assert_eq(gameplay_packets[0], {"type": "resync_request", "lane": "world"})
	assert_true(smoke_packets.is_empty())


func test_webrtc_transport_asteroid_delta_routes_into_realtime_router() -> void:
	var service := ClientConnectionService.new()
	var fake_network := FakeNetworkClient.new()
	var fake_peer := FakeTransportPeer.new()
	var fake_sender := ClientConnectionService.ClientPacketSender.new(fake_network)
	service.network_client = fake_network
	service.client_packet_sender = fake_sender
	service.server_packet_dispatcher = ClientConnectionService.ServerPacketDispatcher.new()
	service.webrtc_transport_factory = Callable(self, "_make_fake_transport_peer").bind(fake_peer)
	add_child_autofree(service)

	var pipeline = service.get_realtime_packet_pipeline()
	pipeline.get_presentation_state().world_lane_state.upsert_asteroid({"id": "asteroid-1", "x": 1.0, "y": 2.0, "rotation": 0.0})

	service._on_packet_received({
		"type": "asteroid_delta",
		"sequence": 1,
		"asteroid_updates": [
			{"id": "asteroid-1", "x": 42, "y": 84},
		],
	})

	assert_eq(pipeline.get_presentation_state().world_lane_state.asteroids["asteroid-1"]["x"], 4.2)
	assert_eq(pipeline.get_presentation_state().world_lane_state.asteroids["asteroid-1"]["y"], 8.4)


func test_webrtc_transport_bullet_delta_routes_into_realtime_router() -> void:
	var service := ClientConnectionService.new()
	var fake_network := FakeNetworkClient.new()
	var fake_peer := FakeTransportPeer.new()
	var fake_sender := ClientConnectionService.ClientPacketSender.new(fake_network)
	service.network_client = fake_network
	service.client_packet_sender = fake_sender
	service.server_packet_dispatcher = ClientConnectionService.ServerPacketDispatcher.new()
	service.webrtc_transport_factory = Callable(self, "_make_fake_transport_peer").bind(fake_peer)
	add_child_autofree(service)

	var pipeline = service.get_realtime_packet_pipeline()
	pipeline.get_presentation_state().world_lane_state.upsert_bullet({"id": "bullet-1", "x": 1.0, "y": 2.0, "rotation": 0.0})

	service._on_packet_received({
		"type": "bullet_delta",
		"sequence": 1,
		"bullet_updates": [
			{"id": "bullet-1", "x": 55, "y": 66},
		],
	})

	assert_eq(pipeline.get_presentation_state().world_lane_state.bullets["bullet-1"]["x"], 5.5)
	assert_eq(pipeline.get_presentation_state().world_lane_state.bullets["bullet-1"]["y"], 6.6)


func test_webrtc_transport_reconnect_ownership_closes_previous_transport_and_starts_new_one() -> void:
	var service := ClientConnectionService.new()
	var fake_network := FakeNetworkClient.new()
	var first_peer := FakeTransportPeer.new()
	var second_peer := FakeTransportPeer.new()
	var fake_sender := ClientConnectionService.ClientPacketSender.new(fake_network)
	service.network_client = fake_network
	service.client_packet_sender = fake_sender
	service.server_packet_dispatcher = ClientConnectionService.ServerPacketDispatcher.new()
	service.webrtc_transport_factory = Callable(self, "_make_fake_transport_peer").bind(first_peer)


	service._on_connected()
	service._on_closed()
	service.webrtc_transport_factory = Callable(self, "_make_fake_transport_peer").bind(second_peer)
	service._on_connected()

	assert_eq(first_peer.started, 1)
	assert_eq(first_peer.closed, 1)
	assert_eq(second_peer.started, 1)
	assert_true(service.realtime_transport_session.transport == second_peer)


func _make_fake_transport_peer(fake_peer: FakeTransportPeer) -> WebRTCTransport:
	return fake_peer


func _make_fake_transport_peer_from_queue(transport_peers: Array) -> WebRTCTransport:
	return transport_peers.pop_front()
