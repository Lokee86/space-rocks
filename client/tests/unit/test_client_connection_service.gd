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

	var close_calls := 0

	func start() -> void:
		pass

	func poll() -> void:
		pass

	func close() -> void:
		close_calls += 1


func _make_fake_transport_peer(fake_peer: FakeTransportPeer) -> WebRTCTransport:
	return fake_peer


func test_protocol_match_begin_end_preserves_transport_without_close() -> void:
	var service := ClientConnectionService.new()
	add_child_autofree(service)
	var peer := FakeTransportPeer.new()
	var retained := ClientConnectionService.RealtimeTransportSession.new()
	retained.transport = peer
	service.realtime_transport_session = retained
	service.begin_realtime_match("match-1")
	service.end_realtime_match()
	assert_true(service.realtime_transport_session == retained)
	assert_true(service.realtime_transport_session.transport == peer)
	assert_eq(peer.close_calls, 0)


func test_resync_required_uses_active_match_and_suppresses_when_inactive() -> void:
	var service := ClientConnectionService.new()
	add_child_autofree(service)
	var fake_network := FakeNetworkClient.new()
	add_child_autofree(fake_network)
	service.network_client = fake_network
	service.client_packet_sender = ClientConnectionService.ClientPacketSender.new(fake_network)
	service.begin_realtime_match("match-1")
	service._on_resync_request_required("world", "baseline-1", 1, "missing_baseline")
	assert_eq(fake_network.sent_packets.size(), 1)
	assert_eq(fake_network.sent_packets[0].get("match_id"), "match-1")
	service.end_realtime_match()
	service._on_resync_request_required("world", "baseline-1", 1, "missing_baseline")
	assert_eq(fake_network.sent_packets.size(), 1)


func test_connection_service_does_not_expose_raw_webrtc_inbound_signals() -> void:
	var service := ClientConnectionService.new()
	add_child_autofree(service)

	assert_false(service.has_signal("webrtc_answer_received"))
	assert_false(service.has_signal("webrtc_ice_candidate_received"))
	assert_false(service.has_signal("webrtc_ready_received"))
	assert_false(service.has_signal("webrtc_smoke_received"))
	assert_false(service.has_signal("webrtc_failed_received"))
	assert_true(service.has_signal("realtime_transport_ready"))


func test_inbound_valid_gameplay_packet_routes_through_pipeline_once() -> void:
	var service := ClientConnectionService.new()
	var callback_state := {"pipeline_packet_count": 0, "state_seen": false}
	add_child_autofree(service)

	assert_true(service.get_realtime_packet_pipeline() == service.realtime_packet_pipeline)
	assert_false(service.get_realtime_packet_pipeline().is_gameplay_ready())

	service.realtime_packet_pipeline.gameplay_packet_applied.connect(func(_packet: Dictionary) -> void:
		callback_state.pipeline_packet_count += 1
		assert_false(service.get_realtime_packet_pipeline().is_gameplay_ready())
		assert_true(service.get_realtime_packet_pipeline().get_presentation_state().world_lane_state != null)
		callback_state.state_seen = true
	)

	service._on_connected()
	assert_true(service.realtime_transport_session != null)
	assert_true(service.realtime_transport_session.transport != null)
	service.begin_realtime_match("match-1")

	service.server_packet_dispatcher.dispatch({
		"type": "world_full",
		"match_id": "match-1",
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



func test_websocket_and_webrtc_gameplay_packets_share_pipeline_application_path() -> void:
	var service := ClientConnectionService.new()
	var callback_state := {"pipeline_packet_count": 0}
	var fake_network := FakeNetworkClient.new()
	service.network_client = fake_network
	service.client_packet_sender = ClientConnectionService.ClientPacketSender.new(fake_network)
	service.server_packet_dispatcher = ClientConnectionService.ServerPacketDispatcher.new()
	service.webrtc_transport_factory = Callable(self, "_make_fake_transport_peer").bind(FakeTransportPeer.new())
	add_child_autofree(service)
	service._on_connected()
	assert_true(service.realtime_transport_session != null)
	assert_true(service.realtime_transport_session.transport != null)
	service.begin_realtime_match("match-1")

	
	service.realtime_packet_pipeline.gameplay_packet_applied.connect(func(_packet: Dictionary) -> void:
		callback_state.pipeline_packet_count += 1
	)

	service._on_packet_received({
		"type": "world_full",
		"match_id": "match-1",
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
		"match_id": "match-1",
		"baseline_id": "world-baseline-1",
		"sequence": 2,
	})

	assert_eq(callback_state.pipeline_packet_count, 2)
	
	assert_false(service.get_realtime_packet_pipeline().is_gameplay_ready())


func test_clock_offset_is_forwarded_and_reset() -> void:
	var service := ClientConnectionService.new()
	var peer := FakeTransportPeer.new()
	service.webrtc_transport_factory = Callable(self, "_make_fake_transport_peer").bind(peer)
	service.set_server_clock_offset_ms(125)
	add_child_autofree(service)
	service._ensure_realtime_transport_session()
	service.realtime_transport_session.start()
	assert_eq(service.realtime_transport_session.transport.server_clock_offset_ms, 125)
	service.set_server_clock_offset_ms(250)
	assert_eq(service.realtime_transport_session.transport.server_clock_offset_ms, 250)
	service.reset_realtime_session()
	assert_eq(service.server_clock_offset_ms, -1)


func test_reset_exposes_fresh_pipeline_and_readiness() -> void:
	var service := ClientConnectionService.new()
	add_child_autofree(service)

	var pipeline := service.get_realtime_packet_pipeline()
	var presentation_state := pipeline.get_presentation_state()
	var world_lane_state: Variant = presentation_state.world_lane_state
	var overlay_lane_state: Variant = presentation_state.overlay_lane_state
	var session_lane_state: Variant = presentation_state.session_lane_state
	var event_batch_applier: Variant = presentation_state.event_batch_applier
	var applied_packets: Array = []

	pipeline.gameplay_packet_applied.connect(func(packet: Dictionary) -> void:
		applied_packets.append(packet)
	)

	assert_false(pipeline.is_gameplay_ready())
	service.begin_realtime_match("match-1")

	service.server_packet_dispatcher.dispatch({
		"type": "world_full",
		"match_id": "match-1",
		"baseline_id": "world-baseline-1",
		"sequence": 1,
		"snapshot_id": "world-snapshot-1",
		"is_final_chunk": true,
		"ships": [],
		"bullets": [],
		"asteroids": [],
		"pickups": [],
	})

	service.reset_realtime_session()
	service.begin_realtime_match("match-2")

	service.server_packet_dispatcher.dispatch({
		"type": "world_full",
		"match_id": "match-2",
		"baseline_id": "world-baseline-2",
		"sequence": 2,
		"snapshot_id": "world-snapshot-2",
		"is_final_chunk": true,
		"ships": [],
		"bullets": [],
		"asteroids": [],
		"pickups": [],
	})

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
	assert_eq(applied_packets.size(), 2)
	assert_eq(applied_packets[1]["baseline_id"], "world-baseline-2")
