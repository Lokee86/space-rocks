extends RefCounted
class_name GameplayBackgroundFlow

const Constants := preload("res://scripts/generated/constants/constants.gd")

var repeated_background: TextureRect
var repeated_foreground_background: TextureRect
var repeated_planet_background: TextureRect
var parallax_target: Node2D
var camera: Camera2D

var background_drift_offset := Vector2.ZERO
var foreground_drift_offset := Vector2.ZERO
var planet_drift_offset := Vector2.ZERO
var last_valid_parallax_position := Vector2.ZERO


func configure(
	background: TextureRect,
	foreground_background: TextureRect,
	planet_background: TextureRect,
	parallax_target_ref: Node2D,
	camera_ref: Camera2D = null
) -> void:
	repeated_background = background
	repeated_foreground_background = foreground_background
	repeated_planet_background = planet_background
	parallax_target = parallax_target_ref
	camera = camera_ref
	if parallax_target != null:
		last_valid_parallax_position = parallax_target.global_position
	_apply_world_sampling(last_valid_parallax_position)


func set_parallax_target(parallax_target_ref: Node2D) -> void:
	parallax_target = parallax_target_ref


func process_frame() -> void:
	background_drift_offset += Constants.BACKGROUND_DRIFT_PER_FRAME
	foreground_drift_offset += Constants.FOREGROUND_BACKGROUND_DRIFT_PER_FRAME
	planet_drift_offset += Constants.PLANET_BACKGROUND_DRIFT_PER_FRAME

	var camera_world_position := last_valid_parallax_position
	if parallax_target != null:
		last_valid_parallax_position = parallax_target.global_position
		camera_world_position = last_valid_parallax_position

	_apply_world_sampling(camera_world_position)


func set_scroll_reference(scroll_position: Vector2) -> void:
	last_valid_parallax_position = scroll_position
	var camera_zoom := _camera_zoom()
	_set_world_sampling(
		repeated_background,
		scroll_position,
		camera_zoom,
		Constants.BACKGROUND_PARALLAX,
		Vector2.ZERO
	)
	_set_world_sampling(
		repeated_foreground_background,
		scroll_position,
		camera_zoom,
		Constants.FOREGROUND_BACKGROUND_PARALLAX,
		Constants.FOREGROUND_BACKGROUND_OFFSET
	)
	_set_world_sampling(
		repeated_planet_background,
		scroll_position,
		camera_zoom,
		Constants.PLANET_BACKGROUND_PARALLAX,
		Constants.PLANET_BACKGROUND_OFFSET
	)


func clear() -> void:
	background_drift_offset = Vector2.ZERO
	foreground_drift_offset = Vector2.ZERO
	planet_drift_offset = Vector2.ZERO
	last_valid_parallax_position = Vector2.ZERO
	_apply_world_sampling(Vector2.ZERO)


func _apply_world_sampling(camera_world_position: Vector2) -> void:
	var camera_zoom := _camera_zoom()
	_set_world_sampling(
		repeated_background,
		camera_world_position,
		camera_zoom,
		Constants.BACKGROUND_PARALLAX,
		background_drift_offset
	)
	_set_world_sampling(
		repeated_foreground_background,
		camera_world_position,
		camera_zoom,
		Constants.FOREGROUND_BACKGROUND_PARALLAX,
		foreground_drift_offset + Constants.FOREGROUND_BACKGROUND_OFFSET
	)
	_set_world_sampling(
		repeated_planet_background,
		camera_world_position,
		camera_zoom,
		Constants.PLANET_BACKGROUND_PARALLAX,
		planet_drift_offset + Constants.PLANET_BACKGROUND_OFFSET
	)


func _camera_zoom() -> float:
	if camera == null:
		return 1.0
	return max(camera.zoom.x, 0.001)


func _shader_material(texture_rect: TextureRect) -> ShaderMaterial:
	if texture_rect == null:
		return null
	return texture_rect.material as ShaderMaterial


func _set_world_sampling(
	texture_rect: TextureRect,
	camera_world_position: Vector2,
	camera_zoom: float,
	parallax_factor: float,
	layer_world_offset: Vector2
) -> void:
	var shader_material := _shader_material(texture_rect)
	if shader_material == null:
		return
	shader_material.set_shader_parameter("camera_world_position", camera_world_position)
	shader_material.set_shader_parameter("camera_zoom", camera_zoom)
	shader_material.set_shader_parameter("parallax_factor", parallax_factor)
	shader_material.set_shader_parameter("layer_world_offset", layer_world_offset)
