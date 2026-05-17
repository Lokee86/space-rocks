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
var target_player_positions := {}
var target_player_rotations := {}
var asteroid_spawn_timer := Timer.new()

func _ready() -> void:
	randomize()
	_setup_asteroid_spawner()

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
	_remove_missing_players(server_players)

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


func _physics_process(_delta: float) -> void:
	_move_local_asteroids()


func _setup_asteroid_spawner() -> void:
	asteroid_spawn_timer.wait_time = Constants.ASTEROID_SPAWN_INTERVAL
	asteroid_spawn_timer.one_shot = false
	asteroid_spawn_timer.timeout.connect(_spawn_local_asteroid)
	add_child(asteroid_spawn_timer)
	asteroid_spawn_timer.start()


func _spawn_local_asteroid() -> void:
	var asteroid := ASTEROID_SCENE.instantiate() as CharacterBody2D
	var asteroid_size := randi_range(1, 4)
	var variant_index := randi_range(0, 3)
	var spawn_position := _get_random_offscreen_position()
	var target_position: Vector2 = player.global_position
	var randomness_radians := deg_to_rad(Constants.ASTEROID_AIM_RANDOMNESS_DEGREES)
	var direction := spawn_position.direction_to(target_position).rotated(
		randf_range(-randomness_radians, randomness_radians)
	)
	var speed := randf_range(Constants.ASTEROID_MIN_SPEED, Constants.ASTEROID_MAX_SPEED)

	asteroid.global_position = spawn_position
	asteroid.scale = Vector2.ONE * float(asteroid_size) * Constants.ASTEROID_SIZE_SCALE
	asteroid.velocity = direction * speed
	asteroids.add_child(asteroid)
	asteroid.set_asteroid_variant(variant_index)


func _get_random_offscreen_position() -> Vector2:
	var screen_size := get_viewport_rect().size
	var edge := randi_range(0, 3)
	var screen_position := Vector2.ZERO

	match edge:
		0:
			screen_position = Vector2(randf_range(0.0, screen_size.x), -Constants.ASTEROID_SPAWN_MARGIN)
		1:
			screen_position = Vector2(
				screen_size.x + Constants.ASTEROID_SPAWN_MARGIN,
				randf_range(0.0, screen_size.y)
			)
		2:
			screen_position = Vector2(
				randf_range(0.0, screen_size.x),
				screen_size.y + Constants.ASTEROID_SPAWN_MARGIN
			)
		_:
			screen_position = Vector2(-Constants.ASTEROID_SPAWN_MARGIN, randf_range(0.0, screen_size.y))

	return get_viewport().get_canvas_transform().affine_inverse() * screen_position


func _move_local_asteroids() -> void:
	for asteroid in asteroids.get_children():
		if asteroid is CharacterBody2D:
			asteroid.move_and_slide()

		if _is_far_offscreen(asteroid.global_position):
			asteroid.queue_free()


func _is_far_offscreen(world_position: Vector2) -> bool:
	var screen_size := get_viewport_rect().size
	var screen_position := get_viewport().get_canvas_transform() * world_position

	return (
		screen_position.x < -Constants.ASTEROID_DESPAWN_MARGIN
		or screen_position.x > screen_size.x + Constants.ASTEROID_DESPAWN_MARGIN
		or screen_position.y < -Constants.ASTEROID_DESPAWN_MARGIN
		or screen_position.y > screen_size.y + Constants.ASTEROID_DESPAWN_MARGIN
	)


func _update_layer_shader(background: TextureRect, parallax: float, offset: Vector2) -> void:
	var background_material := background.material as ShaderMaterial
	if background_material == null:
		return
	
	background_material.set_shader_parameter(
		"scroll_offset",
		(player.global_position * parallax) + offset
	)
