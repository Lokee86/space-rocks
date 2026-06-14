extends GutTest

const ConfirmDeleteScene := preload("res://scenes/ui/transmission_displays/sub-transmissions/confirm_delete.tscn")


func test_confirm_button_emits_requested_item() -> void:
	var dialog := await _create_dialog()
	var item := {
		"identity_kind": "local_profile",
		"local_profile_id": "local-profile-test",
		"display_name": "TEST",
	}

	dialog.configure_delete(item)
	watch_signals(dialog)
	(dialog.get_node_or_null("%ConfirmButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted_with_parameters(dialog, "confirm_requested", [item])


func test_cancel_button_emits_cancel_requested() -> void:
	var dialog := await _create_dialog()

	watch_signals(dialog)
	(dialog.get_node_or_null("%CancelButton") as BaseButton).emit_signal("pressed")

	assert_signal_emitted(dialog, "cancel_requested")


func _create_dialog() -> Control:
	var dialog := ConfirmDeleteScene.instantiate()
	add_child_autofree(dialog)
	await get_tree().process_frame
	return dialog
