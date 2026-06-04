extends CharacterBody2D

const Constants = preload("res://scripts/constants/constants.gd")

@onready var sprite: Sprite2D = $Sprite2D
@onready var laser_sound: AudioStreamPlayer2D = $LaserSound

var base_sprite_scale: Vector2
var pulse_sprite_scale: Vector2
var base_modulate: Color
var pulse_modulate := Color(1.0, 1.0, 1.0, 0.55)
var audio_flow := GameplayAudioFlow.new()


func _ready() -> void:
	base_sprite_scale = sprite.scale
	pulse_sprite_scale = base_sprite_scale * Constants.BULLET_PULSE_MULTIPLIER
	base_modulate = sprite.modulate

	_start_pulse()
	_play_laser_sound()


func _start_pulse() -> void:
	var tween := create_tween()
	tween.set_loops()
	tween.set_trans(Tween.TRANS_SINE)
	tween.set_ease(Tween.EASE_IN_OUT)

	tween.tween_property(sprite, "scale", pulse_sprite_scale, Constants.BULLET_PULSE_TIME)
	tween.parallel().tween_property(sprite, "modulate", pulse_modulate, Constants.BULLET_PULSE_TIME)

	tween.tween_property(sprite, "scale", base_sprite_scale, Constants.BULLET_PULSE_TIME)
	tween.parallel().tween_property(sprite, "modulate", base_modulate, Constants.BULLET_PULSE_TIME)


func _play_laser_sound() -> void:
	audio_flow.play_laser_sound(laser_sound)
