extends CharacterBody2D
class_name Player

@export var turn_left_action := &"turn_left"
@export var turn_right_action := &"turn_right"
@export var move_forward_action := &"move_forward"
@export var move_backward_action := &"move_backward"
@export var shoot_action := &"shoot"

@onready var laser_sound: AudioStreamPlayer2D = $LaserSound
@onready var asteroid_destroyed_sound: AudioStreamPlayer2D = $AsteroidDestroyed


func get_input_packet() -> Dictionary:
	return {
		"type": "input",
		"input": {
			"forward": Input.is_action_pressed(move_forward_action),
			"back": Input.is_action_pressed(move_backward_action),
			"right": Input.is_action_pressed(turn_right_action),
			"left": Input.is_action_pressed(turn_left_action),
			"shoot": Input.is_action_pressed(shoot_action),
		}
	}


func play_laser_sound() -> void:
	laser_sound.play()


func play_asteroid_destroyed_sound() -> void:
	asteroid_destroyed_sound.play()
