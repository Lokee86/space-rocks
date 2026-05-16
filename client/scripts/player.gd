extends CharacterBody2D

@export var rotation_speed := 4.0
@export var thrust_force := 600.0
@export var max_speed := 700.0
@export var damping := 0.98

@export var rotate_left_action := &"ui_left"
@export var rotate_right_action := &"ui_right"
@export var thrust_forward_action := &"ui_up"
@export var thrust_reverse_action := &"ui_down"


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
	var input := Input.get_axis(rotate_left_action, rotate_right_action)

	if Input.is_key_pressed(KEY_A):
		input -= 1.0
	if Input.is_key_pressed(KEY_D):
		input += 1.0

	return clampf(input, -1.0, 1.0)


func _get_thrust_input() -> float:
	var input := Input.get_axis(thrust_reverse_action, thrust_forward_action)

	if Input.is_key_pressed(KEY_S):
		input -= 1.0
	if Input.is_key_pressed(KEY_W):
		input += 1.0

	return clampf(input, -1.0, 1.0)
