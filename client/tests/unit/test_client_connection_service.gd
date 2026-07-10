extends GutTest

const ClientConnectionService := preload("res://scripts/networking/client_connection_service.gd")
const LaneMetadata := preload("res://scripts/protocol/realtime/lane_metadata.gd")
const RealtimeRouter := preload("res://scripts/protocol/realtime/realtime_router.gd")
const WebRTCTransport := preload("res://scripts/networking/webrtc/webrtc_transport.gd")


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


class FakeTransportPeer:
	extends WebRTCTransport

	func start() -> void:
		pass

	func poll() -> void:
		pass

	func close() -> void:
		pass


func _make_fake_transport_peer(fake_peer: FakeTransportPeer) -> WebRTCTransport:
	return fake_peer


func test_inbound_valid_gameplay_packet_routes_through_pipeline_once() -> void:
	var service := ClientConnectionService.new()
	var callback_state := {"gameplay_packet_count": 0, "pipeline_packet_count": 0, "state_seen": false}
	add_child_autofree(service)

	service.gameplay_packet_received.connect(func(packet: Dictionary) -> void:
		callback_state.gameplay_packet_count += 1
	)
	service.realtime_packet_pipeline.gameplay_packet_applied.connect(func(packet: Dictionary) -> void:
		callback_state.pipeline_packet_count += 1
		assert_true(service.get_realtime_router().baseline_tracker.is_lane_synced(LaneMetadata.LANE_WORLD))
		callback_state.state_seen = true
	)

	service._on_connected()
	assert_true(service.realtime_transport_session != null)
	assert_true(service.realtime_transport_session.transport != null)

	service._on_packet_received({
		"type": "world_full",
		"baseline_id": "world-baseline-1",
		"sequence": 1,
		"snapshot_id": "world-snapshot-1",
		"is_final_chunk": true,
		"ships": [],
		"bullets": [],
		"asteroids": [],
		"pickups": [],
	})

	assert_true(callback_state.state_seen)
	assert_eq(callback_state.pipeline_packet_count, 1)
	assert_eq(callback_state.gameplay_packet_count, 1)
	assert_true(service.get_realtime_router().baseline_tracker.is_lane_synced(LaneMetadata.LANE_WORLD))


func test_websocket_and_webrtc_gameplay_packets_share_pipeline_application_path() -> void:
	var service := ClientConnectionService.new()
	var callback_state := {"gameplay_packet_count": 0, "pipeline_packet_count": 0}
	var fake_network := FakeNetworkClient.new()
	service.network_client = fake_network
	service.client_packet_sender = ClientConnectionService.ClientPacketSender.new(fake_network)
	service.server_packet_dispatcher = ClientConnectionService.ServerPacketDispatcher.new()
	service.webrtc_transport_factory = Callable(self, "_make_fake_transport_peer").bind(FakeTransportPeer.new())
	add_child_autofree(service)
	service._on_connected()
	assert_true(service.realtime_transport_session != null)
	assert_true(service.realtime_transport_session.transport != null)

	service.gameplay_packet_received.connect(func(_packet: Dictionary) -> void:
		callback_state.gameplay_packet_count += 1
	)
	service.realtime_packet_pipeline.gameplay_packet_applied.connect(func(_packet: Dictionary) -> void:
		callback_state.pipeline_packet_count += 1
	)

	service._on_packet_received({
		"type": "world_full",
		"baseline_id": "world-baseline-1",
		"sequence": 1,
		"snapshot_id": "world-snapshot-1",
		"is_final_chunk": true,
		"ships": [],
		"bullets": [],
		"asteroids": [],
		"pickups": [],
	})
	service.realtime_transport_session.transport.packet_received.emit({
		"type": "world_delta",
		"baseline_id": "world-baseline-1",
		"sequence": 2,
	})

	assert_eq(callback_state.pipeline_packet_count, 2)
	assert_eq(callback_state.gameplay_packet_count, 2)
	assert_true(service.get_realtime_router().baseline_tracker.is_lane_synced(LaneMetadata.LANE_WORLD))


func test_reset_exposes_fresh_router_and_readiness() -> void:
	var service := ClientConnectionService.new()
	add_child_autofree(service)

	var old_router: RealtimeRouter = service.get_realtime_router()
	service._on_packet_received({
		"type": "world_full",
		"baseline_id": "world-baseline-1",
		"sequence": 1,
		"snapshot_id": "world-snapshot-1",
		"is_final_chunk": true,
		"ships": [],
		"bullets": [],
		"asteroids": [],
		"pickups": [],
	})
	assert_true(old_router.baseline_tracker.is_lane_synced(LaneMetadata.LANE_WORLD))
	assert_eq(service.get_gameplay_readiness(), old_router.get_gameplay_readiness())

	service.reset_realtime_protocol_state()

	var new_router: RealtimeRouter = service.get_realtime_router()
	assert_ne(new_router, old_router)
	assert_false(new_router.baseline_tracker.is_lane_synced(LaneMetadata.LANE_WORLD))
	assert_eq(service.get_gameplay_readiness(), new_router.get_gameplay_readiness())
	assert_false(service.get_gameplay_readiness() == old_router.get_gameplay_readiness())
