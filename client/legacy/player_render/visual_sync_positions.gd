extends RefCounted

const Constants = preload("res://scripts/generated/constants/constants.gd")
const WorldWrapScript = preload("res://scripts/world/world_wrap.gd")


static func relative_to_local_visual(
	local_visual_position: Vector2,
	local_server_position: Vector2,
	server_position: Vector2
) -> Vector2:
	return local_visual_position + WorldWrapScript.shortest_delta(
		local_server_position,
		server_position
	)


static func is_world_copy_mismatch(current_position: Vector2, target_position: Vector2) -> bool:
	var delta := target_position - current_position
	return (
		abs(delta.x) > Constants.WORLD_WIDTH * 0.5 ||
		abs(delta.y) > Constants.WORLD_HEIGHT * 0.5
	)

