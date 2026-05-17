extends Node2D

@export var background_parallax := 0.25
@export var foreground_background_parallax := 0.45
@export var foreground_background_offset := Vector2(480.0, 270.0)

@onready var player: Node2D = $Player
@onready var repeated_background: TextureRect = $ParallaxBackground/BackgroundLayer/RepeatedBackground
@onready var repeated_foreground_background: TextureRect = $ParallaxBackground/ForegroundBackgroundLayer/RepeatedBackground

var socket := WebSocketPeer.new()
var connected := false
var packet := {
  "type": "input",
  "input": {
    "forward": true,
	"back": false,
	"right": true,
	"left": false,
	"shoot": false,
  }
}

func _ready() -> void:
	var err := socket.connect_to_url("ws://localhost:8080/ws")
	if err != OK:
		print("connection failede")
	else:
		print("Connecting...")


func _process(_delta: float) -> void:
	_update_layer_shader(repeated_background, background_parallax, Vector2.ZERO)
	_update_layer_shader(repeated_foreground_background, foreground_background_parallax, foreground_background_offset)

	socket.poll()

	var state := socket.get_ready_state()

	if state == WebSocketPeer.STATE_OPEN and !connected:
		connected = true
		print("Connected!")
		socket.send_text(JSON.stringify(packet))
	elif state == WebSocketPeer.STATE_CLOSED:
		print("Closed")

	while socket.get_available_packet_count() > 0:
		var text := socket.get_packet().get_string_from_utf8()
		var data = JSON.parse_string(text)

		if data == null:
			print("bad json: ", text)
			return

		print(data)
		player.position = Vector2(data["x"], data["y"])


func _update_layer_shader(background: TextureRect, parallax: float, offset: Vector2) -> void:
	var background_material := background.material as ShaderMaterial
	if background_material == null:
		return
	
	background_material.set_shader_parameter(
		"scroll_offset",
		(player.global_position * parallax) + offset
	)
