class_name GameplayBackgroundFlow
extends RefCounted

const Constants := preload("res://scripts/constants/constants.gd")

const BACKGROUND_DRIFT_PER_FRAME := Vector2(0.08, 0.025)
const FOREGROUND_DRIFT_PER_FRAME := Vector2(0.12, 0.04)

var repeated_background: TextureRect
var repeated_foreground_background: TextureRect
var parallax_target: Node2D
var background_drift_offset := Vector2.ZERO
var foreground_drift_offset := Vector2.ZERO


func configure(
	background: TextureRect,
	foreground_background: TextureRect,
	parallax_target_ref: Node2D
) -> void:
	repeated_background = background
	repeated_foreground_background = foreground_background
	parallax_target = parallax_target_ref


func set_parallax_target(parallax_target_ref: Node2D) -> void:
	parallax_target = parallax_target_ref


func process_frame() -> void:
	background_drift_offset += BACKGROUND_DRIFT_PER_FRAME
	foreground_drift_offset += FOREGROUND_DRIFT_PER_FRAME

	var scroll_position := Vector2.ZERO
	if parallax_target != null && parallax_target.visible:
		scroll_position = parallax_target.global_position

	_set_scroll_offset(
		repeated_background,
		background_drift_offset + (scroll_position * Constants.BACKGROUND_PARALLAX)
	)
	_set_scroll_offset(
		repeated_foreground_background,
		foreground_drift_offset
			+ (scroll_position * Constants.FOREGROUND_BACKGROUND_PARALLAX)
			+ Constants.FOREGROUND_BACKGROUND_OFFSET
	)


func set_scroll_reference(scroll_position: Vector2) -> void:
	_set_scroll_offset(repeated_background, scroll_position * Constants.BACKGROUND_PARALLAX)
	_set_scroll_offset(
		repeated_foreground_background,
		(scroll_position * Constants.FOREGROUND_BACKGROUND_PARALLAX) + Constants.FOREGROUND_BACKGROUND_OFFSET
	)


func clear() -> void:
	background_drift_offset = Vector2.ZERO
	foreground_drift_offset = Vector2.ZERO
	_set_scroll_offset(repeated_background, Vector2.ZERO)
	_set_scroll_offset(repeated_foreground_background, Constants.FOREGROUND_BACKGROUND_OFFSET)


func _shader_material(texture_rect: TextureRect) -> ShaderMaterial:
	if texture_rect == null:
		return null
	return texture_rect.material as ShaderMaterial


func _set_scroll_offset(texture_rect: TextureRect, scroll_offset: Vector2) -> void:
	var shader_material := _shader_material(texture_rect)
	if shader_material == null:
		return
	shader_material.set_shader_parameter("scroll_offset", scroll_offset)
