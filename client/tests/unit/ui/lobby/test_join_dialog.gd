extends GutTest

const JoinDialogScene := preload("res://scenes/ui/dialogs/join_dialog.tscn")


func test_join_button_emits_trimmed_room_code() -> void:
	var dialog := await _create_dialog()

	(dialog.get_node_or_null("%RoomCodeInput") as LineEdit).text = "  ABCD  "
	watch_signals(dialog)

	(dialog.get_node_or_null("%JoinButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted_with_parameters(dialog, "join_requested", ["ABCD"])


func test_join_button_emits_empty_room_code_when_blank() -> void:
	var dialog := await _create_dialog()

	(dialog.get_node_or_null("%RoomCodeInput") as LineEdit).text = "   "
	watch_signals(dialog)

	(dialog.get_node_or_null("%JoinButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted_with_parameters(dialog, "join_requested", [""])


func test_cancel_button_emits_cancel_requested() -> void:
	var dialog := await _create_dialog()

	watch_signals(dialog)
	(dialog.get_node_or_null("%CancelButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted(dialog, "cancel_requested")


func test_set_status_updates_status_label() -> void:
	var dialog := await _create_dialog()

	dialog.set_status("Must enter ID")

	assert_eq((dialog.get_node_or_null("%StatusLabel") as Label).text, "Must enter ID")


func test_clear_status_clears_status_label() -> void:
	var dialog := await _create_dialog()

	dialog.set_status("Must enter ID")
	dialog.clear_status()

	assert_eq((dialog.get_node_or_null("%StatusLabel") as Label).text, "")


func _create_dialog() -> Control:
	var dialog := JoinDialogScene.instantiate()
	add_child_autofree(dialog)
	await get_tree().process_frame
	return dialog
