extends GutTest

const ClientPacketSender := preload("res://scripts/networking/outbound/client_packet_sender.gd")
const NetworkClient := preload("res://scripts/networking/network_client.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")

class FakeNetworkClient:
	extends NetworkClient
	var sent_packets: Array[Dictionary] = []
	func send_raw_packet(packet: Dictionary, trace_id: String = "") -> void:
		sent_packets.append(packet)


class FakeWriter extends RefCounted:
	var written_lines: Array[String] = []
	var failure_count := 0
	var last_failure_message := ""

	func write_line(line: String) -> void:
		written_lines.append(line)

	func close() -> void:
		pass


func before_each() -> void:
	ClientLogger.reset_for_tests()


func after_each() -> void:
	ClientLogger.reset_for_tests()


func test_configure_network_client_resets_missing_dependency_report() -> void:
	var sender := ClientPacketSender.new()
	sender._missing_network_client_reported = true
	var fake_network := FakeNetworkClient.new()
	sender.configure(fake_network)
	sender.send_input_packet({"type": "input"})
	assert_eq(fake_network.sent_packets.size(), 1)
	assert_false(sender._missing_network_client_reported)

func test_empty_auth_token_remains_intentional_noop() -> void:
	var fake_network := FakeNetworkClient.new()
	var sender := ClientPacketSender.new(fake_network)
	sender.send_authenticate_request("")
	assert_eq(fake_network.sent_packets.size(), 0)

func test_missing_network_client_is_reported_once_then_resets_on_assignment() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var sender := ClientPacketSender.new()
	sender._missing_network_client_reported = false
	sender.send_input_packet({"type": "input"})
	sender.send_input_packet({"type": "input"})
	assert_true(sender._missing_network_client_reported)
	assert_eq(writer.written_lines.size(), 1)
	var record = JSON.parse_string(writer.written_lines[0])
	assert_eq(record["event"], ObservabilityContract.EVENT_CLIENT_DEPENDENCY_UNAVAILABLE)
	assert_eq(record["fields"]["subsystem"], "networking_outbound")
	assert_eq(record["fields"]["dependency"], "network_client")
	assert_eq(record["fields"]["failure_mode"], "not_configured")

	var fake_network := FakeNetworkClient.new()
	sender.network_client = fake_network
	sender.send_input_packet({"type": "input"})
	assert_false(sender._missing_network_client_reported)
	assert_eq(fake_network.sent_packets.size(), 1)

	sender.network_client = null
	sender.send_input_packet({"type": "input"})
	assert_push_error_count(2)
	assert_true(sender._missing_network_client_reported)
	assert_eq(writer.written_lines.size(), 2)
