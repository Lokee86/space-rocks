class_name JoinDialogFlow
extends RefCounted

const Constants := preload("res://scripts/generated/constants/constants.gd")

var join_dialog: Control
var close_join_dialog_callable: Callable
var join_room_callable: Callable
var clear_for_room_transition_callable: Callable


func configure(
		join_dialog_ref: Control,
		close_join_dialog_callable_ref: Callable,
		join_room_callable_ref: Callable,
		clear_for_room_transition_callable_ref: Callable = Callable()) -> void:
	join_dialog = join_dialog_ref
	close_join_dialog_callable = close_join_dialog_callable_ref
	join_room_callable = join_room_callable_ref
	clear_for_room_transition_callable = clear_for_room_transition_callable_ref

	if join_dialog != null and join_dialog.has_signal("join_requested"):
		if not join_dialog.join_requested.is_connected(_on_join_requested):
			join_dialog.join_requested.connect(_on_join_requested)
	if join_dialog != null and join_dialog.has_signal("cancel_requested"):
		if not join_dialog.cancel_requested.is_connected(_on_cancel_requested):
			join_dialog.cancel_requested.connect(_on_cancel_requested)


func _on_cancel_requested() -> void:
	if close_join_dialog_callable.is_valid():
		close_join_dialog_callable.call()


func _on_join_requested(room_code: String) -> void:
	var stripped_room_code := room_code.strip_edges()
	if stripped_room_code.is_empty():
		if join_dialog != null and join_dialog.has_method("set_status"):
			join_dialog.set_status(Constants.DIALOG_STATUS_MUST_ENTER_ID)
		return
	if clear_for_room_transition_callable.is_valid():
		clear_for_room_transition_callable.call()
	if join_room_callable.is_valid():
		join_room_callable.call(stripped_room_code)
