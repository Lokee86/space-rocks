extends GutTest

const OffscreenIndicatorsScript := preload("res://scripts/ui/offscreen_indicators.gd")
const IndicatorScene := preload("res://scenes/ui/elements/osindicator.tscn")


class RecordingIndicator:
	extends Control

	var last_hue_shift := -1.0

	func set_indicator(_screen_position: Vector2, _direction: Vector2, hue_shift: float = 0.0) -> void:
		last_hue_shift = hue_shift
		visible = true

	func hide_indicator() -> void:
		visible = false


var manager: Control
var camera: Camera2D


func before_each() -> void:
	manager = OffscreenIndicatorsScript.new()
	manager.indicator_scene = _recording_indicator_scene()
	add_child(manager)
	manager.size = Vector2(800.0, 600.0)

	camera = Camera2D.new()
	add_child(camera)


func after_each() -> void:
	if manager != null:
		manager.free()
		manager = null
	if camera != null:
		camera.free()
		camera = null


func test_indicator_manager_passes_remote_hue_to_indicator() -> void:
	manager.update_indicators(
		{"remote-player": Vector2(10000.0, 0.0)},
		camera,
		{"remote-player": 0.42}
	)

	var indicator = manager.indicators["remote-player"]
	assert_almost_eq(indicator.last_hue_shift, 0.42, 0.0001)


func test_indicator_manager_missing_hue_falls_back_to_zero() -> void:
	manager.update_indicators(
		{"remote-player": Vector2(10000.0, 0.0)},
		camera,
		{}
	)

	var indicator = manager.indicators["remote-player"]
	assert_eq(indicator.last_hue_shift, 0.0)


func test_offscreen_indicator_applies_hue_to_shader_parameter() -> void:
	var indicator := IndicatorScene.instantiate()
	add_child(indicator)

	indicator.set_indicator(Vector2(50.0, 60.0), Vector2.RIGHT, 0.42)

	var texture_rect := indicator.get_node("TextureRect") as TextureRect
	var material := texture_rect.material as ShaderMaterial
	assert_not_null(material)
	assert_almost_eq(float(material.get_shader_parameter("hue_shift")), 0.42, 0.0001)

	indicator.free()


func _recording_indicator_scene() -> PackedScene:
	var indicator := RecordingIndicator.new()
	var scene := PackedScene.new()
	var result := scene.pack(indicator)
	assert_eq(result, OK)
	indicator.free()
	return scene
