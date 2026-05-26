extends RefCounted
class_name OSIndicatorController

const Constants = preload("res://scripts/constants/constants.gd")
const OS_INDICATOR_SCENE = preload("res://scenes/ui/elements/osindicator.tscn")

const INDICATOR_SIZE := Vector2(14.0, 14.0)
const INDICATOR_ROTATION_OFFSET := PI * 0.5

var hud: Control
var indicator_nodes := {}


func configure(hud_ref: Control) -> void:
	hud = hud_ref


func reset() -> void:
	for node in indicator_nodes.values():
		if is_instance_valid(node):
			node.queue_free()
	indicator_nodes.clear()


func update_indicators(
	camera: Camera2D,
	target_positions: Dictionary,
	target_hues: Dictionary
) -> void:
	if hud == null || camera == null:
		for indicator in indicator_nodes.values():
			if is_instance_valid(indicator):
				indicator.hide()
		return

	for target_id in target_positions:
		var indicator := _indicator_for_target(str(target_id))
		var target_position: Vector2 = target_positions[target_id]
		var screen_position := _target_screen_position(camera, target_position)
		if _is_inside_visible_area(screen_position):
			indicator.hide()
			continue

		var hue := Constants.REMOTE_PLAYER_FALLBACK_HUE
		if target_hues.has(target_id):
			hue = target_hues[target_id]
		_apply_indicator_hue(indicator, float(hue))
		var indicator_size := _indicator_size(indicator)
		var clamped_center := _clamp_to_indicator_bounds(screen_position, indicator_size)
		indicator.position = clamped_center - (indicator_size * 0.5)
		var viewport_center := hud.get_viewport_rect().size * 0.5
		var direction := screen_position - viewport_center
		indicator.rotation = direction.angle() + INDICATOR_ROTATION_OFFSET
		indicator.show()

	for target_id in indicator_nodes.keys():
		if target_positions.has(target_id):
			continue
		var indicator = indicator_nodes[target_id]
		if is_instance_valid(indicator):
			indicator.queue_free()
		indicator_nodes.erase(target_id)


func _target_screen_position(camera: Camera2D, target_position: Vector2) -> Vector2:
	return camera.get_viewport().get_canvas_transform() * target_position


func _indicator_bounds() -> Rect2:
	var viewport_size := hud.get_viewport_rect().size
	var margin := Constants.OSINDICATOR_EDGE_MARGIN
	return Rect2(
		Vector2(margin, margin),
		viewport_size - Vector2(margin * 2.0, margin * 2.0)
	)


func _is_inside_visible_area(screen_position: Vector2) -> bool:
	var viewport_size := hud.get_viewport_rect().size
	var padding := Constants.OSINDICATOR_TARGET_VISIBILITY_PADDING
	return Rect2(
		Vector2(-padding, -padding),
		viewport_size + Vector2(padding * 2.0, padding * 2.0)
	).has_point(screen_position)


func _clamp_to_indicator_bounds(screen_position: Vector2, indicator_size: Vector2) -> Vector2:
	var bounds := _indicator_bounds()
	var half_size := indicator_size * 0.5
	return Vector2(
		clamp(screen_position.x, bounds.position.x + half_size.x, bounds.end.x - half_size.x),
		clamp(screen_position.y, bounds.position.y + half_size.y, bounds.end.y - half_size.y)
	)


func _indicator_size(indicator: Control) -> Vector2:
	if indicator.size != Vector2.ZERO:
		return indicator.size
	if indicator.custom_minimum_size != Vector2.ZERO:
		return indicator.custom_minimum_size
	return INDICATOR_SIZE


func _indicator_for_target(target_id: String) -> Control:
	if indicator_nodes.has(target_id) && is_instance_valid(indicator_nodes[target_id]):
		return indicator_nodes[target_id]

	var indicator := OS_INDICATOR_SCENE.instantiate() as Control
	indicator.visible = false
	hud.add_child(indicator)
	indicator_nodes[target_id] = indicator
	return indicator


func _apply_indicator_hue(indicator: Control, hue: float) -> void:
	var graphic := indicator.get_node_or_null("TextureRect") as CanvasItem
	if graphic == null:
		return
	var shader_material := graphic.material as ShaderMaterial
	if shader_material == null:
		return
	graphic.material = shader_material.duplicate() as ShaderMaterial
	(graphic.material as ShaderMaterial).set_shader_parameter("hue_shift", fposmod(hue, 1.0))
