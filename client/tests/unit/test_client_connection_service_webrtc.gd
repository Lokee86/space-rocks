extends GutTest

const ClientConnectionService := preload("res://scripts/networking/client_connection_service.gd")
const WebRTCSmokePeerScript := preload("res://scripts/networking/webrtc/webrtc_smoke_peer.gd")


class FakeSmokePeer:
	extends WebRTCSmokePeer

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


func test_connection_service_starts_and_wires_webrtc_smoke_peer() -> void:
	var service := ClientConnectionService.new()
	var fake_network := FakeNetworkClient.new()
	var fake_peer := FakeSmokePeer.new()
	var fake_sender := ClientConnectionService.ClientPacketSender.new(fake_network)
	service.network_client = fake_network
	service.client_packet_sender = fake_sender
	service.server_packet_dispatcher = null
	service.realtime_router = null
	service.webrtc_smoke_peer_factory = Callable(self, "_make_fake_smoke_peer").bind(fake_peer)
	add_child_autofree(service)

	service._on_connected()
	service.webrtc_smoke_peer.offer_created.emit("answer-type", "answer-sdp")
	service.webrtc_smoke_peer.ice_candidate_created.emit("audio", 2, "candidate-name")
	service.webrtc_smoke_peer.ready.emit("sr.reliable", 1)
	service.webrtc_smoke_peer.smoke_received.emit({"type": "webrtc_smoke", "smoke_id": "server-smoke", "origin": "server"})
	service.webrtc_smoke_peer.failed.emit("peer_error", "boom")
	service._process(0.0)
	service._on_closed()

	assert_eq(fake_peer.started, 1)
	assert_true(fake_network.sent_packets.has({"type": "webrtc_offer", "description_type": "answer-type", "sdp": "answer-sdp"}))
	assert_true(fake_network.sent_packets.has({"type": "webrtc_ice_candidate", "media": "audio", "index": 2, "name": "candidate-name"}))
	assert_eq(fake_peer.sent_smokes, [["client-smoke-1", "client smoke peer ready"]])
	assert_true(fake_network.sent_packets.has({"type": "webrtc_failed", "error_code": "peer_error", "message": "boom"}))
	assert_true(fake_peer.polled >= 1)
	assert_true(fake_peer.closed >= 1)


func _make_fake_smoke_peer(fake_peer: FakeSmokePeer) -> WebRTCSmokePeer:
	return fake_peer

