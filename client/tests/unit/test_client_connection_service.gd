extends GutTest

const ClientConnectionService := preload("res://scripts/networking/client_connection_service.gd")
const ClientPacketSender := preload("res://scripts/networking/outbound/client_packet_sender.gd")
const RealtimeTransportSession := preload("res://scripts/networking/webrtc/realtime_transport_session.gd")
const RealtimePacketPipeline := preload("res://scripts/networking/realtime/realtime_packet_pipeline.gd")
const ServerPacketDispatcher := preload("res://scripts/networking/inbound/server_packet_dispatcher.gd")
const WebRTCTransport := preload("res://scripts/networking/webrtc/webrtc_transport.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")
const ClientOperationTrace := preload("res://scripts/observability/client_operation_trace.gd")


class FakeNetworkClient:
	extends NetworkClient

	var sent_packets: Array = []
	var connect_error: Error = OK
	var network_snapshot: Dictionary = {}

	func send_raw_packet(packet: Dictionary, _trace_id: String = "") -> void:
		sent_packets.append(packet)

	func is_connected_to_server() -> bool:
		return true

	func connect_to_server(_url: String) -> Error:
		return connect_error

	func network_metrics_snapshot() -> Dictionary:
		return network_snapshot.duplicate(true)

	func begin_graceful_close() -> bool:
		return true

	func close_gracefully() -> void:
		pass


class FakeTransportPeer:
	extends WebRTCTransport

	var close_calls := 0
	var tooling_packets: Array = []
	var network_snapshot: Dictionary = {}

	func start() -> void:
		pass

	func poll() -> void:
		pass

	func close() -> void:
		close_calls += 1

	func send_tooling_json(packet: Dictionary) -> void:
		tooling_packets.append(packet)

	func network_metrics_snapshot() -> Dictionary:
		return network_snapshot.duplicate(true)


class FakeRealtimePacketPipeline:
	extends RealtimePacketPipeline

	var recovery_calls := 0

	func recover_active_match_baseline() -> void:
		recovery_calls += 1


func _make_fake_transport_peer(fake_peer: FakeTransportPeer) -> WebRTCTransport:
	return fake_peer


class FakeWriter extends RefCounted:
	var written_lines: Array[String] = []
	var failure_count := 0
	var last_failure_message := ""

	func write_line(line: String) -> void:
		written_lines.append(line)

	func close() -> void:
		pass


func _new_service() -> ClientConnectionService:
	var service := ClientConnectionService.new()
	autofree(service)
	return service


func _new_fake_network_client() -> FakeNetworkClient:
	var network_client := FakeNetworkClient.new()
	autofree(network_client)
	return network_client


func before_each() -> void:
	ClientLogger.reset_for_tests()


func after_each() -> void:
	ClientLogger.reset_for_tests()


func test_authenticated_integer_websocket_user_id_is_an_identity() -> void:
	var service := _new_service()
	service._on_authenticate_result_received({"authenticated": true, "user_id": 42})

	assert_true(service.is_websocket_auth_authenticated())
	assert_eq(service.websocket_auth_user_id, 42)
	assert_true(service.has_websocket_auth_identity())


func test_unauthenticated_or_malformed_websocket_user_id_uses_sentinel() -> void:
	var cases := [
		{"authenticated": false, "user_id": 42},
		{"authenticated": true},
		{"authenticated": true, "user_id": "42"},
		{"authenticated": true, "user_id": 42.0},
	]

	for packet in cases:
		var service := _new_service()
		service._on_authenticate_result_received(packet)

		assert_eq(service.websocket_auth_user_id, -1)
		assert_false(service.has_websocket_auth_identity())

func test_missing_packet_sender_is_reported_once_then_resets_on_assignment() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var service := _new_service()
	service._missing_client_packet_sender_reported = false
	service.send_input_packet({"type": "input"})
	service.send_input_packet({"type": "input"})
	assert_push_error_count(1)
	assert_true(service._missing_client_packet_sender_reported)
	assert_eq(writer.written_lines.size(), 1)
	var record = JSON.parse_string(writer.written_lines[0])
	assert_eq(record["event"], ObservabilityContract.EVENT_CLIENT_DEPENDENCY_UNAVAILABLE)
	assert_eq(record["fields"]["subsystem"], "networking_outbound")
	assert_eq(record["fields"]["dependency"], "client_packet_sender")
	assert_eq(record["fields"]["failure_mode"], "not_configured")

	var fake_network := _new_fake_network_client()
	var sender := ClientPacketSender.new(fake_network)
	service.client_packet_sender = sender
	service.send_input_packet({"type": "input"})
	assert_false(service._missing_client_packet_sender_reported)
	assert_eq(fake_network.sent_packets.size(), 1)


func test_close_resets_websocket_auth_identity() -> void:
	var service := _new_service()
	service.websocket_auth_authenticated = true
	service.websocket_auth_user_id = 42
	service.websocket_auth_display_name = "Pilot"
	service._on_closed()

	assert_false(service.is_websocket_auth_authenticated())
	assert_eq(service.websocket_auth_user_id, -1)
	assert_false(service.has_websocket_auth_identity())
	assert_eq(service.websocket_auth_display_name, "")


func test_protocol_match_begin_end_preserves_transport_without_close() -> void:
	var service := _new_service()
	add_child(service)
	var peer := FakeTransportPeer.new()
	var retained := RealtimeTransportSession.new()
	retained.transport = peer
	service.realtime_transport_session = retained
	service.begin_realtime_match("match-1")
	service.end_realtime_match()
	assert_true(service.realtime_transport_session == retained)
	assert_true(service.realtime_transport_session.transport == peer)
	assert_eq(peer.close_calls, 0)


func test_realtime_transport_signals_forward_and_recovery_calls_pipeline() -> void:
	var service := _new_service()
	var pipeline := FakeRealtimePacketPipeline.new()
	var forwarded_tooling: Array = []
	var started_lanes: Array = []
	service.realtime_packet_pipeline = pipeline
	service._ensure_realtime_transport_session()
	service.tooling_packet_received.connect(func(packet: Dictionary) -> void:
		forwarded_tooling.append(packet)
	)
	service.recovery_started.connect(func(lane: String) -> void:
		started_lanes.append(lane)
	)
	assert_true(service.realtime_transport_session.has_signal("recovery_succeeded"))
	assert_true(service.realtime_transport_session.is_connected("recovery_succeeded", Callable(service, "_on_recovery_succeeded")))
	assert_true(service.realtime_transport_session.is_connected("recovery_failed", Callable(service, "_on_recovery_failed")))

	service.realtime_transport_session.tooling_packet_received.emit({"type": "tooling_packet"})
	service.realtime_transport_session.recovery_started.emit("world")
	service._on_recovery_succeeded()
	service._on_recovery_failed()

	assert_eq(forwarded_tooling, [{"type": "tooling_packet"}])
	assert_eq(started_lanes, ["world"])
	assert_eq(pipeline.recovery_calls, 1)


func test_recovery_failure_marks_replay_unavailable_once() -> void:
	var service := _new_service()
	var availability_changes: Array = []
	service.realtime_replay_availability_changed.connect(func(available: bool) -> void:
		availability_changes.append(available)
	)

	service._on_recovery_failed()
	service._on_recovery_failed()

	assert_false(service.is_realtime_replay_available())
	assert_eq(availability_changes, [false])


func test_new_connection_attempt_restores_replay_availability() -> void:
	var service := _new_service()
	var network_client := _new_fake_network_client()
	var availability_changes: Array = []
	service.network_client = network_client
	service.realtime_replay_availability_changed.connect(func(available: bool) -> void:
		availability_changes.append(available)
	)
	service._on_recovery_failed()

	assert_eq(service.connect_to_server("ws://example"), OK)
	assert_true(service.is_realtime_replay_available())
	assert_eq(availability_changes, [false, true])


func test_send_tooling_packet_routes_through_realtime_transport() -> void:
	var service := _new_service()
	var peer := FakeTransportPeer.new()
	service.webrtc_transport_factory = Callable(self, "_make_fake_transport_peer").bind(peer)
	service._ensure_realtime_transport_session()
	service.realtime_transport_session.start()
	service.send_tooling_packet({"type": "tooling_packet", "value": 1})

	assert_eq(peer.tooling_packets, [{"type": "tooling_packet", "value": 1}])


func test_telemetry_ping_uses_tooling_transport_and_request_correlation() -> void:
	var trace_id := "00000000-0000-4000-8000-000000000095"
	var service := _new_service()
	var peer := FakeTransportPeer.new()
	service._operation_trace_factory = func(operation_name: String):
		return ClientOperationTrace.new(operation_name, func() -> String: return trace_id)
	service.webrtc_transport_factory = Callable(self, "_make_fake_transport_peer").bind(peer)
	service._ensure_realtime_transport_session()
	service.realtime_transport_session.start()
	service._tooling_ready = true

	assert_true(service.send_telemetry_ping(7, 1234))
	assert_eq(peer.tooling_packets.size(), 1)
	assert_eq(peer.tooling_packets[0]["type"], "telemetry_ping")
	assert_eq(peer.tooling_packets[0]["request_id"], trace_id)
	assert_eq(peer.tooling_packets[0]["sequence"], 7)
	assert_eq(peer.tooling_packets[0]["client_sent_msec"], 1234)


func test_telemetry_subscription_follows_visibility_room_and_recovery() -> void:
	var trace_index := {"value": 0}
	var service := _new_service()
	var peer := FakeTransportPeer.new()
	service._operation_trace_factory = func(operation_name: String):
		trace_index["value"] += 1
		var trace_id := "00000000-0000-4000-8000-%012d" % trace_index["value"]
		return ClientOperationTrace.new(operation_name, func() -> String: return trace_id)
	service.webrtc_transport_factory = Callable(self, "_make_fake_transport_peer").bind(peer)
	service._ensure_realtime_transport_session()
	service.realtime_transport_session.start()
	service._active_room_code = "ROOM1"
	service._active_room_state = "in_game"
	service._tooling_ready = true

	service.set_telemetry_subscription_enabled(true)
	service.set_telemetry_subscription_enabled(true)
	assert_eq(peer.tooling_packets.filter(func(packet): return packet["type"] == "telemetry_subscribe").size(), 1)

	service._on_recovery_started("tooling")
	service._on_realtime_transport_ready()
	assert_eq(peer.tooling_packets.filter(func(packet): return packet["type"] == "telemetry_subscribe").size(), 2)

	service.set_telemetry_subscription_enabled(false)
	assert_eq(peer.tooling_packets.filter(func(packet): return packet["type"] == "telemetry_unsubscribe").size(), 1)


func test_telemetry_subscription_clears_locally_after_room_detach() -> void:
	var service := _new_service()
	var peer := FakeTransportPeer.new()
	service.webrtc_transport_factory = Callable(self, "_make_fake_transport_peer").bind(peer)
	service._ensure_realtime_transport_session()
	service.realtime_transport_session.start()
	service._active_room_code = "ROOM1"
	service._active_room_state = "in_game"
	service._tooling_ready = true
	service.set_telemetry_subscription_enabled(true)

	service._active_room_code = ""
	service._active_room_state = ""
	service._sync_telemetry_subscription()

	assert_eq(peer.tooling_packets.filter(func(packet): return packet["type"] == "telemetry_subscribe").size(), 1)
	assert_eq(peer.tooling_packets.filter(func(packet): return packet["type"] == "telemetry_unsubscribe").size(), 0)
	assert_eq(service._telemetry_subscription_room_code, "")


func test_network_metrics_merge_websocket_and_webrtc_lane_sources() -> void:
	var service := _new_service()
	var network := _new_fake_network_client()
	var peer := FakeTransportPeer.new()
	network.network_snapshot = {"packets_in": 2, "packets_out": 3, "bytes_in": 20, "bytes_out": 30}
	peer.network_snapshot = {
		"packets_in": 5,
		"packets_out": 7,
		"bytes_in": 50,
		"bytes_out": 70,
		"lanes": {"tooling": {"packets_in": 2}},
	}
	service.network_client = network
	service.webrtc_transport_factory = Callable(self, "_make_fake_transport_peer").bind(peer)
	service._ensure_realtime_transport_session()
	service.realtime_transport_session.start()

	var snapshot := service.network_metrics_snapshot()
	assert_eq(snapshot["packets_in"], 7)
	assert_eq(snapshot["packets_out"], 10)
	assert_eq(snapshot["bytes_in"], 70)
	assert_eq(snapshot["bytes_out"], 100)
	assert_eq(snapshot["webrtc_lanes"]["tooling"]["packets_in"], 2)


func test_tooling_command_result_is_forwarded_from_tooling_router() -> void:
	var service := _new_service()
	add_child(service)
	var received: Array = []
	service.tooling_command_result_received.connect(func(packet: Dictionary) -> void:
		received.append(packet)
	)
	var packet := {
		"type": "tooling_command_result",
		"request_id": "request-1",
		"command_type": "debug_clear_bullets",
		"applied": true,
	}

	service._on_tooling_packet_received(packet)

	assert_eq(received, [packet])


func test_debug_readouts_are_forwarded_from_tooling_router() -> void:
	var service := _new_service()
	add_child(service)
	var statuses: Array = []
	var catalogs: Array = []
	service.debug_status_received.connect(func(packet: Dictionary) -> void:
		statuses.append(packet)
	)
	service.debug_shape_catalog_received.connect(func(packet: Dictionary) -> void:
		catalogs.append(packet)
	)
	var status_packet := {"type": "debug_status", "request_id": "status-1"}
	var catalog_packet := {"type": "debug_shape_catalog", "request_id": "catalog-1", "shapes": {}}

	service._on_tooling_packet_received(status_packet)
	service._on_tooling_packet_received(catalog_packet)

	assert_eq(statuses, [status_packet])
	assert_eq(catalogs, [catalog_packet])


func test_tooling_ready_requests_room_debug_readouts_once() -> void:
	var trace_ids := [
		"00000000-0000-4000-8000-000000000091",
		"00000000-0000-4000-8000-000000000092",
		"00000000-0000-4000-8000-000000000093",
		"00000000-0000-4000-8000-000000000094",
	]
	var trace_state := {"index": 0}
	var service := _new_service()
	var peer := FakeTransportPeer.new()
	service._operation_trace_factory = func(operation_name: String):
		var trace_id: String = trace_ids[trace_state["index"]]
		trace_state["index"] += 1
		return ClientOperationTrace.new(operation_name, func() -> String: return trace_id)
	service.webrtc_transport_factory = Callable(self, "_make_fake_transport_peer").bind(peer)
	add_child(service)
	service._ensure_realtime_transport_session()
	service.realtime_transport_session.start()

	service._on_room_snapshot_received({"room_code": "ROOM1", "room_state": "in_game"})
	assert_eq(peer.tooling_packets.size(), 0)
	service._on_realtime_transport_ready()
	service._on_room_snapshot_received({"room_code": "ROOM1", "room_state": "in_game"})

	assert_eq(peer.tooling_packets.size(), 2)
	assert_eq(peer.tooling_packets[0]["type"], "debug_status_subscribe")
	assert_eq(peer.tooling_packets[0]["request_id"], trace_ids[0])
	assert_eq(peer.tooling_packets[1]["type"], "debug_shape_catalog_request")
	assert_eq(peer.tooling_packets[1]["request_id"], trace_ids[1])

	var replacement_peer := FakeTransportPeer.new()
	service.reset_realtime_session()
	service.webrtc_transport_factory = Callable(self, "_make_fake_transport_peer").bind(replacement_peer)
	service._ensure_realtime_transport_session()
	service.realtime_transport_session.start()
	service._on_realtime_transport_ready()

	assert_eq(replacement_peer.tooling_packets.size(), 2)
	assert_eq(replacement_peer.tooling_packets[0]["type"], "debug_status_subscribe")
	assert_eq(replacement_peer.tooling_packets[0]["request_id"], trace_ids[2])
	assert_eq(replacement_peer.tooling_packets[1]["type"], "debug_shape_catalog_request")
	assert_eq(replacement_peer.tooling_packets[1]["request_id"], trace_ids[3])


func test_resync_required_uses_active_match_and_suppresses_when_inactive() -> void:
	var service := _new_service()
	add_child(service)
	var fake_network := _new_fake_network_client()
	add_child(fake_network)
	service.network_client = fake_network
	service.client_packet_sender = ClientPacketSender.new(fake_network)
	service.begin_realtime_match("match-1")
	service._on_resync_request_required("world", "baseline-1", 1, "missing_baseline")
	assert_eq(fake_network.sent_packets.size(), 1)
	assert_eq(fake_network.sent_packets[0].get("match_id"), "match-1")
	service.end_realtime_match()
	service._on_resync_request_required("world", "baseline-1", 1, "missing_baseline")
	assert_eq(fake_network.sent_packets.size(), 1)


func test_connection_service_does_not_expose_raw_webrtc_inbound_signals() -> void:
	var service := _new_service()
	add_child(service)

	assert_false(service.has_signal("webrtc_answer_received"))
	assert_false(service.has_signal("webrtc_ice_candidate_received"))
	assert_false(service.has_signal("webrtc_ready_received"))
	assert_false(service.has_signal("webrtc_smoke_received"))
	assert_false(service.has_signal("webrtc_failed_received"))
	assert_true(service.has_signal("realtime_transport_ready"))


func test_inbound_valid_gameplay_packet_routes_through_pipeline_once() -> void:
	var service := _new_service()
	var callback_state := {"pipeline_packet_count": 0, "state_seen": false}
	add_child(service)

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
	var service := _new_service()
	var callback_state := {"pipeline_packet_count": 0}
	var fake_network := _new_fake_network_client()
	service.network_client = fake_network
	service.client_packet_sender = ClientPacketSender.new(fake_network)
	service.server_packet_dispatcher = ServerPacketDispatcher.new()
	service.webrtc_transport_factory = Callable(self, "_make_fake_transport_peer").bind(FakeTransportPeer.new())
	add_child(service)
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
	}, "world")

	assert_eq(callback_state.pipeline_packet_count, 2)
	
	assert_false(service.get_realtime_packet_pipeline().is_gameplay_ready())


func test_clock_offset_is_forwarded_and_reset() -> void:
	var service := _new_service()
	var peer := FakeTransportPeer.new()
	service.webrtc_transport_factory = Callable(self, "_make_fake_transport_peer").bind(peer)
	service.set_server_clock_offset_ms(125)
	add_child(service)
	service._ensure_realtime_transport_session()
	service.realtime_transport_session.start()
	assert_eq(service.realtime_transport_session.transport.server_clock_offset_ms, 125)
	service.set_server_clock_offset_ms(250)
	assert_eq(service.realtime_transport_session.transport.server_clock_offset_ms, 250)
	service.reset_realtime_session()
	assert_eq(service.server_clock_offset_ms, -1)


func test_reset_exposes_fresh_pipeline_and_readiness() -> void:
	var service := _new_service()
	add_child(service)

	var pipeline = service.get_realtime_packet_pipeline()
	var presentation_state = pipeline.get_presentation_state()
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

func test_connection_trace_survives_connect_and_is_visible_during_close() -> void:
	var state := {"index": 0}
	var ids := [
		"00000000-0000-4000-8000-000000000011",
		"00000000-0000-4000-8000-000000000012",
	]
	var factory := func(operation_name: String):
		var trace_id: String = ids[state["index"]]
		state["index"] += 1
		return ClientOperationTrace.new(operation_name, func() -> String: return trace_id)

	var service := _new_service()
	service._operation_trace_factory = factory
	service.network_client = _new_fake_network_client()
	var closed_trace_ids: Array[String] = []
	service.closed.connect(func() -> void: closed_trace_ids.append(service.current_connection_trace_id()))

	assert_eq(service.connect_to_server("ws://example"), OK)
	assert_eq(service.current_connection_trace_id(), ids[0])

	service._on_connected()
	assert_eq(service.current_connection_trace_id(), ids[0])
	service._on_closed()

	assert_eq(closed_trace_ids, [ids[0]])
	assert_eq(service.current_connection_trace_id(), "")
	assert_eq(service.connect_to_server("ws://example"), OK)
	assert_eq(service.current_connection_trace_id(), ids[1])


func test_room_operation_trace_is_retained_until_explicitly_cleared() -> void:
	var service := _new_service()
	service.begin_room_operation("join_room", "00000000-0000-4000-8000-000000000013")

	assert_eq(service.active_room_operation_type(), "join_room")
	assert_eq(service.active_room_operation_trace_id(), "00000000-0000-4000-8000-000000000013")

	service.clear_room_operation_context()

	assert_eq(service.active_room_operation_type(), "")
	assert_eq(service.active_room_operation_trace_id(), "")

func test_graceful_close_does_not_emit_client_disconnected() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var service := _new_service()
	service._operation_trace_factory = func(operation_name: String):
		return ClientOperationTrace.new(operation_name, func() -> String: return "00000000-0000-4000-8000-000000000081")
	service.network_client = _new_fake_network_client()
	service.connect_to_server("ws://example")
	service._on_connected()
	service._on_connection_close_result(1000, true)

	var events: Array[String] = []
	for line in writer.written_lines:
		events.append(str(JSON.parse_string(line)["event"]))
	assert_eq(events.count(ObservabilityContract.EVENT_CLIENT_DISCONNECTED), 0)


func test_close_before_connected_emits_connection_failed() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var service := _new_service()
	service._operation_trace_factory = func(operation_name: String):
		return ClientOperationTrace.new(operation_name, func() -> String: return "00000000-0000-4000-8000-000000000082")
	service.network_client = _new_fake_network_client()
	service.connect_to_server("ws://example")
	service._on_connection_close_result(1006, false)
	service._on_closed()

	var records: Array = []
	for line in writer.written_lines:
		records.append(JSON.parse_string(line))
	assert_eq(records[-1]["event"], ObservabilityContract.EVENT_CLIENT_CONNECTION_FAILED)
	assert_eq(records[-1]["fields"]["failure_mode"], "socket_closed_before_connected")


func test_attempt_and_connected_events_share_one_trace() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	ClientLogger.enable_debug()
	var service := _new_service()
	service._operation_trace_factory = func(operation_name: String):
		return ClientOperationTrace.new(operation_name, func() -> String: return "00000000-0000-4000-8000-000000000083")
	service.network_client = _new_fake_network_client()

	service.connect_to_server("ws://example")
	service._on_connected()

	var records: Array = []
	for line in writer.written_lines:
		records.append(JSON.parse_string(line))
	assert_eq(records.size(), 2)
	assert_eq(records[0]["event"], ObservabilityContract.EVENT_CONNECTION_ATTEMPT_STARTED)
	assert_eq(records[1]["event"], ObservabilityContract.EVENT_CLIENT_CONNECTED)
	assert_eq(records[0]["trace_id"], records[1]["trace_id"])


func test_immediate_connection_failure_emits_once_and_clears_trace() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var service := _new_service()
	service._operation_trace_factory = func(operation_name: String):
		return ClientOperationTrace.new(operation_name, func() -> String: return "00000000-0000-4000-8000-000000000084")
	var network_client := _new_fake_network_client()
	network_client.connect_error = ERR_CANT_CONNECT
	service.network_client = network_client

	assert_eq(service.connect_to_server("ws://example"), ERR_CANT_CONNECT)

	var records: Array = []
	for line in writer.written_lines:
		records.append(JSON.parse_string(line))
	var failure_count := 0
	for record in records:
		if record["event"] == ObservabilityContract.EVENT_CLIENT_CONNECTION_FAILED:
			failure_count += 1
	assert_eq(failure_count, 1)
	assert_eq(service.current_connection_trace_id(), "")


func test_unexpected_close_after_connected_emits_disconnected() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var service := _new_service()
	service._operation_trace_factory = func(operation_name: String):
		return ClientOperationTrace.new(operation_name, func() -> String: return "00000000-0000-4000-8000-000000000085")
	service.network_client = _new_fake_network_client()

	service.connect_to_server("ws://example")
	service._on_connected()
	service._on_connection_close_result(1006, false)
	service._on_closed()

	var records: Array = []
	for line in writer.written_lines:
		records.append(JSON.parse_string(line))
	assert_eq(records[-1]["event"], ObservabilityContract.EVENT_CLIENT_DISCONNECTED)
	assert_eq(records[-1]["fields"]["failure_mode"], "unexpected_close")
func test_active_room_operation_trace_reaches_room_packet() -> void:
	var trace_id := "00000000-0000-4000-8000-000000000026"
	var fake_network := _new_fake_network_client()
	var service := _new_service()
	service.client_packet_sender = ClientConnectionService.ClientPacketSender.new(fake_network)
	service.begin_room_operation("join_room", trace_id)

	service.send_join_room_request("ROOM1")

	assert_eq(fake_network.sent_packets.size(), 1)
	assert_eq(fake_network.sent_packets[0]["trace_id"], trace_id)