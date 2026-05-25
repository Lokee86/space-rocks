extends RefCounted

const Packets = preload("res://scripts/networking/packets.gd")

var debug_invincible_input_armed := true
var debug_infinite_lives_input_armed := true
var debug_freeze_world_input_armed := true
var debug_freeze_player_input_armed := true


func handle_input(network_client: NetworkClient) -> void:
	if !Input.is_key_pressed(KEY_F1) && !Input.is_key_pressed(KEY_F2) && !Input.is_key_pressed(KEY_F3) && !Input.is_key_pressed(KEY_F4):
		debug_invincible_input_armed = true
		debug_infinite_lives_input_armed = true
		debug_freeze_world_input_armed = true
		debug_freeze_player_input_armed = true
		return
	if network_client == null || !network_client.is_connected_to_server():
		return
	if Input.is_key_pressed(KEY_F1) && debug_invincible_input_armed:
		debug_invincible_input_armed = false
		network_client.send_packet(Packets.toggle_debug_invincible_packet())
	if Input.is_key_pressed(KEY_F2) && debug_infinite_lives_input_armed:
		debug_infinite_lives_input_armed = false
		network_client.send_packet(Packets.toggle_debug_infinite_lives_packet())
	if Input.is_key_pressed(KEY_F3) && debug_freeze_world_input_armed:
		debug_freeze_world_input_armed = false
		network_client.send_packet(Packets.toggle_debug_freeze_world_packet())
	if Input.is_key_pressed(KEY_F4) && debug_freeze_player_input_armed:
		debug_freeze_player_input_armed = false
		network_client.send_packet(Packets.toggle_debug_freeze_player_packet())
