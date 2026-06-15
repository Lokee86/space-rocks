extends GutTest

const TransmissionFlow := preload("res://scripts/ui/menu_flow/transmission_flow.gd")
const PregameMenuScene := preload("res://scenes/ui/pregame_menu.tscn")


func test_configure_finds_screen_display() -> void:
	var pregame_menu := await _create_pregame_menu()
	var flow := TransmissionFlow.new()

	flow.configure(pregame_menu)

	assert_not_null(flow.get_screen_display())
	assert_true(flow.get_screen_display() is Control)


func test_mount_adds_scene_to_screen_display() -> void:
	var pregame_menu := await _create_pregame_menu()
	var flow := TransmissionFlow.new()
	var transmission_scene := _create_transmission_scene("FirstTransmission")

	flow.configure(pregame_menu)

	var mounted := flow.mount(transmission_scene)

	assert_not_null(mounted)
	assert_eq(flow.get_screen_display().get_child_count(), 1)
	assert_eq(flow.get_screen_display().get_child(0), mounted)
	assert_eq(mounted.name, "FirstTransmission")


func test_mount_replaces_existing_child() -> void:
	var pregame_menu := await _create_pregame_menu()
	var flow := TransmissionFlow.new()
	var first_scene := _create_transmission_scene("FirstTransmission")
	var second_scene := _create_transmission_scene("SecondTransmission")

	flow.configure(pregame_menu)

	var first_mounted := flow.mount(first_scene)
	var second_mounted := flow.mount(second_scene)

	assert_eq(first_mounted.is_queued_for_deletion(), true)
	assert_eq(flow.get_screen_display().get_child_count(), 1)
	assert_eq(flow.get_screen_display().get_child(0), second_mounted)
	assert_eq(second_mounted.name, "SecondTransmission")
	await get_tree().process_frame


func test_clear_removes_children_and_deactivates_transmission() -> void:
	var pregame_menu := await _create_pregame_menu()
	var flow := TransmissionFlow.new()
	var transmission_scene := _create_transmission_scene("MountedTransmission")

	flow.configure(pregame_menu)
	flow.mount(transmission_scene)

	flow.clear()

	assert_eq(flow.get_screen_display().get_child_count(), 0)
	assert_false(flow.has_active_transmission())
	await get_tree().process_frame


func test_has_active_transmission_false_after_clear() -> void:
	var pregame_menu := await _create_pregame_menu()
	var flow := TransmissionFlow.new()
	var transmission_scene := _create_transmission_scene("MountedTransmission")

	flow.configure(pregame_menu)
	flow.mount(transmission_scene)
	flow.clear()

	assert_false(flow.has_active_transmission())
	await get_tree().process_frame


func _create_pregame_menu() -> Control:
	var pregame_menu := PregameMenuScene.instantiate()
	add_child_autofree(pregame_menu)
	await get_tree().process_frame
	return pregame_menu


func _create_transmission_scene(node_name: String) -> PackedScene:
	var root := Control.new()
	root.name = node_name
	var scene := PackedScene.new()
	var pack_error := scene.pack(root)
	assert_eq(pack_error, OK)
	root.free()
	return scene
