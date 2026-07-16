extends GutTest

const LobbyPlayerListView := preload("res://scripts/ui/lobby/lobby_player_list_view.gd")
const PlayerRow := preload("res://scripts/ui/lobby/player_row.gd")
const PlayerRowScene := preload("res://scenes/ui/elements/player_row.tscn")


func test_render_adds_and_configures_valid_player_row_scene() -> void:
	var container := VBoxContainer.new()
	add_child_autofree(container)
	var row_scene := PlayerRowScene

	LobbyPlayerListView.render(container, row_scene, "Player-1", "Player-1", [{"player_id": "Player-1", "ready": true}])

	assert_eq(container.get_child_count(), 1)
	assert_true(container.get_child(0) is PlayerRow)
	assert_eq(container.get_child(0).player_name_label.text, "Player-1 (You)")
	assert_eq(container.get_child(0).player_ready_label.text, "Ready")


func test_render_rejects_invalid_player_row_scene() -> void:
	var container := VBoxContainer.new()
	add_child_autofree(container)
	var invalid_scene := _create_scene(Control.new())

	LobbyPlayerListView.render(container, invalid_scene, "Player-1", "", [{"player_id": "Player-1"}])
	assert_push_error_count(1)

	assert_eq(container.get_child_count(), 0)
	await get_tree().process_frame


func _create_scene(root: Node) -> PackedScene:
	var scene := PackedScene.new()
	assert_eq(scene.pack(root), OK)
	root.free()
	return scene