extends Node2D

const Constants = preload("res://scripts/constants.gd")
const Packets = preload("res://scripts/packets.gd")
const PLAYER_SCENE := preload("res://scenes/player.tscn")
const BULLET_SCENE := preload("res://scenes/bullet.tscn")
const ASTEROID_SCENE := preload("res://scenes/asteroid.tscn")
const BULLET_BLAST_SCENE := preload("res://scenes/bullet_blast.tscn")
const ASTEROID_Z_INDEX := 10
const BULLET_Z_INDEX := 20
const PLAYER_Z_INDEX := 30
const EFFECT_Z_INDEX := 40

@onready var player: Player = $Player
@onready var bullets = $Bullets
@onready var asteroids: Node2D = $Asteroids
@onready var repeated_background: TextureRect = $ParallaxBackground/BackgroundLayer/RepeatedBackground
@onready var repeated_foreground_background: TextureRect = $ParallaxBackground/ForegroundBackgroundLayer/RepeatedBackground

var socket := WebSocketPeer.new()
var connected := false
var has_received_state := false
var self_id := ""
var player_nodes := {}
var bullet_nodes := {}
var asteroid_nodes := {}
var initialized_players := {}
var initialized_bullets := {}
var initialized_asteroids := {}
var target_player_positions := {}
var target_player_rotations := {}
var target_bullet_positions := {}
var target_bullet_rotations := {}
var target_asteroid_positions := {}

func _ready() -> void:
	DisplayServer.window_set_min_size(Vector2i(1280, 720))
	asteroids.z_index = ASTEROID_Z_INDEX
	bullets.z_index = BULLET_Z_INDEX
	player.z_index = PLAYER_Z_INDEX
	get_viewport().size_changed.connect(_send_client_config)

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
			_send_client_config()

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
	if data.get(Packets.FIELD_TYPE, "") != Packets.TYPE_STATE:
		return

	self_id = data[Packets.FIELD_SELF_ID]
	var server_players: Dictionary = data[Packets.FIELD_PLAYERS]
	var server_bullets: Dictionary = data.get(Packets.FIELD_BULLETS, {})
	var server_asteroids: Dictionary = data.get(Packets.FIELD_ASTEROIDS, {})
	var server_events: Array = []
	var events_data = data.get(Packets.FIELD_EVENTS, [])
	if events_data is Array:
		server_events = events_data

	_remove_missing_players(server_players)
	_remove_missing_bullets(server_bullets)
	_remove_missing_asteroids(server_asteroids)
	_apply_bullets(server_bullets, has_received_state)
	_apply_asteroids(server_asteroids)
	_apply_events(server_events)
	has_received_state = true

	for player_id in server_players.keys():
		var state: Dictionary = server_players[player_id]
		var player_node = _get_player_node(player_id)
		var server_position := Vector2(state[Packets.FIELD_X], state[Packets.FIELD_Y])
		var server_rotation: float = state[Packets.FIELD_ROTATION]

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
		player.z_index = PLAYER_Z_INDEX
		player_nodes[player_id] = player
		return player

	var remote_player = PLAYER_SCENE.instantiate()
	remote_player.z_index = PLAYER_Z_INDEX
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

	for bullet_id in bullet_nodes.keys():
		if !target_bullet_positions.has(bullet_id):
			continue

		var bullet_node = bullet_nodes[bullet_id]
		bullet_node.global_position = bullet_node.global_position.lerp(
			target_bullet_positions[bullet_id],
			weight
		)
		bullet_node.rotation = lerp_angle(bullet_node.rotation, target_bullet_rotations[bullet_id], weight)

	for asteroid_id in asteroid_nodes.keys():
		if !target_asteroid_positions.has(asteroid_id):
			continue

		var asteroid_node = asteroid_nodes[asteroid_id]
		asteroid_node.global_position = asteroid_node.global_position.lerp(
			target_asteroid_positions[asteroid_id],
			weight
		)


func _apply_bullets(server_bullets: Dictionary, play_new_bullet_sounds: bool) -> void:
	for bullet_id in server_bullets.keys():
		var state: Dictionary = server_bullets[bullet_id]
		var is_new_bullet := !bullet_nodes.has(bullet_id)
		var bullet_node = _get_bullet_node(bullet_id)
		var server_position := Vector2(state[Packets.FIELD_X], state[Packets.FIELD_Y])
		var server_rotation: float = state[Packets.FIELD_ROTATION]

		target_bullet_positions[bullet_id] = server_position
		target_bullet_rotations[bullet_id] = server_rotation

		if !initialized_bullets.has(bullet_id):
			initialized_bullets[bullet_id] = true
			bullet_node.global_position = server_position
			bullet_node.rotation = server_rotation

		if is_new_bullet && play_new_bullet_sounds:
			player.play_laser_sound()


func _get_bullet_node(bullet_id):
	if bullet_nodes.has(bullet_id):
		return bullet_nodes[bullet_id]

	var bullet_node = BULLET_SCENE.instantiate()
	bullets.add_child(bullet_node)
	bullet_nodes[bullet_id] = bullet_node

	return bullet_node


func _remove_missing_bullets(server_bullets: Dictionary) -> void:
	for bullet_id in bullet_nodes.keys():
		if server_bullets.has(bullet_id):
			continue

		bullet_nodes[bullet_id].queue_free()
		bullet_nodes.erase(bullet_id)
		initialized_bullets.erase(bullet_id)
		target_bullet_positions.erase(bullet_id)
		target_bullet_rotations.erase(bullet_id)


func _apply_asteroids(server_asteroids: Dictionary) -> void:
	for asteroid_id in server_asteroids.keys():
		var state: Dictionary = server_asteroids[asteroid_id]
		var asteroid_node = _get_asteroid_node(asteroid_id)
		var server_position := Vector2(state[Packets.FIELD_X], state[Packets.FIELD_Y])

		target_asteroid_positions[asteroid_id] = server_position

		if !initialized_asteroids.has(asteroid_id):
			initialized_asteroids[asteroid_id] = true
			asteroid_node.global_position = server_position
			asteroid_node.scale = Vector2.ONE * float(state[Packets.FIELD_SIZE]) * Constants.ASTEROID_SIZE_SCALE
			asteroid_node.set_asteroid_variant(state[Packets.FIELD_VARIANT])


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


func _apply_events(server_events: Array) -> void:
	for event in server_events:
		if event.get(Packets.FIELD_TYPE, "") == Packets.TYPE_BULLET_BLAST:
			player.play_asteroid_destroyed_sound()
			_spawn_bullet_blast(Vector2(event[Packets.FIELD_X], event[Packets.FIELD_Y]))


func _spawn_bullet_blast(event_position: Vector2) -> void:
	var blast_node := BULLET_BLAST_SCENE.instantiate()
	blast_node.global_position = event_position
	blast_node.z_index = EFFECT_Z_INDEX
	add_child(blast_node)

	var sprite := blast_node.get_node_or_null("AnimatedSprite2D") as AnimatedSprite2D
	if sprite == null:
		blast_node.queue_free()
		return

	sprite.animation_finished.connect(blast_node.queue_free)
	sprite.play("bullet_blast")


func _update_layer_shader(background: TextureRect, parallax: float, offset: Vector2) -> void:
	var background_material := background.material as ShaderMaterial
	if background_material == null:
		return
	
	background_material.set_shader_parameter(
		"scroll_offset",
		(player.global_position * parallax) + offset
	)


func _send_client_config() -> void:
	if socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		return

	var visible_size := get_viewport_rect().size
	socket.send_text(JSON.stringify(Packets.client_config_packet(
		visible_size.x,
		visible_size.y
	)))
