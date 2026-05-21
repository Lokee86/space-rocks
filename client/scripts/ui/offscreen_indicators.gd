extends Control

const SCREEN_MARGIN: float = 32.0

@export var indicator_scene: PackedScene

var indicators: Dictionary = {}


func update_indicators(remote_visual_positions: Dictionary, camera: Camera2D) -> void:
	_remove_stale_indicators(remote_visual_positions)
	if remote_visual_positions.is_empty():
		return
	if camera == null:
		_hide_all_indicators()
		return
	if indicator_scene == null:
		_hide_all_indicators()
		return

	var screen_size: Vector2 = get_viewport_rect().size
	var screen_center: Vector2 = screen_size * 0.5

	for player_id in remote_visual_positions.keys():
		var indicator: Control = _get_or_create_indicator(str(player_id))
		if indicator != null:
			indicator.call("hide_indicator")
			var screen_position: Vector2 = camera.get_canvas_transform() * remote_visual_positions[player_id]
			if _is_inside_screen(screen_position, screen_size, SCREEN_MARGIN):
				indicator.call("hide_indicator")
				continue

			var direction: Vector2 = (screen_position - screen_center).normalized()
			var edge_position: Vector2 = edge_position_from_direction(direction, screen_size, SCREEN_MARGIN)
			indicator.call("set_indicator", edge_position, direction)


func _get_or_create_indicator(player_id: String) -> Control:
	if indicators.has(player_id):
		return indicators[player_id] as Control

	if indicator_scene == null:
		return null

	var indicator: Control = indicator_scene.instantiate() as Control
	add_child(indicator)
	indicators[player_id] = indicator
	return indicator


func _remove_stale_indicators(remote_visual_positions: Dictionary) -> void:
	for player_id in indicators.keys():
		if remote_visual_positions.has(player_id):
			continue

		var indicator: Control = indicators[player_id] as Control
		if indicator != null:
			indicator.queue_free()
		indicators.erase(player_id)


func _hide_all_indicators() -> void:
	for indicator in indicators.values():
		if indicator != null && indicator.has_method("hide_indicator"):
			indicator.call("hide_indicator")


func _is_inside_screen(screen_position: Vector2, screen_size: Vector2, margin: float) -> bool:
	return (
		screen_position.x >= margin &&
		screen_position.x <= screen_size.x - margin &&
		screen_position.y >= margin &&
		screen_position.y <= screen_size.y - margin
	)


func edge_position_from_direction(direction: Vector2, screen_size: Vector2, margin: float) -> Vector2:
	var center: Vector2 = screen_size * 0.5
	var half: Vector2 = center - Vector2(margin, margin)
	var scale_x: float = INF
	var scale_y: float = INF

	if abs(direction.x) > 0.001:
		scale_x = half.x / abs(direction.x)
	if abs(direction.y) > 0.001:
		scale_y = half.y / abs(direction.y)

	var scale: float = min(scale_x, scale_y)
	return center + direction * scale
