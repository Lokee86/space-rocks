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


func test_packet_handlers_forward_control_values_to_transport_session() -> void:
	var coordinator := ClientInboundCoordinator.new()
	var transport := FakeTransportSession.new()
	coordinator.set_realtime_transport_session(transport)

	coordinator.handle_webrtc_answer({"description_type": "answer", "sdp": "remote-sdp"})
	coordinator.handle_webrtc_ice_candidate({"media": "video", "index": 3, "name": "candidate"})
	coordinator.handle_webrtc_failed({"error_code": "failed"})

	assert_eq(transport.answers, [["answer", "remote-sdp"]])
	assert_eq(transport.remote_ice, [["video", 3, "candidate"]])
	assert_eq(transport.failures, 1)


func test_ready_emits_once_and_smoke_does_not_emit_ready_without_transport() -> void:
	var coordinator := ClientInboundCoordinator.new()
	watch_signals(coordinator)

	coordinator.handle_webrtc_ready({"type": "webrtc_ready"})
	coordinator.handle_webrtc_smoke({"type": "webrtc_smoke"})

	assert_signal_emit_count(coordinator, "realtime_transport_ready", 1)


func test_configure_routes_all_dispatcher_webrtc_packets_idempotently() -> void:
	var coordinator := ClientInboundCoordinator.new()
	var dispatcher := ServerPacketDispatcher.new()
	var transport := FakeTransportSession.new()
	add_child_autofree(dispatcher)
	coordinator.configure(dispatcher, transport)
	coordinator.configure(dispatcher, transport)
	watch_signals(coordinator)

	dispatcher.dispatch({"type": "webrtc_answer", "description_type": "answer", "sdp": "dispatcher-sdp"})
	dispatcher.dispatch({"type": "webrtc_ice_candidate", "media": "audio", "index": 2, "name": "dispatcher-candidate"})
	dispatcher.dispatch({"type": "webrtc_failed", "error_code": "remote_failure", "message": "failed"})
	dispatcher.dispatch({"type": "webrtc_ready"})
	dispatcher.dispatch({"type": "webrtc_smoke", "smoke_id": "smoke-1", "origin": "server", "message": "diagnostic"})

	assert_eq(transport.answers, [["answer", "dispatcher-sdp"]])
	assert_eq(transport.remote_ice, [["audio", 2, "dispatcher-candidate"]])
	assert_eq(transport.failures, 1)
	assert_signal_emit_count(coordinator, "realtime_transport_ready", 1)


func test_cleared_transport_makes_control_handlers_no_ops_until_replaced() -> void:
	var coordinator := ClientInboundCoordinator.new()
	var first_transport := FakeTransportSession.new()
	var second_transport := FakeTransportSession.new()
	coordinator.set_realtime_transport_session(first_transport)
	coordinator.set_realtime_transport_session(null)

	coordinator.handle_webrtc_answer({"description_type": "answer", "sdp": "ignored"})
	coordinator.handle_webrtc_ice_candidate({"media": "audio", "index": 1, "name": "ignored"})
	coordinator.handle_webrtc_failed({})

	assert_true(first_transport.answers.is_empty())
	assert_true(first_transport.remote_ice.is_empty())
	assert_eq(first_transport.failures, 0)

	coordinator.set_realtime_transport_session(second_transport)
	coordinator.handle_webrtc_answer({"description_type": "answer", "sdp": "replacement"})

	assert_eq(second_transport.answers, [["answer", "replacement"]])
