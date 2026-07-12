extends GutTest

const RealtimeTransportSession := preload("res://scripts/networking/webrtc/realtime_transport_session.gd")
const WebRTCTransport := preload("res://scripts/networking/webrtc/webrtc_transport.gd")

class FakeTransport:
	extends WebRTCTransport

	var start_count := 0
	var poll_count := 0
	var close_count := 0
	var handled_answers: Array = []

	func start() -> void:
		start_count += 1

	func poll() -> void:
		poll_count += 1

	func close() -> void:
		close_count += 1

	func handle_answer(description_type: String, sdp: String) -> void:
		handled_answers.append([description_type, sdp])

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
