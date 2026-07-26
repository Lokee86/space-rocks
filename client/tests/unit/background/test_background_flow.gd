extends GutTest

const GameplayBackgroundFlowScript := preload("res://scripts/presentation/background/background_flow.gd")
const BackgroundShader := preload("res://shaders/repeating_background.gdshader")
const Constants := preload("res://scripts/generated/constants/constants.gd")


func test_process_frame_writes_camera_world_state_to_every_layer() -> void:
	var camera := Camera2D.new()
	camera.zoom = Vector2(2.0, 2.0)
	add_child_autofree(camera)
	var target := Node2D.new()
	target.global_position = Vector2(320.0, 180.0)
	add_child_autofree(target)
	var layers := _create_layers()
	var flow = GameplayBackgroundFlowScript.new()
	flow.configure(layers[0], layers[1], layers[2], target, camera)

	flow.process_frame()

	for layer in layers:
		var material := layer.material as ShaderMaterial
		assert_eq(material.get_shader_parameter("camera_world_position"), Vector2(320.0, 180.0))
		assert_almost_eq(float(material.get_shader_parameter("camera_zoom")), 2.0, 0.0001)


func test_process_frame_writes_viewport_stretch_scale_to_every_layer() -> void:
	var camera := Camera2D.new()
	camera.zoom = Vector2.ONE
	add_child_autofree(camera)
	var layers := _create_layers()
	var flow = GameplayBackgroundFlowScript.new()
	flow.configure(
		layers[0],
		layers[1],
		layers[2],
		null,
		camera,
		func() -> Vector2: return Vector2(1.5, 1.25)
	)

	flow.process_frame()

	for layer in layers:
		var material := layer.material as ShaderMaterial
		assert_eq(
			material.get_shader_parameter("viewport_stretch_scale"),
			Vector2(1.5, 1.25)
		)


func test_process_frame_writes_distinct_world_parallax_and_offsets() -> void:
	var layers := _create_layers()
	var flow = GameplayBackgroundFlowScript.new()
	flow.configure(layers[0], layers[1], layers[2], null)

	flow.process_frame()

	var background_material := layers[0].material as ShaderMaterial
	var foreground_material := layers[1].material as ShaderMaterial
	var planet_material := layers[2].material as ShaderMaterial
	assert_almost_eq(
		float(background_material.get_shader_parameter("parallax_factor")),
		Constants.BACKGROUND_PARALLAX,
		0.0001
	)
	assert_eq(
		background_material.get_shader_parameter("layer_world_offset"),
		Constants.BACKGROUND_DRIFT_PER_FRAME
	)
	assert_almost_eq(
		float(foreground_material.get_shader_parameter("parallax_factor")),
		Constants.FOREGROUND_BACKGROUND_PARALLAX,
		0.0001
	)
	assert_eq(
		foreground_material.get_shader_parameter("layer_world_offset"),
		Constants.FOREGROUND_BACKGROUND_DRIFT_PER_FRAME + Constants.FOREGROUND_BACKGROUND_OFFSET
	)
	assert_almost_eq(
		float(planet_material.get_shader_parameter("parallax_factor")),
		float(Constants.PLANET_BACKGROUND_PARALLAX),
		0.0001
	)
	assert_eq(
		planet_material.get_shader_parameter("layer_world_offset"),
		Constants.PLANET_BACKGROUND_DRIFT_PER_FRAME + Constants.PLANET_BACKGROUND_OFFSET
	)


func test_missing_camera_uses_neutral_world_zoom() -> void:
	var layers := _create_layers()
	var flow = GameplayBackgroundFlowScript.new()
	flow.configure(layers[0], layers[1], layers[2], null)

	flow.process_frame()

	for layer in layers:
		var material := layer.material as ShaderMaterial
		assert_almost_eq(float(material.get_shader_parameter("camera_zoom")), 1.0, 0.0001)


func test_set_scroll_reference_updates_world_origin_without_drift() -> void:
	var layers := _create_layers()
	var flow = GameplayBackgroundFlowScript.new()
	flow.configure(layers[0], layers[1], layers[2], null)

	flow.set_scroll_reference(Vector2(640.0, 360.0))

	for layer in layers:
		var material := layer.material as ShaderMaterial
		assert_eq(material.get_shader_parameter("camera_world_position"), Vector2(640.0, 360.0))
	assert_eq(
		(layers[0].material as ShaderMaterial).get_shader_parameter("layer_world_offset"),
		Vector2.ZERO
	)
	assert_eq(
		(layers[1].material as ShaderMaterial).get_shader_parameter("layer_world_offset"),
		Constants.FOREGROUND_BACKGROUND_OFFSET
	)


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
