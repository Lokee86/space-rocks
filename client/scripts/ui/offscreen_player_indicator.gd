extends Control


func set_indicator(screen_position: Vector2, direction: Vector2) -> void:
	position = screen_position
	rotation = direction.angle() + PI / 2.0
	visible = true


func hide_indicator() -> void:
	visible = false
