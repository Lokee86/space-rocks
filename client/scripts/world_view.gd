extends RefCounted
class_name WorldView


static func visual_position_relative_to_local(
	local_server_position: Vector2,
	local_visual_position: Vector2,
	target_server_position: Vector2
) -> Vector2:
	return local_visual_position + (target_server_position - local_server_position)
