extends Node2D

const Constants = preload("res://scripts/constants.gd")
const Packets = preload("res://scripts/packets.gd")
const PLAYER_SCENE := preload("res://scenes/player.tscn")
const BULLET_SCENE := preload("res://scenes/bullet.tscn")
const ASTEROID_SCENE := preload("res://scenes/asteroid.tscn")
const BULLET_BLAST_SCENE := preload("res://scenes/animations/bullet_blast.tscn")
const SHIP_DEATH_SCENE := preload("res://scenes/animations/ship_death.tscn")
const ASTEROID_Z_INDEX := 10
const BULLET_Z_INDEX := 20
const PLAYER_Z_INDEX := 30
const EFFECT_Z_INDEX := 40

@onready var player: Player = $Player
@onready var bullets = $Bullets
@onready var asteroids: Node2D = $Asteroids
@onready var repeated_background: TextureRect = $ParallaxBackground/BackgroundLayer/RepeatedBackground
@onready var repeated_foreground_background: TextureRect = $ParallaxBackground/ForegroundBackgroundLayer/RepeatedBackground

var score_label: Label
var lives_label: Label
var death_overlay: Control
var game_over_overlay: Control
var game_over_sound: AudioStreamPlayer
var respawn_timer_label: Label
var respawn_tell_label: Label
var respawn_timer_template := "Respawn in X"
var is_dead := false
var can_respawn := false
var respawn_requested := false
var game_over_sound_played := false
var respawn_countdown_remaining := 0.0
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
	score_label = _find_score_label()
	lives_label = _find_label("LivesCount")
	death_overlay = _find_control("DeathOverlay")
	game_over_overlay = _find_control("GameOverOverlay")
	game_over_sound = _find_game_over_sound()
	respawn_timer_label = _find_label("RespawnTimer")
	respawn_tell_label = _find_label("RespawnTell")
	if respawn_timer_label != null:
		respawn_timer_template = respawn_timer_label.text
	_set_score(0)
	_set_lives(Constants.PLAYER_STARTING_LIVES)
	_set_alive_hud()

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
	_update_respawn_countdown(delta)

	var state := socket.get_ready_state()

	if state == WebSocketPeer.STATE_OPEN:
		if !connected:
			connected = true
			print("Connected!")
			_send_client_config()

		socket.send_text(JSON.stringify(player.get_input_packet()))
		if can_respawn && !respawn_requested && Input.is_key_pressed(KEY_R):
			respawn_requested = true
			socket.send_text(JSON.stringify(Packets.respawn_packet()))
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

	_set_lives(int(data.get(Packets.FIELD_LIVES, Constants.PLAYER_STARTING_LIVES)))
	if server_players.has(self_id):
		_set_score(int(server_players[self_id].get(Packets.FIELD_SCORE, 0)))
		if is_dead && respawn_requested:
			_set_alive_hud()

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
		player.visible = true
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

		_remove_player_node(player_id)


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
			_spawn_bullet_blast(Vector2(event[Packets.FIELD_X], event[Packets.FIELD_Y]))
		elif event.get(Packets.FIELD_TYPE, "") == Packets.TYPE_SHIP_DEATH:
			if event[Packets.FIELD_PLAYER_ID] == self_id:
				_apply_self_death_event(event)
			_spawn_ship_death(Vector2(event[Packets.FIELD_X], event[Packets.FIELD_Y]))


func _remove_player_node(player_id: String) -> void:
	if !player_nodes.has(player_id):
		if player_id == self_id:
			player.visible = false
		return

	var player_node = player_nodes[player_id]
	if player_node == player:
		player.visible = false
	else:
		player_node.queue_free()

	player_nodes.erase(player_id)
	initialized_players.erase(player_id)
	target_player_positions.erase(player_id)
	target_player_rotations.erase(player_id)


func _spawn_bullet_blast(event_position: Vector2) -> void:
	var blast_node := BULLET_BLAST_SCENE.instantiate()
	blast_node.global_position = event_position
	blast_node.z_index = EFFECT_Z_INDEX
	add_child(blast_node)

	var sprite := blast_node.get_node_or_null("AnimatedSprite2D") as AnimatedSprite2D
	var sound := blast_node.get_node_or_null("AsteroidDestroyed") as AudioStreamPlayer2D
	if sprite == null || sound == null:
		blast_node.queue_free()
		return

	var free_blast := func() -> void:
		if is_instance_valid(blast_node):
			blast_node.queue_free()

	sprite.animation_finished.connect(func() -> void:
		sprite.visible = false
	)
	sound.finished.connect(free_blast)

	sprite.play("bullet_blast")
	sound.play()

	var sound_length := 1.0
	if sound.stream != null:
		sound_length = max(sound.stream.get_length(), sound_length)
	get_tree().create_timer(sound_length + 0.25).timeout.connect(free_blast)


func _spawn_ship_death(event_position: Vector2) -> void:
	var death_node := SHIP_DEATH_SCENE.instantiate()
	death_node.global_position = event_position
	death_node.z_index = EFFECT_Z_INDEX
	add_child(death_node)

	var sprite := death_node.get_node_or_null("AnimatedSprite2D") as AnimatedSprite2D
	var sound := death_node.get_node_or_null("ShipDeath") as AudioStreamPlayer2D
	if sprite == null || sound == null:
		death_node.queue_free()
		return

	var death_finished := false
	var free_death := func() -> void:
		if death_finished:
			return
		death_finished = true
		if is_instance_valid(death_node):
			death_node.queue_free()

	sprite.animation_finished.connect(func() -> void:
		sprite.visible = false
	)
	sound.finished.connect(free_death)

	sprite.frame = 0
	sprite.frame_progress = 0.0
	sprite.play("default")
	sound.play()

	var sound_length := 0.0
	if sound.stream != null:
		sound_length = sound.stream.get_length()
	if sound_length > 0:
		get_tree().create_timer(sound_length + 0.05).timeout.connect(free_death)


func _set_score(score: int) -> void:
	if score_label == null:
		return

	score_label.text = "SCORE: %d" % score


func _set_lives(lives: int) -> void:
	if lives_label == null:
		return

	lives_label.text = "%d x " % lives


func _apply_self_death_event(event: Dictionary) -> void:
	var lives := int(event.get(Packets.FIELD_LIVES, 0))
	_set_lives(lives)
	if lives <= 0:
		_set_game_over_hud()
		return

	_set_dead_hud(float(event.get(Packets.FIELD_RESPAWN_DELAY, Constants.PLAYER_RESPAWN_DELAY)))


func _set_alive_hud() -> void:
	is_dead = false
	can_respawn = false
	respawn_requested = false
	game_over_sound_played = false
	respawn_countdown_remaining = 0.0
	if death_overlay != null:
		death_overlay.visible = false
	if game_over_overlay != null:
		game_over_overlay.visible = false
	if game_over_sound != null:
		game_over_sound.stop()


func _set_dead_hud(respawn_delay: float) -> void:
	is_dead = true
	can_respawn = false
	respawn_requested = false
	respawn_countdown_remaining = respawn_delay
	if death_overlay != null:
		death_overlay.visible = true
	if game_over_overlay != null:
		game_over_overlay.visible = false
	if game_over_sound != null:
		game_over_sound.stop()
	if respawn_timer_label != null:
		respawn_timer_label.visible = true
	if respawn_tell_label != null:
		respawn_tell_label.visible = false
	_update_respawn_timer_label()


func _set_game_over_hud() -> void:
	is_dead = true
	can_respawn = false
	respawn_requested = false
	respawn_countdown_remaining = 0.0
	if death_overlay != null:
		death_overlay.visible = false
	if game_over_overlay != null:
		game_over_overlay.visible = true
	_play_game_over_sound_after_delay()


func _play_game_over_sound_after_delay() -> void:
	if Constants.GAME_OVER_SOUND_DELAY <= 0:
		_play_game_over_sound()
		return

	get_tree().create_timer(Constants.GAME_OVER_SOUND_DELAY).timeout.connect(func() -> void:
		if is_dead && game_over_overlay != null && game_over_overlay.visible:
			_play_game_over_sound()
	)


func _play_game_over_sound() -> void:
	if game_over_sound != null && !game_over_sound_played:
		game_over_sound_played = true
		game_over_sound.play()


func _update_respawn_countdown(delta: float) -> void:
	if !is_dead || can_respawn || respawn_countdown_remaining <= 0:
		return

	respawn_countdown_remaining = max(0.0, respawn_countdown_remaining - delta)
	_update_respawn_timer_label()
	if respawn_countdown_remaining == 0:
		can_respawn = true
		if respawn_timer_label != null:
			respawn_timer_label.visible = false
		if respawn_tell_label != null:
			respawn_tell_label.visible = true


func _update_respawn_timer_label() -> void:
	if respawn_timer_label == null:
		return

	var seconds_remaining := int(ceil(respawn_countdown_remaining))
	respawn_timer_label.text = respawn_timer_template.replace("X", str(seconds_remaining))


func _find_score_label() -> Label:
	return _find_label("Score")


func _find_label(node_name: String) -> Label:
	var scene := get_tree().current_scene
	if scene != null:
		var label := scene.find_child(node_name, true, false) as Label
		if label != null:
			return label

	return null


func _find_control(node_name: String) -> Control:
	var scene := get_tree().current_scene
	if scene != null:
		var control := scene.find_child(node_name, true, false) as Control
		if control != null:
			return control

	return null


func _find_game_over_sound() -> AudioStreamPlayer:
	if game_over_overlay == null:
		return null

	return game_over_overlay.find_child("GameOverSound", true, false) as AudioStreamPlayer


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
