extends RefCounted

const Constants = preload("res://scripts/generated/constants/constants.gd")
const WORLD_SIZE := Vector2(Constants.WORLD_WIDTH, Constants.WORLD_HEIGHT)


static func wrap_position(pos: Vector2) -> Vector2:
	return Vector2(
		_wrap_coordinate(pos.x, WORLD_SIZE.x),
		_wrap_coordinate(pos.y, WORLD_SIZE.y)
	)


static func shortest_delta(from: Vector2, to: Vector2) -> Vector2:
	return Vector2(
		_shortest_coordinate_delta(to.x - from.x, WORLD_SIZE.x),
		_shortest_coordinate_delta(to.y - from.y, WORLD_SIZE.y)
	)


static func visual_position_relative_to(reference_position: Vector2, target: Vector2) -> Vector2:
	return reference_position + shortest_delta(reference_position, target)


static func visual_copy_offset_to_anchor(
	current_visual_position: Vector2,
	anchor_visual_position: Vector2,
	anchor_server_position: Vector2,
	entity_server_position: Vector2
) -> Vector2:
	var desired_visual_position := anchor_visual_position + shortest_delta(
		anchor_server_position,
		entity_server_position
	)
	var desired_offset := desired_visual_position - current_visual_position
	return Vector2(
		_world_copy_axis_offset(desired_offset.x, WORLD_SIZE.x),
		_world_copy_axis_offset(desired_offset.y, WORLD_SIZE.y)
	)


static func _wrap_coordinate(value: float, size: float) -> float:
	if size <= 0.0:
		return value

	var wrapped := fmod(value, size)
	if wrapped < 0.0:
		wrapped += size
	return wrapped


static func _shortest_coordinate_delta(delta: float, size: float) -> float:
	if size <= 0.0:
		return delta

	var half_size := size * 0.5
	if delta > half_size:
		return delta - size
	if delta < -half_size:
		return delta + size
	return delta


static func _world_copy_axis_offset(desired_offset: float, size: float) -> float:
	if size <= 0.0 || abs(desired_offset) <= size * 0.5:
		return 0.0
	return round(desired_offset / size) * size

