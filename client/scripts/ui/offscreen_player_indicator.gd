extends Control


func set_indicator(screen_position: Vector2, direction: Vector2) -> void:
	var center_offset: Vector2 = pivot_offset
	if center_offset == Vector2.ZERO:
		center_offset = size * 0.5
	position = screen_position - center_offset
	rotation = direction.angle() + PI / 2.0
	visible = true


func hide_indicator() -> void:
	visible = false
