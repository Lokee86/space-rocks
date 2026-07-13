extends CharacterBody2D
class_name BulletPresentation

const Constants = preload("res://scripts/generated/constants/constants.gd")

@onready var sprite: Sprite2D = $Sprite2D
var base_sprite_scale: Vector2
var pulse_sprite_scale: Vector2
var base_modulate: Color
var pulse_modulate := Color(1.0, 1.0, 1.0, 0.55)
var pulse_tween: Tween


func _ready() -> void:
	base_sprite_scale = sprite.scale
	pulse_sprite_scale = base_sprite_scale * Constants.BULLET_PULSE_MULTIPLIER
	base_modulate = sprite.modulate

	reset_from_pool()


func _start_pulse() -> void:
	pulse_tween = create_tween()
	pulse_tween.set_loops()
	pulse_tween.set_trans(Tween.TRANS_SINE)
	pulse_tween.set_ease(Tween.EASE_IN_OUT)

	pulse_tween.tween_property(sprite, "scale", pulse_sprite_scale, Constants.BULLET_PULSE_TIME)
	pulse_tween.parallel().tween_property(sprite, "modulate", pulse_modulate, Constants.BULLET_PULSE_TIME)

	pulse_tween.tween_property(sprite, "scale", base_sprite_scale, Constants.BULLET_PULSE_TIME)
	pulse_tween.parallel().tween_property(sprite, "modulate", base_modulate, Constants.BULLET_PULSE_TIME)


func reset_from_pool() -> void:
	if pulse_tween != null:
		pulse_tween.kill()
	modulate = Color.WHITE
	rotation = 0.0
	scale = Vector2.ONE
	sprite.scale = base_sprite_scale
	sprite.modulate = base_modulate
	visible = false
	_start_pulse()


func reset_for_pool() -> void:
	if pulse_tween != null:
		pulse_tween.kill()
	sprite.scale = base_sprite_scale
	sprite.modulate = base_modulate
	visible = false

