extends GutTest

const NetworkClient := preload("res://scripts/networking/network_client.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")

var _had_original_origin := false
var _original_origin := ""


class FakeSocket:
	extends RefCounted

	var handshake_headers := PackedStringArray()
	var sent_texts: Array[String] = []
	var packets: Array = []
	var ready_state: int = WebSocketPeer.STATE_CLOSED
	var close_code := 1000
	var connect_error: Error = OK

	func poll() -> void:
		pass

	func get_ready_state() -> int:
		return ready_state

	func get_close_code() -> int:
		return close_code

	func get_available_packet_count() -> int:
		return packets.size()

	func get_packet():
		if packets.is_empty():
			return PackedByteArray()
		return packets.pop_front()

	func send_text(text: String) -> void:
		sent_texts.append(text)

	func connect_to_url(url: String) -> Error:
		return connect_error

	func close(code: int, reason: String) -> void:
		ready_state = WebSocketPeer.STATE_CLOSED


class FakeWriter extends RefCounted:
	var written_lines: Array[String] = []

	func write_line(line: String) -> void:
		written_lines.append(line)

	func close() -> void:
		pass


func before_each() -> void:
	ClientLogger.reset_for_tests()
	_had_original_origin = OS.has_environment(Constants.MULTIPLAYER_WS_ORIGIN_ENV)
	_original_origin = OS.get_environment(Constants.MULTIPLAYER_WS_ORIGIN_ENV)
	OS.unset_environment(Constants.MULTIPLAYER_WS_ORIGIN_ENV)


func after_each() -> void:
	ClientLogger.reset_for_tests()
	if _had_original_origin:
		OS.set_environment(Constants.MULTIPLAYER_WS_ORIGIN_ENV, _original_origin)
	else:
		OS.unset_environment(Constants.MULTIPLAYER_WS_ORIGIN_ENV)


func _trace_id() -> String:
	return "00000000-0000-4000-8000-000000000041"


func _last_record(writer: FakeWriter) -> Dictionary:
	return JSON.parse_string(writer.written_lines.back())


func test_connect_uses_local_client_origin_for_insecure_websocket_target() -> void:
	var client := NetworkClient.new()
	add_child_autofree(client)
	var fake_socket := FakeSocket.new()
	client.set_socket_for_tests(fake_socket)

	assert_eq(client.connect_to_server("ws://localhost:8080/ws"), OK)
	assert_eq(
		fake_socket.handshake_headers,
		PackedStringArray(["Origin: http://localhost"])
	)


func test_connect_uses_host_only_origin_for_custom_loopback_port() -> void:
	var client := NetworkClient.new()
	add_child_autofree(client)
	var fake_socket := FakeSocket.new()
	client.set_socket_for_tests(fake_socket)

	assert_eq(client.connect_to_server("ws://127.0.0.1:43127/ws"), OK)
	assert_eq(
		fake_socket.handshake_headers,
		PackedStringArray(["Origin: http://127.0.0.1"])
	)


func test_connect_uses_official_client_origin_for_secure_websocket_target() -> void:
	var client := NetworkClient.new()
	add_child_autofree(client)
	var fake_socket := FakeSocket.new()
	client.set_socket_for_tests(fake_socket)

	assert_eq(client.connect_to_server(Constants.MULTIPLAYER_WS_URL), OK)
	assert_eq(
		fake_socket.handshake_headers,
		PackedStringArray(["Origin: %s" % Constants.MULTIPLAYER_WS_ORIGIN])
	)


func test_connect_accepts_full_websocket_origin_environment_override() -> void:
	OS.set_environment(Constants.MULTIPLAYER_WS_ORIGIN_ENV, " https://client.example.test/ ")
	var client := NetworkClient.new()
	add_child_autofree(client)
	var fake_socket := FakeSocket.new()
	client.set_socket_for_tests(fake_socket)

	client.connect_to_server(Constants.MULTIPLAYER_WS_URL)

	assert_eq(fake_socket.handshake_headers, PackedStringArray(["Origin: https://client.example.test"]))


func test_network_metrics_snapshot_reports_default_transport_with_fake_socket() -> void:
	var client := NetworkClient.new()
	var fake_socket := FakeSocket.new()

	add_child_autofree(client)
	client.set_socket_for_tests(fake_socket)

	var snapshot := client.network_metrics_snapshot()
	assert_eq(typeof(snapshot), TYPE_DICTIONARY)
	assert_eq(snapshot["transport"], "websocket")


func test_poll_records_inbound_metrics_for_valid_packet() -> void:
	var client := NetworkClient.new()
	var fake_socket := FakeSocket.new()
	fake_socket.ready_state = WebSocketPeer.STATE_OPEN
	fake_socket.packets = ["{\"type\":\"example_packet\"}".to_utf8_buffer()]

	add_child_autofree(client)
	client.set_socket_for_tests(fake_socket)
	client.poll()

	var snapshot := client.network_metrics_snapshot()
	assert_eq(snapshot["packets_in"], 1)
	assert_true(int(snapshot["bytes_in"]) > 0)
	assert_true(int(snapshot["last_in_packet_bytes"]) > 0)
	assert_true(int(snapshot["max_in_packet_bytes"]) > 0)
	assert_eq(snapshot["last_packet_type_in"], "example_packet")


func test_poll_records_decode_failure_metrics_for_invalid_packet_json() -> void:
	var client := NetworkClient.new()
	var fake_socket := FakeSocket.new()
	fake_socket.ready_state = WebSocketPeer.STATE_OPEN
	fake_socket.packets = ["{invalid json".to_utf8_buffer()]

	add_child_autofree(client)
	client.set_socket_for_tests(fake_socket)
	client.poll()

	var snapshot := client.network_metrics_snapshot()
	assert_eq(snapshot["decode_failures"], 1)


func test_send_raw_packet_records_outbound_metrics_for_valid_packet() -> void:
	var client := NetworkClient.new()
	var fake_socket := FakeSocket.new()
	fake_socket.ready_state = WebSocketPeer.STATE_OPEN

	add_child_autofree(client)
	client.set_socket_for_tests(fake_socket)
	client.send_raw_packet({"type": "example_packet"})

	var snapshot := client.network_metrics_snapshot()
	assert_eq(snapshot["packets_out"], 1)
	assert_true(int(snapshot["bytes_out"]) > 0)
	assert_true(int(snapshot["last_out_packet_bytes"]) > 0)
	assert_true(int(snapshot["max_out_packet_bytes"]) > 0)
	assert_eq(snapshot["last_packet_type_out"], "example_packet")


func test_send_raw_packet_records_encode_failure_metrics_for_invalid_packet_shape() -> void:
	var client := NetworkClient.new()
	var fake_socket := FakeSocket.new()
	fake_socket.ready_state = WebSocketPeer.STATE_OPEN

	add_child_autofree(client)
	client.set_socket_for_tests(fake_socket)
	client.send_raw_packet({})

	var snapshot := client.network_metrics_snapshot()
	assert_eq(snapshot["encode_failures"], 1)


func test_decode_failure_emits_bounded_canonical_event_with_trace() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var client := NetworkClient.new()
	autofree(client)
	var fake_socket := FakeSocket.new()
	fake_socket.ready_state = WebSocketPeer.STATE_OPEN
	fake_socket.packets = ["{invalid json".to_utf8_buffer()]
	client.set_socket_for_tests(fake_socket)
	client.set_connection_trace_provider(Callable(self, "_trace_id"))

	client.poll()

	var record := _last_record(writer)
	assert_eq(record["event"], ObservabilityContract.EVENT_PACKET_DECODE_FAILED)
	assert_eq(record["trace_id"], _trace_id())
	assert_eq(record["fields"]["error_code"], "invalid_json")
	assert_eq(record["fields"]["raw_byte_count"], 13)
	assert_eq(record["fields"]["raw_text_length"], 13)
	assert_false(record["fields"].has("error"))
	assert_false(writer.written_lines[0].contains("{invalid json"))


func test_encode_failure_emits_bounded_canonical_event_without_packet_payload() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var client := NetworkClient.new()
	autofree(client)
	var fake_socket := FakeSocket.new()
	fake_socket.ready_state = WebSocketPeer.STATE_OPEN
	client.set_socket_for_tests(fake_socket)
	client.set_connection_trace_provider(Callable(self, "_trace_id"))

	client.send_raw_packet({})
	assert_push_error_count(1)

	var record := _last_record(writer)
	assert_eq(record["event"], ObservabilityContract.EVENT_OUTBOUND_PACKET_ENCODE_FAILED)
	assert_eq(record["trace_id"], _trace_id())
	assert_eq(record["fields"]["error_code"], "missing_type")
	assert_eq(record["fields"]["failure_mode"], "invalid_packet_shape")
	assert_false(record["fields"].has("error"))
	assert_false(record["fields"].has("packet_payload"))


func test_close_result_signal_distinguishes_expected_and_unexpected_without_changing_legacy_close_signal() -> void:
	var client := NetworkClient.new()
	add_child_autofree(client)
	var fake_socket := FakeSocket.new()
	fake_socket.ready_state = WebSocketPeer.STATE_CLOSED
	fake_socket.close_code = 1006
	client.set_socket_for_tests(fake_socket)
	var close_results: Array = []
	var legacy_closes: Array[int] = [0]
	client.connection_closed_result.connect(func(close_code: int, expected: bool) -> void:
		close_results.append({"close_code": close_code, "expected": expected})
	)
	client.connection_closed.connect(func() -> void:
		legacy_closes[0] += 1
	)
	client.connected = true
	client.poll()
	assert_eq(close_results, [{"close_code": 1006, "expected": false}])
	assert_eq(legacy_closes[0], 1)

	var graceful_client := NetworkClient.new()
	add_child_autofree(graceful_client)
	var graceful_socket := FakeSocket.new()
	graceful_socket.ready_state = WebSocketPeer.STATE_OPEN
	graceful_client.set_socket_for_tests(graceful_socket)
	var graceful_results: Array = []
	var graceful_legacy_closes: Array[int] = [0]
	graceful_client.connection_closed_result.connect(func(close_code: int, expected: bool) -> void:
		graceful_results.append({"close_code": close_code, "expected": expected})
	)
	graceful_client.connection_closed.connect(func() -> void:
		graceful_legacy_closes[0] += 1
	)
	graceful_client.begin_graceful_close()
	graceful_client.poll()
	assert_eq(graceful_results, [{"close_code": 1000, "expected": true}])
	assert_eq(graceful_legacy_closes[0], 0)

func test_encode_failure_prefers_explicit_operation_trace() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var client := NetworkClient.new()
	autofree(client)
	var fake_socket := FakeSocket.new()
	fake_socket.ready_state = WebSocketPeer.STATE_OPEN
	client.set_socket_for_tests(fake_socket)
	client.set_connection_trace_provider(Callable(self, "_trace_id"))

	var operation_trace_id := "00000000-0000-4000-8000-000000000099"
	client.send_raw_packet({}, operation_trace_id)
	assert_push_error_count(1)

	var record := _last_record(writer)
	assert_eq(record["event"], ObservabilityContract.EVENT_OUTBOUND_PACKET_ENCODE_FAILED)
	assert_eq(record["trace_id"], operation_trace_id)
