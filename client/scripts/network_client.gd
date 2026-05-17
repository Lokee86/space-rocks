extends Node
class_name NetworkClient

signal connected_to_server
signal connection_closed
signal packet_received(data: Dictionary)
signal packet_parse_failed(text: String)

var socket := WebSocketPeer.new()
var connected := false
var closed_notified := false


func connect_to_server(url: String) -> Error:
	closed_notified = false
	var err := socket.connect_to_url(url)
	if err != OK:
		print("connection failede")
	else:
		print("Connecting...")

	return err


func poll() -> void:
	socket.poll()

	var state := socket.get_ready_state()
	if state == WebSocketPeer.STATE_OPEN:
		if !connected:
			connected = true
			connected_to_server.emit()
	elif state == WebSocketPeer.STATE_CLOSED:
		connected = false
		if !closed_notified:
			closed_notified = true
			connection_closed.emit()

	while socket.get_available_packet_count() > 0:
		var text := socket.get_packet().get_string_from_utf8()
		var data = JSON.parse_string(text)
		if data == null:
			packet_parse_failed.emit(text)
			continue
		if data is Dictionary:
			packet_received.emit(data)


func send_packet(packet: Dictionary) -> void:
	if !is_connected_to_server():
		return

	socket.send_text(JSON.stringify(packet))


func is_connected_to_server() -> bool:
	return socket.get_ready_state() == WebSocketPeer.STATE_OPEN
