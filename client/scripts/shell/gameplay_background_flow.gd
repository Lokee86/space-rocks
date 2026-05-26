class_name GameplayBackgroundFlow
extends RefCounted

const Constants := preload("res://scripts/constants/constants.gd")

var repeated_background: TextureRect
var repeated_foreground_background: TextureRect


func configure(background: TextureRect, foreground_background: TextureRect) -> void:
	repeated_background = background
	repeated_foreground_background = foreground_background


func set_scroll_reference(scroll_position: Vector2) -> void:
	_set_scroll_offset(repeated_background, scroll_position * Constants.BACKGROUND_PARALLAX)
	_set_scroll_offset(
		repeated_foreground_background,
		(scroll_position * Constants.FOREGROUND_BACKGROUND_PARALLAX) + Constants.FOREGROUND_BACKGROUND_OFFSET
	)


func clear() -> void:
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
