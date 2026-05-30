extends RefCounted
class_name LocalVisualSync

const WorldWrapScript = preload("res://scripts/world/world_wrap.gd")

var local_server_position := Vector2.ZERO
var local_visual_position := Vector2.ZERO
var has_local_visual_position := false


func update_from_server_position(incoming_server_position: Vector2) -> void:
	var wrapped_server_position := WorldWrapScript.wrap_position(incoming_server_position)
	if has_local_visual_position:
		local_visual_position += WorldWrapScript.shortest_delta(local_server_position, wrapped_server_position)
		local_server_position = wrapped_server_position
		return

	local_server_position = wrapped_server_position
	local_visual_position = local_server_position
	has_local_visual_position = true


func server_position() -> Vector2:
	return local_server_position


func visual_position() -> Vector2:
	return local_visual_position


func is_initialized() -> bool:
	return has_local_visual_position


func visual_position_for_server_position(server_authoritive_position: Vector2) -> Vector2:
	if !has_local_visual_position:
		return server_authoritive_position

	return local_visual_position + WorldWrapScript.shortest_delta(
		local_server_position,
		server_authoritive_position
	)


func server_position_for_visual_position(visual_position: Vector2) -> Vector2:
	if !has_local_visual_position:
		return WorldWrapScript.wrap_position(visual_position)

	return WorldWrapScript.wrap_position(
		local_server_position + WorldWrapScript.shortest_delta(local_visual_position, visual_position)
	)
