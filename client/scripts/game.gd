extends Node2D

const Constants = preload("res://scripts/constants.gd")
const PLAYER_SCENE := preload("res://scenes/player.tscn")
const BULLET_SCENE := preload("res://scenes/bullet.tscn")
const ASTEROID_SCENE := preload("res://scenes/asteroid.tscn")

@onready var player = $Player
@onready var bullets = $Bullets
@onready var asteroids: Node2D = $Asteroids
@onready var repeated_background: TextureRect = $ParallaxBackground/BackgroundLayer/RepeatedBackground
@onready var repeated_foreground_background: TextureRect = $ParallaxBackground/ForegroundBackgroundLayer/RepeatedBackground

var socket := WebSocketPeer.new()
var connected := false
var self_id := ""
var player_nodes := {}
var bullet_nodes := {}
var asteroid_nodes := {}
var initialized_players := {}
var initialized_asteroids := {}
var target_player_positions := {}
var target_player_rotations := {}
var target_asteroid_positions := {}

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
	_update_layer_shader(repeated_background, Constants.BACKGROUND_PARALLAX, Vector2.ZERO)
	_update_layer_shader(
		repeated_foreground_background,
		Constants.FOREGROUND_BACKGROUND_PARALLAX,
		Constants.FOREGROUND_BACKGROUND_OFFSET
	)


func _apply_state(data: Dictionary) -> void:
	if data.get("type", "") != "state":
		return

	self_id = data["self_id"]
	var server_players: Dictionary = data["players"]
	var server_asteroids: Dictionary = data.get("asteroids", {})

	_remove_missing_players(server_players)
	_remove_missing_asteroids(server_asteroids)
	_apply_asteroids(server_asteroids)

	for player_id in server_players.keys():
		var state: Dictionary = server_players[player_id]
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
	var weight := 1.0 - exp(-Constants.PLAYER_INTERPOLATION_SPEED * delta)
	for player_id in player_nodes.keys():
		if !target_player_positions.has(player_id):
			continue

		var player_node = player_nodes[player_id]
		player_node.position = player_node.position.lerp(target_player_positions[player_id], weight)
		player_node.rotation = lerp_angle(player_node.rotation, target_player_rotations[player_id], weight)

	for asteroid_id in asteroid_nodes.keys():
		if !target_asteroid_positions.has(asteroid_id):
			continue

		var asteroid_node = asteroid_nodes[asteroid_id]
		asteroid_node.global_position = asteroid_node.global_position.lerp(
			target_asteroid_positions[asteroid_id],
			weight
		)


func _apply_asteroids(server_asteroids: Dictionary) -> void:
	for asteroid_id in server_asteroids.keys():
		var state: Dictionary = server_asteroids[asteroid_id]
		var asteroid_node = _get_asteroid_node(asteroid_id)
		var server_position := Vector2(state["x"], state["y"])

		target_asteroid_positions[asteroid_id] = server_position

		if !initialized_asteroids.has(asteroid_id):
			initialized_asteroids[asteroid_id] = true
			asteroid_node.global_position = server_position
			asteroid_node.scale = Vector2.ONE * float(state["size"]) * Constants.ASTEROID_SIZE_SCALE
			asteroid_node.set_asteroid_variant(state["variant"])


func _get_asteroid_node(asteroid_id):
	if asteroid_nodes.has(asteroid_id):
		return asteroid_nodes[asteroid_id]

	var asteroid_node = ASTEROID_SCENE.instantiate()
	asteroids.add_child(asteroid_node)
	asteroid_nodes[asteroid_id] = asteroid_node

	return asteroid_node


func _remove_missing_asteroids(server_asteroids: Dictionary) -> void:
	for asteroid_id in asteroid_nodes.keys():
		if server_asteroids.has(asteroid_id):
			continue

		asteroid_nodes[asteroid_id].queue_free()
		asteroid_nodes.erase(asteroid_id)
		initialized_asteroids.erase(asteroid_id)
		target_asteroid_positions.erase(asteroid_id)


func _update_layer_shader(background: TextureRect, parallax: float, offset: Vector2) -> void:
	var background_material := background.material as ShaderMaterial
	if background_material == null:
		return
	
	background_material.set_shader_parameter(
		"scroll_offset",
		(player.global_position * parallax) + offset
	)
