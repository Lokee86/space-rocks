extends GutTest

const DiscreteListViewScene := preload("res://scenes/ui/elements/discrete_list_view.tscn")
const PilotSelectRowScene := preload("res://scenes/ui/elements/pilot_select_row.tscn")


func test_render_mounts_pilot_rows_and_wires_selection() -> void:
	var view := DiscreteListViewScene.instantiate()
	add_child_autofree(view)
	view.row_scene = PilotSelectRowScene
	view.set_items([{"display_name": "Pilot One", "pilot_id": "pilot-1"}])

	assert_eq(view.rows.get_child_count(), 1)
	var row: DiscreteListRow = view.rows.get_child(0) as DiscreteListRow
	var expected_item := {"display_name": "Pilot One", "pilot_id": "pilot-1"}
	assert_eq((row.get_node("PanelContainer/Label") as Label).text, "Pilot One")

	var selection := []
	view.selection_changed.connect(func(item: Dictionary): selection.append(item))
	row.selected.emit(expected_item)

	assert_eq(selection, [{"display_name": "Pilot One", "pilot_id": "pilot-1"}])


func test_render_rejects_invalid_row_scene() -> void:
	var view := DiscreteListViewScene.instantiate()
	add_child_autofree(view)
	var invalid_scene := PackedScene.new()
	var invalid_root := Control.new()
	assert_eq(invalid_scene.pack(invalid_root), OK)
	invalid_root.free()
	view.row_scene = invalid_scene

	view.set_items([{"display_name": "Invalid"}])
	assert_push_error("Discrete list row scene must instantiate DiscreteListRow")

	assert_eq(view.rows.get_child_count(), 0)
	await get_tree().process_frame
