extends GutTest

const RealtimeTransportSession := preload("res://scripts/networking/webrtc/realtime_transport_session.gd")
const WebRTCTransport := preload("res://scripts/networking/webrtc/webrtc_transport.gd")

class FakeTransport:
	extends WebRTCTransport

	var start_count := 0
	var poll_count := 0
	var close_count := 0
	var handled_answers: Array = []
	var tooling_packets: Array = []

	func start() -> void:
		start_count += 1
		offer_created.emit("offer", "sdp-%d" % start_count)

	func poll() -> void:
		poll_count += 1

	func close() -> void:
		close_count += 1

	func handle_answer(description_type: String, sdp: String) -> void:
		handled_answers.append([description_type, sdp])

	func send_tooling_json(packet: Dictionary) -> void:
		tooling_packets.append(packet)

func test_start_poll_and_close_delegate_once_and_clear_transport() -> void:
	var session := RealtimeTransportSession.new()
	var transport := FakeTransport.new()

	session.set_server_clock_offset_ms(125)
	session.transport_factory = func():
		return transport
	session.start()
	assert_eq(transport.server_clock_offset_ms, 125)
	session.set_server_clock_offset_ms(250)
	assert_eq(transport.server_clock_offset_ms, 250)
	session.poll()
	session.close()

	assert_eq(transport.start_count, 1)
	assert_eq(transport.poll_count, 1)
	assert_eq(transport.close_count, 1)
	assert_null(session.transport)


func test_packet_routing_separates_tooling_from_gameplay_dispatch() -> void:
	var session := RealtimeTransportSession.new()
	var gameplay_packets: Array = []
	var tooling_packets: Array = []
	session.dispatch_packet = func(packet: Dictionary) -> void:
		gameplay_packets.append(packet)
	session.tooling_packet_received.connect(func(packet: Dictionary) -> void:
		tooling_packets.append(packet)
	)

	session._on_packet_received({"type": "tooling_packet"}, "tooling")
	session._on_packet_received({"type": "world_delta"}, "world")

	assert_eq(tooling_packets, [{"type": "tooling_packet"}])
	assert_eq(gameplay_packets, [{"type": "world_delta"}])


func test_channel_close_replaces_transport_and_starts_new_offer_path() -> void:
	var session := RealtimeTransportSession.new()
	var first_transport := FakeTransport.new()
	var replacement_transport := FakeTransport.new()
	var transports: Array = [first_transport, replacement_transport]
	var started_lanes: Array = []
	var offers: Array = []
	session.transport_factory = func():
		return transports.pop_front()
	session.send_offer = func(description_type: String, sdp: String) -> void:
		offers.append([description_type, sdp])
	session.recovery_started.connect(func(lane: String) -> void:
		started_lanes.append(lane)
	)
	session.start()
	first_transport.channel_closed.emit("overlay")

	assert_eq(first_transport.close_count, 1)
	assert_eq(replacement_transport.start_count, 1)
	assert_eq(started_lanes, ["overlay"])
	assert_eq(offers, [["offer", "sdp-1"], ["offer", "sdp-1"]])
	assert_eq(session.transport, replacement_transport)


func test_recovery_ready_clears_state_and_emits_success() -> void:
	var session := RealtimeTransportSession.new()
	var first_transport := FakeTransport.new()
	var replacement_transport := FakeTransport.new()
	var transports: Array = [first_transport, replacement_transport]
	var success_state := {"count": 0}
	session.transport_factory = func():
		return transports.pop_front()
	session.recovery_succeeded.connect(func(_unused = null) -> void:
		success_state["count"] += 1
	)
	session.start()
	first_transport.channel_closed.emit("world")
	assert_true(session._recovery_active)
	assert_true(replacement_transport.ready.is_connected(session._on_ready))
	replacement_transport.ready.emit([])
	assert_false(session._recovery_active)

	assert_eq(success_state["count"], 1)
	assert_eq(replacement_transport.close_count, 0)
	assert_eq(session.transport, replacement_transport)


func test_two_sequential_successful_recoveries_are_allowed() -> void:
	var session := RealtimeTransportSession.new()
	var first_transport := FakeTransport.new()
	var first_replacement := FakeTransport.new()
	var second_replacement := FakeTransport.new()
	var transports: Array = [first_transport, first_replacement, second_replacement]
	var success_state := {"count": 0}
	session.transport_factory = func():
		return transports.pop_front()
	session.recovery_succeeded.connect(func(_unused = null) -> void:
		success_state["count"] += 1
	)
	session.start()

	first_transport.channel_closed.emit("world")
	first_replacement.ready.emit([])
	first_replacement.channel_closed.emit("overlay")
	second_replacement.ready.emit([])

	assert_eq(success_state["count"], 2)
	assert_eq(first_replacement.close_count, 1)
	assert_eq(second_replacement.close_count, 0)
	assert_eq(session.transport, second_replacement)


func test_recovery_timeout_closes_replacement_and_emits_failure_once() -> void:
	var clock_state := {"now_msec": 1000}
	var session := RealtimeTransportSession.new(func():
		return clock_state["now_msec"]
	)
	var first_transport := FakeTransport.new()
	var replacement_transport := FakeTransport.new()
	var transports: Array = [first_transport, replacement_transport]
	var failure_state := {"count": 0}
	session.transport_factory = func():
		return transports.pop_front()
	session.recovery_failed.connect(func(_unused = null) -> void:
		failure_state["count"] += 1
	)
	session.start()
	first_transport.channel_closed.emit("world")
	assert_true(session._recovery_active)
	assert_eq(session._now_msec(), 1000)
	assert_eq(session._recovery_deadline_msec, 11000)
	clock_state["now_msec"] += 10000
	assert_eq(session._now_msec(), 11000)
	session.poll()
	session.poll()
	assert_false(session._recovery_active)

	assert_eq(failure_state["count"], 1)
	assert_eq(replacement_transport.close_count, 1)
	assert_null(session.transport)


func test_recovery_setup_failure_emits_failure_once() -> void:
	var session := RealtimeTransportSession.new()
	var first_transport := FakeTransport.new()
	var transports: Array = [first_transport, null]
	var failure_state := {"count": 0}
	session.transport_factory = func():
		return transports.pop_front()
	session.recovery_failed.connect(func(_unused = null) -> void:
		failure_state["count"] += 1
	)
	session.start()
	first_transport.channel_closed.emit("bullets")
	first_transport.channel_closed.emit("event")

	assert_eq(failure_state["count"], 1)
	assert_null(session.transport)


func test_send_tooling_packet_uses_current_transport() -> void:
	var session := RealtimeTransportSession.new()
	var transport := FakeTransport.new()
	session.transport_factory = func():
		return transport
	session.start()
	session.send_tooling_packet({"type": "tooling_packet", "value": 1})

	assert_eq(transport.tooling_packets, [{"type": "tooling_packet", "value": 1}])
