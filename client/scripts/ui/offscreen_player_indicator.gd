extends Control

const PLAYER_HUE_SHIFT_SHADER := preload("res://shaders/player_hue_shift.gdshader")

@export var indicator_visual_path: NodePath = ^"TextureRect"

@onready var indicator_visual: CanvasItem = get_node_or_null(indicator_visual_path) as CanvasItem

var indicator_hue_material: ShaderMaterial


func set_indicator(
	screen_position: Vector2,
	direction: Vector2,
	hue_shift: float = 0.0
) -> void:
	var center_offset: Vector2 = pivot_offset
	if center_offset == Vector2.ZERO:
		center_offset = size * 0.5
	position = screen_position - center_offset
	rotation = direction.angle() + PI / 2.0
	set_indicator_hue(hue_shift)
	visible = true


func hide_indicator() -> void:
	visible = false


func set_indicator_hue(hue_shift: float) -> void:
	var material := _ensure_unique_indicator_hue_material()
	if material == null:
		return

	material.set_shader_parameter("hue_shift", fposmod(hue_shift, 1.0))


func _ensure_unique_indicator_hue_material() -> ShaderMaterial:
	if indicator_hue_material != null:
		return indicator_hue_material
	if indicator_visual == null:
		return null

	var existing_material := indicator_visual.material as ShaderMaterial
	if existing_material != null && existing_material.shader == PLAYER_HUE_SHIFT_SHADER:
		indicator_hue_material = existing_material.duplicate() as ShaderMaterial
	else:
		indicator_hue_material = ShaderMaterial.new()
		indicator_hue_material.shader = PLAYER_HUE_SHIFT_SHADER

	indicator_visual.material = indicator_hue_material
	return indicator_hue_material
