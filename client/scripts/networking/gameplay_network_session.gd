extends RefCounted

const NetworkClientScript = preload("res://scripts/networking/network_client.gd")
const Packets = preload("res://scripts/networking/packets.gd")

var owner_node
var injected_network_client
var connected_callback: Callable
var closed_callback: Callable
var packet_received_callback: Callable
var parse_failed_callback: Callable
var store_room_state_callback: Callable
var forward_packet_to_shell_callback: Callable
var apply_gameplay_state_callback: Callable


func configure(
	owner,
	injected_client,
	connected_to_server_callback: Callable,
	connection_closed_callback: Callable,
	packet_received_callback_method: Callable,
	packet_parse_failed_callback: Callable,
	store_room_state_callback_method: Callable,
	forward_packet_to_shell_callback_method: Callable,
	apply_gameplay_state_callback_method: Callable
) -> void:
	owner_node = owner
	injected_network_client = injected_client
	connected_callback = connected_to_server_callback
	closed_callback = connection_closed_callback
	packet_received_callback = packet_received_callback_method
	parse_failed_callback = packet_parse_failed_callback
	store_room_state_callback = store_room_state_callback_method
	forward_packet_to_shell_callback = forward_packet_to_shell_callback_method
	apply_gameplay_state_callback = apply_gameplay_state_callback_method


func setup_network_client() -> Dictionary:
	var network_client
	var should_preserve_network_on_exit := false
	if injected_network_client != null:
		network_client = injected_network_client
		should_preserve_network_on_exit = true
		if network_client.get_parent() != owner_node:
			network_client.reparent(owner_node)
	else:
		network_client = NetworkClientScript.new()
		owner_node.add_child(network_client)

	if !network_client.connected_to_server.is_connected(connected_callback):
		network_client.connected_to_server.connect(connected_callback)
	if !network_client.connection_closed.is_connected(closed_callback):
		network_client.connection_closed.connect(closed_callback)
	if !network_client.packet_received.is_connected(packet_received_callback):
		network_client.packet_received.connect(packet_received_callback)
	if !network_client.packet_parse_failed.is_connected(parse_failed_callback):
		network_client.packet_parse_failed.connect(parse_failed_callback)

	return {
		"network_client": network_client,
		"preserve_network_on_exit": should_preserve_network_on_exit,
	}


func disconnect_gameplay_signals(network_client) -> void:
	if network_client.connected_to_server.is_connected(connected_callback):
		network_client.connected_to_server.disconnect(connected_callback)
	if network_client.connection_closed.is_connected(closed_callback):
		network_client.connection_closed.disconnect(closed_callback)
	if network_client.packet_received.is_connected(packet_received_callback):
		network_client.packet_received.disconnect(packet_received_callback)
	if network_client.packet_parse_failed.is_connected(parse_failed_callback):
		network_client.packet_parse_failed.disconnect(parse_failed_callback)


func release_network_client_for_lobby(network_client) -> Dictionary:
	if network_client == null:
		return {
			"network_client": null,
			"preserve_network_on_exit": false,
			"released_client": null,
		}

	disconnect_gameplay_signals(network_client)

	var released_client = network_client
	owner_node.set_process(false)
	if released_client.get_parent() == owner_node && owner_node.get_parent() != null:
		released_client.reparent(owner_node.get_parent())

	return {
		"network_client": null,
		"preserve_network_on_exit": true,
		"released_client": released_client,
	}


func websocket_url() -> String:
	return "ws://localhost:8080/ws"


func send_client_config(network_client) -> void:
	if network_client == null || !network_client.is_connected_to_server():
		return

	var visible_size: Vector2 = owner_node.get_viewport_rect().size
	network_client.send_packet(Packets.client_config_packet(
		visible_size.x,
		visible_size.y
	))


func handle_connected(network_client) -> void:
	print("Connected!")
	send_client_config(network_client)


func handle_closed() -> void:
	print("Closed")


func close_network_connection(network_client) -> void:
	if network_client != null:
		await network_client.close_gracefully()


func handle_room_packet(data: Dictionary) -> bool:
	var packet_type := str(data.get(Packets.FIELD_TYPE, ""))
	if packet_type == Packets.TYPE_ROOM_SNAPSHOT || packet_type == Packets.TYPE_ROOM_STATE_CHANGED:
		store_room_state_callback.call(data)
		forward_packet_to_shell_callback.call(data)
		return true

	if packet_type == Packets.TYPE_ROOM_ERROR:
		forward_packet_to_shell_callback.call(data)
		return true

	return false


func handle_packet_received(data: Dictionary) -> void:
	if handle_room_packet(data):
		return

	apply_gameplay_state_callback.call(data)


func handle_packet_parse_failed(text: String) -> void:
	print("bad json: ", text)
