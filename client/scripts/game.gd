extends Node2D

@export var background_parallax := 0.25
@export var foreground_background_parallax := 0.45
@export var foreground_background_offset := Vector2(480.0, 270.0)
@export var player_interpolation_speed := 18.0

@onready var player = $Player
@onready var repeated_background: TextureRect = $ParallaxBackground/BackgroundLayer/RepeatedBackground
@onready var repeated_foreground_background: TextureRect = $ParallaxBackground/ForegroundBackgroundLayer/RepeatedBackground

var socket := WebSocketPeer.new()
var connected := false
var has_server_state := false
var target_player_position := Vector2.ZERO
var target_player_rotation := 0.0

func _ready() -> void:
	target_player_position = player.position
	target_player_rotation = player.rotation

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

		_set_player_target(Vector2(data["x"], data["y"]), data["rotation"])

	_interpolate_player(delta)
	_update_layer_shader(repeated_background, background_parallax, Vector2.ZERO)
	_update_layer_shader(repeated_foreground_background, foreground_background_parallax, foreground_background_offset)


func _set_player_target(server_position: Vector2, server_rotation: float) -> void:
	target_player_position = server_position
	target_player_rotation = server_rotation

	if !has_server_state:
		has_server_state = true
		player.position = target_player_position
		player.rotation = target_player_rotation


func _interpolate_player(delta: float) -> void:
	if !has_server_state:
		return

	var weight := 1.0 - exp(-player_interpolation_speed * delta)
	player.position = player.position.lerp(target_player_position, weight)
	player.rotation = lerp_angle(player.rotation, target_player_rotation, weight)


func _update_layer_shader(background: TextureRect, parallax: float, offset: Vector2) -> void:
	var background_material := background.material as ShaderMaterial
	if background_material == null:
		return
	
	background_material.set_shader_parameter(
		"scroll_offset",
		(player.global_position * parallax) + offset
	)
