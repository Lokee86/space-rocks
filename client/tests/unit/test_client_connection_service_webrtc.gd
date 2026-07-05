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
	service.realtime_router = null
	service.webrtc_transport_factory = Callable(self, "_make_fake_transport_peer").bind(fake_peer)
	add_child_autofree(service)

	service._on_connected()
	service.webrtc_transport.offer_created.emit("answer-type", "answer-sdp")
	service.webrtc_transport.ice_candidate_created.emit("audio", 2, "candidate-name")
	service.webrtc_transport.ready.emit([
		{"lane": "world", "channel_label": "sr.world", "channel_id": 1},
		{"lane": "overlay", "channel_label": "sr.overlay", "channel_id": 2},
		{"lane": "session", "channel_label": "sr.session", "channel_id": 3},
		{"lane": "event", "channel_label": "sr.event", "channel_id": 4},
	])
	service.webrtc_transport.packet_received.emit({"type": "webrtc_smoke", "smoke_id": "server-smoke", "origin": "server"})
	service.webrtc_transport.failed.emit("peer_error", "boom")
	service._process(0.0)
	service._on_closed()

	assert_eq(fake_peer.started, 1)
	assert_true(fake_network.sent_packets.has({"type": "webrtc_offer", "description_type": "answer-type", "sdp": "answer-sdp"}))
	assert_true(fake_network.sent_packets.has({"type": "webrtc_ice_candidate", "media": "audio", "index": 2, "name": "candidate-name"}))
	assert_eq(fake_peer.sent_smokes, [["client-smoke-1", "client smoke peer ready"]])
	assert_true(fake_network.sent_packets.has({"type": "webrtc_failed", "error_code": "peer_error", "message": "boom"}))
	assert_true(fake_peer.polled >= 1)
	assert_true(fake_peer.closed >= 1)


func test_webrtc_transport_non_smoke_packet_dispatches_to_gameplay() -> void:
	var service := ClientConnectionService.new()
	var fake_network := FakeNetworkClient.new()
	var fake_peer := FakeTransportPeer.new()
	var fake_sender := ClientConnectionService.ClientPacketSender.new(fake_network)
	var gameplay_packets: Array = []
	var smoke_packets: Array = []
	service.network_client = fake_network
	service.client_packet_sender = fake_sender
	service.server_packet_dispatcher = ClientConnectionService.ServerPacketDispatcher.new()
	service.realtime_router = null
	service.webrtc_transport_factory = Callable(self, "_make_fake_transport_peer").bind(fake_peer)
	add_child_autofree(service)
	service.gameplay_packet_received.connect(func(packet: Dictionary) -> void:
		gameplay_packets.append(packet)
	)
	service.webrtc_smoke_received.connect(func(packet: Dictionary) -> void:
		smoke_packets.append(packet)
	)
	service._on_connected()
	service.webrtc_transport.packet_received.emit({"type": "resync_request", "lane": "world"})
	service.webrtc_transport.packet_received.emit({"type": "webrtc_smoke", "smoke_id": "smoke-1", "origin": "server"})

	assert_eq(gameplay_packets.size(), 1)
	assert_eq(gameplay_packets[0]["type"], "resync_request")
	assert_eq(gameplay_packets[0]["lane"], "world")
	assert_true(smoke_packets.is_empty())


func _make_fake_transport_peer(fake_peer: FakeTransportPeer) -> WebRTCTransport:
	return fake_peer

