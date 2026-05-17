extends Node2D

const PLAYER_SCENE := preload("res://scenes/player.tscn")

@export var background_parallax := 0.25
@export var foreground_background_parallax := 0.45
@export var foreground_background_offset := Vector2(480.0, 270.0)
@export var player_interpolation_speed := 18.0

@onready var player = $Player
@onready var repeated_background: TextureRect = $ParallaxBackground/BackgroundLayer/RepeatedBackground
@onready var repeated_foreground_background: TextureRect = $ParallaxBackground/ForegroundBackgroundLayer/RepeatedBackground

var socket := WebSocketPeer.new()
var connected := false
var self_id := ""
var player_nodes := {}
var initialized_players := {}
var target_player_positions := {}
var target_player_rotations := {}

func _ready() -> void:
	var err := socket.connect_to_url("ws://localhost:8080/ws")
	if err != OK:
		print("connection failede")
	else:
		print("Connecting...")


func _process(delta: float) -> void:
	socket.poll()

	var state := socket.get_ready_state()

	if state == WebSocketPeer.STATE_OPEN:
		if !connected:
			connected = true
			print("Connected!")

		socket.send_text(JSON.stringify(player.get_input_packet()))
	elif state == WebSocketPeer.STATE_CLOSED:
		print("Closed")

	while socket.get_available_packet_count() > 0:
		var text := socket.get_packet().get_string_from_utf8()
		var data = JSON.parse_string(text)

		if data == null:
			print("bad json: ", text)
			return

		_apply_state(data)

	_interpolate_player(delta)
	_update_layer_shader(repeated_background, background_parallax, Vector2.ZERO)
	_update_layer_shader(repeated_foreground_background, foreground_background_parallax, foreground_background_offset)


func _apply_state(data: Dictionary) -> void:
	if data.get("type", "") != "state":
		return

	self_id = data["self_id"]
	var players: Dictionary = data["players"]
	_remove_missing_players(players)

	for player_id in players.keys():
		var state: Dictionary = players[player_id]
		var player_node = _get_player_node(player_id)
		var server_position := Vector2(state["x"], state["y"])
		var server_rotation: float = state["rotation"]

		target_player_positions[player_id] = server_position
		target_player_rotations[player_id] = server_rotation

		if !initialized_players.has(player_id):
			initialized_players[player_id] = true
			player_node.position = server_position
			player_node.rotation = server_rotation


func _get_player_node(player_id):
	if player_nodes.has(player_id):
		return player_nodes[player_id]

	if player_id == self_id:
		player_nodes[player_id] = player
		return player

	var remote_player = PLAYER_SCENE.instantiate()
	add_child(remote_player)
	player_nodes[player_id] = remote_player

	return remote_player


func _remove_missing_players(server_players: Dictionary) -> void:
	for player_id in player_nodes.keys():
		if server_players.has(player_id):
			continue

		if player_nodes[player_id] != player:
			player_nodes[player_id].queue_free()

		player_nodes.erase(player_id)
		initialized_players.erase(player_id)
		target_player_positions.erase(player_id)
		target_player_rotations.erase(player_id)


func _interpolate_player(delta: float) -> void:
	var weight := 1.0 - exp(-player_interpolation_speed * delta)
	for player_id in player_nodes.keys():
		if !target_player_positions.has(player_id):
			continue

		var player_node = player_nodes[player_id]
		player_node.position = player_node.position.lerp(target_player_positions[player_id], weight)
		player_node.rotation = lerp_angle(player_node.rotation, target_player_rotations[player_id], weight)


func _update_layer_shader(background: TextureRect, parallax: float, offset: Vector2) -> void:
	var background_material := background.material as ShaderMaterial
	if background_material == null:
		return
	
	background_material.set_shader_parameter(
		"scroll_offset",
		(player.global_position * parallax) + offset
	)
