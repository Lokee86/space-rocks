class_name GameplayBackgroundFlow
extends RefCounted

const Constants := preload("res://scripts/constants/constants.gd")

var repeated_background: TextureRect
var repeated_foreground_background: TextureRect
var player: Player
var has_received_gameplay_state := false


func configure(background: TextureRect, foreground_background: TextureRect, player_ref: Player) -> void:
	repeated_background = background
	repeated_foreground_background = foreground_background
	player = player_ref


func mark_gameplay_state_received() -> void:
	has_received_gameplay_state = true


func process() -> void:
	if !has_received_gameplay_state:
		return
	if player == null:
		return
	if !player.visible:
		return
	set_scroll_reference(player.global_position)


func set_scroll_reference(scroll_position: Vector2) -> void:
	_set_scroll_offset(repeated_background, scroll_position * Constants.BACKGROUND_PARALLAX)
	_set_scroll_offset(
		repeated_foreground_background,
		(scroll_position * Constants.FOREGROUND_BACKGROUND_PARALLAX) + Constants.FOREGROUND_BACKGROUND_OFFSET
	)


func clear() -> void:
	has_received_gameplay_state = false
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
