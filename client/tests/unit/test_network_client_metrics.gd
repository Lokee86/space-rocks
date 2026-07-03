extends GutTest

const NetworkClient := preload("res://scripts/networking/network_client.gd")


class FakeSocket:
	extends RefCounted

	var handshake_headers := PackedStringArray()
	var sent_texts: Array[String] = []
	var packets: Array = []
	var ready_state: int = WebSocketPeer.STATE_CLOSED

	func poll() -> void:
		pass

	func get_ready_state() -> int:
		return ready_state

	func get_available_packet_count() -> int:
		return packets.size()

	func get_packet():
		if packets.is_empty():
			return PackedByteArray()
		return packets.pop_front()

	func send_text(text: String) -> void:
		sent_texts.append(text)

	func connect_to_url(url: String) -> Error:
		return OK

	func close(code: int, reason: String) -> void:
		ready_state = WebSocketPeer.STATE_CLOSED


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
