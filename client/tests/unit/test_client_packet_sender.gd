extends GutTest

const ClientPacketSender := preload("res://scripts/networking/outbound/client_packet_sender.gd")
const NetworkClient := preload("res://scripts/networking/network_client.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")

class FakeNetworkClient:
	extends NetworkClient
	var sent_packets: Array[Dictionary] = []
	func send_raw_packet(packet: Dictionary) -> void:
		sent_packets.append(packet)

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
	ClientLogger.disable()
	var sender := ClientPacketSender.new()
	sender._missing_network_client_reported = false
	sender.send_input_packet({"type": "input"})
	sender.send_input_packet({"type": "input"})
	assert_true(sender._missing_network_client_reported)

	var fake_network := FakeNetworkClient.new()
	sender.network_client = fake_network
	sender.send_input_packet({"type": "input"})
	assert_false(sender._missing_network_client_reported)
	assert_eq(fake_network.sent_packets.size(), 1)
	ClientLogger.reset_for_tests()
