extends CharacterBody2D

const Constants = preload("res://scripts/constants.gd")

@export var rotation_speed := Constants.PLAYER_ROTATION_SPEED
@export var thrust_force := Constants.PLAYER_THRUST_FORCE
@export var max_speed := Constants.PLAYER_MAX_SPEED
@export var damping := Constants.PLAYER_DAMPING

@export var turn_left_action := &"turn_left"
@export var turn_right_action := &"turn_right"
@export var move_forward_action := &"move_forward"
@export var move_backward_action := &"move_backward"


func _ready() -> void:
	position = get_viewport_rect().size * 0.5


func _physics_process(delta: float) -> void:
	var rotation_input := _get_rotation_input()
	rotation += rotation_input * rotation_speed * delta

	var thrust_input := _get_thrust_input()
	if thrust_input != 0.0:
		velocity += Vector2.UP.rotated(rotation) * thrust_force * thrust_input * delta

	velocity *= damping
	velocity = velocity.limit_length(max_speed)

	move_and_slide()


func _get_rotation_input() -> float:
	return Input.get_axis(turn_left_action, turn_right_action)


func _get_thrust_input() -> float:
	return Input.get_axis(move_backward_action, move_forward_action)
