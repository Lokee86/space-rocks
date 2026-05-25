extends RefCounted

const Constants = preload("res://scripts/constants/constants.gd")
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
