extends Node2D

var hitbox_entries: Array = []


func _ready() -> void:
	z_index = 1000


func set_enabled(enabled: bool) -> void:
	visible = enabled
	if !enabled:
		hitbox_entries.clear()
	queue_redraw()


func is_enabled() -> bool:
	return visible


func set_hitbox_entries(next_entries: Array) -> void:
	if !visible:
		hitbox_entries.clear()
		return

	hitbox_entries = next_entries.duplicate(true)
	queue_redraw()


func _draw() -> void:
	if !visible:
		return

	for entry in hitbox_entries:
		if !(entry is Dictionary):
			continue

		var points_value = entry.get("points", PackedVector2Array())
		if !(points_value is PackedVector2Array):
			continue

		var points: PackedVector2Array = points_value
		if points.is_empty():
			continue

		draw_polyline(_closed_outline(points), Color(0.95, 0.95, 0.95, 0.9), 2.0, true)


func _closed_outline(points: PackedVector2Array) -> PackedVector2Array:
	if points.size() < 2:
		return points

	var closed_outline := points.duplicate()
	closed_outline.append(points[0])
	return closed_outline
