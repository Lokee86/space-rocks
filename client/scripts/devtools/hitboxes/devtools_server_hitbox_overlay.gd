extends Node2D

const DevtoolsHitboxTemplateCatalog = preload("res://scripts/devtools/hitboxes/devtools_hitbox_template_catalog.gd")

var hitbox_entries: Array = []
var hitbox_template_catalog := DevtoolsHitboxTemplateCatalog.new()


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

		var kind := str(entry.get("kind", ""))
		var template := PackedVector2Array()
		if kind == "player":
			template = hitbox_template_catalog.player_polygon()
		elif kind == "asteroid":
			template = hitbox_template_catalog.asteroid_polygon(int(entry.get("variant", 0)))
		elif kind == "bullet":
			template = hitbox_template_catalog.bullet_polygon()
		else:
			continue

		if template.is_empty():
			continue

		var rotation := float(entry.get("rotation", 0.0))
		var scale := float(entry.get("scale", 1.0))
		var visual_position: Vector2 = entry.get("visual_position", Vector2.ZERO)
		var outline := PackedVector2Array()
		outline.resize(template.size())

		for index in range(template.size()):
			var point := template[index]
			var transformed_point := point.rotated(rotation) * scale + visual_position
			outline[index] = transformed_point

		draw_polyline(_closed_outline(outline), Color(0.95, 0.95, 0.95, 0.9), 2.0, true)


func _closed_outline(points: PackedVector2Array) -> PackedVector2Array:
	if points.size() < 2:
		return points

	var closed_outline := points.duplicate()
	closed_outline.append(points[0])
	return closed_outline
