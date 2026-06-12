extends GutTest

const JoinDialogFlow := preload("res://scripts/ui/lobby/join_dialog_flow.gd")


class FakeJoinDialog:
	extends Control

	signal join_requested(room_code: String)
	signal cancel_requested

	var status_text := ""

	func set_status(message: String) -> void:
		status_text = message


class Probe:
	extends RefCounted

	var calls := 0

	func mark_called() -> void:
		calls += 1


class JoinProbe:
	extends RefCounted

	var calls := 0
	var last_room_code := ""

	func mark_join(room_code: String) -> void:
		calls += 1
		last_room_code = room_code


func test_cancel_calls_close_callback() -> void:
	var dialog := FakeJoinDialog.new()
	var flow := JoinDialogFlow.new()
	var close_probe := Probe.new()

	add_child_autofree(dialog)
	flow.configure(dialog, Callable(close_probe, "mark_called"), Callable())

	dialog.cancel_requested.emit()

	assert_eq(close_probe.calls, 1)


func test_empty_join_code_sets_status_and_does_not_clear_or_join() -> void:
	var dialog := FakeJoinDialog.new()
	var flow := JoinDialogFlow.new()
	var clear_probe := Probe.new()
	var join_probe := JoinProbe.new()

	add_child_autofree(dialog)
	flow.configure(
		dialog,
		Callable(),
		Callable(join_probe, "mark_join"),
		Callable(clear_probe, "mark_called"))

	dialog.join_requested.emit("")

	assert_eq(dialog.status_text, "Must enter an ID to join.")
	assert_eq(clear_probe.calls, 0)
	assert_eq(join_probe.calls, 0)


func test_whitespace_join_code_sets_status_and_does_not_clear_or_join() -> void:
	var dialog := FakeJoinDialog.new()
	var flow := JoinDialogFlow.new()
	var clear_probe := Probe.new()
	var join_probe := JoinProbe.new()

	add_child_autofree(dialog)
	flow.configure(
		dialog,
		Callable(),
		Callable(join_probe, "mark_join"),
		Callable(clear_probe, "mark_called"))

	dialog.join_requested.emit("   ")

	assert_eq(dialog.status_text, "Must enter an ID to join.")
	assert_eq(clear_probe.calls, 0)
	assert_eq(join_probe.calls, 0)


func test_valid_join_code_calls_clear_and_join_with_stripped_code() -> void:
	var dialog := FakeJoinDialog.new()
	var flow := JoinDialogFlow.new()
	var clear_probe := Probe.new()
	var join_probe := JoinProbe.new()

	add_child_autofree(dialog)
	flow.configure(
		dialog,
		Callable(),
		Callable(join_probe, "mark_join"),
		Callable(clear_probe, "mark_called"))

	dialog.join_requested.emit("  ABCD  ")

	assert_eq(dialog.status_text, "")
	assert_eq(clear_probe.calls, 1)
	assert_eq(join_probe.calls, 1)
	assert_eq(join_probe.last_room_code, "ABCD")
