extends GutTest

const GameplayBackgroundFlowScript := preload("res://scripts/presentation/background/background_flow.gd")
const BackgroundShader := preload("res://shaders/repeating_background.gdshader")


func test_process_frame_applies_camera_zoom_to_every_background_layer() -> void:
	var camera := Camera2D.new()
	camera.zoom = Vector2(2.0, 2.0)
	add_child_autofree(camera)
	var layers := _create_layers()
	var flow = GameplayBackgroundFlowScript.new()
	flow.configure(layers[0], layers[1], layers[2], null, camera)

	flow.process_frame()

	for layer in layers:
		var material := layer.material as ShaderMaterial
		assert_almost_eq(float(material.get_shader_parameter("tile_scale")), 2.0, 0.0001)


func test_process_frame_applies_zoomed_out_scale() -> void:
	var camera := Camera2D.new()
	camera.zoom = Vector2(0.5, 0.5)
	add_child_autofree(camera)
	var layers := _create_layers()
	var flow = GameplayBackgroundFlowScript.new()
	flow.configure(layers[0], layers[1], layers[2], null, camera)

	flow.process_frame()

	for layer in layers:
		var material := layer.material as ShaderMaterial
		assert_almost_eq(float(material.get_shader_parameter("tile_scale")), 0.5, 0.0001)


func test_missing_camera_uses_neutral_background_scale() -> void:
	var layers := _create_layers()
	var flow = GameplayBackgroundFlowScript.new()
	flow.configure(layers[0], layers[1], layers[2], null)

	flow.process_frame()

	for layer in layers:
		var material := layer.material as ShaderMaterial
		assert_almost_eq(float(material.get_shader_parameter("tile_scale")), 1.0, 0.0001)


func _create_layers() -> Array[TextureRect]:
	var layers: Array[TextureRect] = []
	for index in 3:
		var layer := TextureRect.new()
		var material := ShaderMaterial.new()
		material.shader = BackgroundShader
		layer.material = material
		add_child_autofree(layer)
		layers.append(layer)
	return layers
