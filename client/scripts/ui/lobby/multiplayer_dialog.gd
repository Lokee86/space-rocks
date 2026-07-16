extends Control

const Constants := preload("res://scripts/generated/constants/constants.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")

signal create_room_requested
signal join_room_requested(room_code: String)
signal canceled

var room_code_input: LineEdit
var status_label: Label


func _ready() -> void:
	room_code_input = find_child("RoomCodeInput", true, false) as LineEdit
	status_label = find_child("StatusLabel", true, false) as Label
	var create_room_button := find_child("CreateButton", true, false) as BaseButton
	var join_room_button := find_child("JoinButton", true, false) as BaseButton
	var cancel_button := find_child("CancelButton", true, false) as BaseButton

	if room_code_input == null:
		ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Missing required lobby presentation node",
		{},
		{
			"subsystem": "lobby",
			"failure_mode": "missing_node",
			"node_name": "RoomCodeInput",
			"resource_kind": "input",
			"expected_type": "LineEdit",
			"actual_type": "null",
		}
	)
	if status_label == null:
		ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Missing required lobby presentation node",
		{},
		{
			"subsystem": "lobby",
			"failure_mode": "missing_node",
			"node_name": "StatusLabel",
			"resource_kind": "label",
			"expected_type": "Label",
			"actual_type": "null",
		}
	)

	if create_room_button == null:
		ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Missing required lobby presentation node",
		{},
		{
			"subsystem": "lobby",
			"failure_mode": "missing_node",
			"node_name": "CreateButton",
			"resource_kind": "button",
			"expected_type": "BaseButton",
			"actual_type": "null",
		}
	)
	else:
		create_room_button.pressed.connect(_on_create_room_pressed)

	if join_room_button == null:
		ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Missing required lobby presentation node",
		{},
		{
			"subsystem": "lobby",
			"failure_mode": "missing_node",
			"node_name": "JoinButton",
			"resource_kind": "button",
			"expected_type": "BaseButton",
			"actual_type": "null",
		}
	)
	else:
		join_room_button.pressed.connect(_on_join_room_pressed)

	if cancel_button == null:
		ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Missing required lobby presentation node",
		{},
		{
			"subsystem": "lobby",
			"failure_mode": "missing_node",
			"node_name": "CancelButton",
			"resource_kind": "button",
			"expected_type": "BaseButton",
			"actual_type": "null",
		}
	)
	else:
		cancel_button.pressed.connect(_on_cancel_pressed)


func _on_create_room_pressed() -> void:
	create_room_requested.emit()


func _on_join_room_pressed() -> void:
	var room_code := ""
	if room_code_input != null:
		room_code = room_code_input.text.strip_edges()
	set_status(Constants.DIALOG_STATUS_JOINING_ROOM)
	join_room_requested.emit(room_code)


func _on_cancel_pressed() -> void:
	canceled.emit()


func set_status(message: String) -> void:
	if status_label != null:
		status_label.text = message


func show_join_error(message: String) -> void:
	set_status(message)

