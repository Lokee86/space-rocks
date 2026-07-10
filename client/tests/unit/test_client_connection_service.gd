extends GutTest

const ClientConnectionService := preload("res://scripts/networking/client_connection_service.gd")
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

	assert_true(service.get_realtime_packet_pipeline() == service.realtime_packet_pipeline)
	assert_false(service.get_realtime_packet_pipeline().is_gameplay_ready())

	service.gameplay_packet_received.connect(func(_packet: Dictionary) -> void:
		callback_state.gameplay_packet_count += 1
	)
	service.realtime_packet_pipeline.gameplay_packet_applied.connect(func(_packet: Dictionary) -> void:
		callback_state.pipeline_packet_count += 1
		assert_false(service.get_realtime_packet_pipeline().is_gameplay_ready())
		assert_true(service.get_realtime_packet_pipeline().get_presentation_state().world_lane_state != null)
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
	assert_true(service.get_realtime_packet_pipeline().is_gameplay_ready())


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
	assert_false(service.get_realtime_packet_pipeline().is_gameplay_ready())


func test_reset_exposes_fresh_pipeline_and_readiness() -> void:
	var service := ClientConnectionService.new()
	add_child_autofree(service)

	var pipeline := service.get_realtime_packet_pipeline()
	var presentation_state := pipeline.get_presentation_state()
	var world_lane_state := presentation_state.world_lane_state
	var overlay_lane_state := presentation_state.overlay_lane_state
	var session_lane_state := presentation_state.session_lane_state
	var event_batch_applier := presentation_state.event_batch_applier
	assert_true(pipeline == service.get_realtime_packet_pipeline())
	assert_false(pipeline.is_gameplay_ready())

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
	assert_true(pipeline.is_gameplay_ready())

	service.reset_realtime_protocol_state()

	assert_true(service.get_realtime_packet_pipeline() == pipeline)
	assert_false(pipeline.is_gameplay_ready())
	assert_true(service.get_realtime_packet_pipeline().get_presentation_state() == presentation_state)
	assert_true(presentation_state.world_lane_state != world_lane_state)
	assert_true(presentation_state.overlay_lane_state != overlay_lane_state)
	assert_true(presentation_state.session_lane_state != session_lane_state)
	assert_true(presentation_state.event_batch_applier != event_batch_applier)
	assert_true(presentation_state.world_lane_state != null)
	assert_true(presentation_state.overlay_lane_state != null)
	assert_true(presentation_state.session_lane_state != null)
	assert_true(presentation_state.event_batch_applier != null)
