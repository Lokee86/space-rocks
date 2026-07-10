extends GutTest

const ClientInboundCoordinator := preload("res://scripts/networking/inbound/client_inbound_coordinator.gd")
const ServerPacketDispatcher := preload("res://scripts/networking/inbound/server_packet_dispatcher.gd")


class FakeTransportSession:
	extends RefCounted

	var answers: Array = []
	var remote_ice: Array = []
	var failures := 0

	func handle_answer(description_type: String, sdp: String) -> void:
		answers.append([description_type, sdp])

	func handle_remote_ice(media: String, index: int, name: String) -> void:
		remote_ice.append([media, index, name])

	func handle_remote_failure() -> void:
		failures += 1


class FakeRealtimePipeline:
	extends RefCounted

	var calls := {}

	func _record(method: String, packet: Dictionary) -> void:
		calls[method] = packet

	func apply_world_full(packet: Dictionary) -> void: _record("apply_world_full", packet)
	func apply_world_delta(packet: Dictionary) -> void: _record("apply_world_delta", packet)
	func apply_asteroid_delta(packet: Dictionary) -> void: _record("apply_asteroid_delta", packet)
	func apply_bullet_delta(packet: Dictionary) -> void: _record("apply_bullet_delta", packet)
	func apply_asteroids_lifecycle(packet: Dictionary) -> void: _record("apply_asteroids_lifecycle", packet)
	func apply_bullets_lifecycle(packet: Dictionary) -> void: _record("apply_bullets_lifecycle", packet)
	func apply_overlay_full(packet: Dictionary) -> void: _record("apply_overlay_full", packet)
	func apply_overlay_delta(packet: Dictionary) -> void: _record("apply_overlay_delta", packet)
	func apply_session_full(packet: Dictionary) -> void: _record("apply_session_full", packet)
	func apply_session_delta(packet: Dictionary) -> void: _record("apply_session_delta", packet)
	func apply_event_batch(packet: Dictionary) -> void: _record("apply_event_batch", packet)
	func apply_resync_request(packet: Dictionary) -> void: _record("apply_resync_request", packet)
	func apply_resync_required(packet: Dictionary) -> void: _record("apply_resync_required", packet)


func _configured_coordinator(dispatcher: ServerPacketDispatcher, pipeline: FakeRealtimePipeline, transport = null) -> ClientInboundCoordinator:
	var coordinator := ClientInboundCoordinator.new()
	coordinator.configure(dispatcher, pipeline, transport)
	return coordinator


func test_application_events_relay_original_packets() -> void:
	var dispatcher := ServerPacketDispatcher.new()
	var coordinator := _configured_coordinator(dispatcher, FakeRealtimePipeline.new())
	var received := {}
	add_child_autofree(dispatcher)
	for signal_name in ["authenticate_result_received", "room_snapshot_received", "debug_status_received", "telemetry_pong_received", "unknown_packet_received"]:
		coordinator.connect(signal_name, func(packet: Dictionary) -> void: received[signal_name] = packet)

	var packets := {
		"authenticate_result_received": {"type": "authenticate_result", "authenticated": true},
		"room_snapshot_received": {"type": "room_snapshot", "room_code": "ABCD"},
		"debug_status_received": {"type": "debug_status", "enabled": true},
		"telemetry_pong_received": {"type": "telemetry_pong", "sequence": 7},
		"unknown_packet_received": {"type": "not_registered", "value": 9},
	}
	for signal_name in packets:
		dispatcher.emit_signal(signal_name, packets[signal_name])
		assert_true(received[signal_name] == packets[signal_name])


func test_realtime_packets_route_to_pipeline_unchanged() -> void:
	var dispatcher := ServerPacketDispatcher.new()
	var pipeline := FakeRealtimePipeline.new()
	_configured_coordinator(dispatcher, pipeline)
	add_child_autofree(dispatcher)
	var routes := {
		"world_full_received": ["apply_world_full", {"type": "world_full", "sequence": 1}],
		"event_batch_received": ["apply_event_batch", {"type": "event_batch", "sequence": 2}],
		"resync_required_received": ["apply_resync_required", {"type": "resync_required", "lane": "world"}],
	}
	for signal_name in routes:
		dispatcher.emit_signal(signal_name, routes[signal_name][1])
		assert_true(pipeline.calls[routes[signal_name][0]] == routes[signal_name][1])


func test_transport_control_routing_and_ready_signal() -> void:
	var dispatcher := ServerPacketDispatcher.new()
	var transport := FakeTransportSession.new()
	var coordinator := _configured_coordinator(dispatcher, FakeRealtimePipeline.new(), transport)
	add_child_autofree(dispatcher)
	watch_signals(coordinator)

	dispatcher.emit_signal("webrtc_answer_received", {"description_type": "answer", "sdp": "remote-sdp"})
	dispatcher.emit_signal("webrtc_ice_candidate_received", {"media": "video", "index": 3, "name": "candidate"})
	dispatcher.emit_signal("webrtc_failed_received", {"error_code": "failed"})
	dispatcher.emit_signal("webrtc_ready_received", {"type": "webrtc_ready"})

	assert_eq(transport.answers, [["answer", "remote-sdp"]])
	assert_eq(transport.remote_ice, [["video", 3, "candidate"]])
	assert_eq(transport.failures, 1)
	assert_signal_emit_count(coordinator, "realtime_transport_ready", 1)


func test_configure_is_idempotent() -> void:
	var dispatcher := ServerPacketDispatcher.new()
	var pipeline := FakeRealtimePipeline.new()
	var transport := FakeTransportSession.new()
	var coordinator := _configured_coordinator(dispatcher, pipeline, transport)
	add_child_autofree(dispatcher)
	coordinator.configure(dispatcher, pipeline, transport)

	dispatcher.emit_signal("webrtc_answer_received", {"description_type": "answer", "sdp": "once"})
	assert_eq(transport.answers, [["answer", "once"]])
