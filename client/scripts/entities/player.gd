extends CharacterBody2D
class_name Player

const Packets = preload("res://scripts/networking/packets.gd")
const AFTERBURNER_SCENE := preload("res://scenes/animations/blue_afterburner.tscn")

@export var turn_left_action := &"turn_left"
@export var turn_right_action := &"turn_right"
@export var move_forward_action := &"move_forward"
@export var move_backward_action := &"move_backward"
@export var shoot_action := &"shoot"

@onready var laser_sound: AudioStreamPlayer2D = $LaserSound
@onready var afterburner_marker: Marker2D = $AfterburnerMarker

var afterburner: Node2D
var afterburner_sprite: AnimatedSprite2D
var afterburner_active := false


func _ready() -> void:
	afterburner = AFTERBURNER_SCENE.instantiate()
	afterburner.rotation_degrees = -90.0
	afterburner.scale = Vector2(0.45, 0.45)
	afterburner.visible = false
	afterburner_marker.add_child(afterburner)

	afterburner_sprite = afterburner.get_node_or_null("AnimatedSprite2D") as AnimatedSprite2D


func get_input_packet() -> Dictionary:
	return Packets.input_packet(
		Input.is_action_pressed(move_forward_action),
		Input.is_action_pressed(move_backward_action),
		Input.is_action_pressed(turn_right_action),
		Input.is_action_pressed(turn_left_action),
		Input.is_action_pressed(shoot_action)
	)


func play_laser_sound() -> void:
	laser_sound.play()


func set_afterburner_active(active: bool) -> void:
	if afterburner == null || afterburner_active == active:
		return

	afterburner_active = active
	afterburner.visible = active
	if afterburner_sprite == null:
		return

	if active:
		afterburner_sprite.play("default")
	else:
		afterburner_sprite.stop()
