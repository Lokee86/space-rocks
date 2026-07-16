extends Control

const ClientLogger := preload("res://scripts/logging/logger.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")

signal single_player_requested
signal multiplayer_requested
signal logout_requested

var login_status_label: Label
var single_player_button: BaseButton
var logout_button: BaseButton
var multiplayer_button: BaseButton
var multiplayer_label: Label
var sign_in_label: Label
var _signed_in := false


func _ready() -> void:
	login_status_label = find_child("LoginStatusLabel", true, false) as Label
	single_player_button = get_node_or_null("%SinglePlayerButton") as BaseButton
	logout_button = get_node_or_null("%LogoutButton") as BaseButton
	multiplayer_button = get_node_or_null("%MultiplayerButton") as BaseButton
	multiplayer_label = find_child("Multi-player", true, false) as Label
	sign_in_label = find_child("Sign-in", true, false) as Label
	var quit_button := get_node_or_null("%QuitButton") as BaseButton

	if login_status_label == null:
		ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Missing required main menu presentation node",
		{},
		{
			"subsystem": "main_menu",
			"failure_mode": "missing_node",
			"node_name": "LoginStatusLabel",
			"resource_kind": "label",
			"expected_type": "Label",
			"actual_type": "null",
		}
	)
	if single_player_button == null:
		ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Missing required main menu presentation node",
		{},
		{
			"subsystem": "main_menu",
			"failure_mode": "missing_node",
			"node_name": "SinglePlayerButton",
			"resource_kind": "button",
			"expected_type": "BaseButton",
			"actual_type": "null",
		}
	)
	if logout_button == null:
		ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Missing required main menu presentation node",
		{},
		{
			"subsystem": "main_menu",
			"failure_mode": "missing_node",
			"node_name": "LogoutButton",
			"resource_kind": "button",
			"expected_type": "BaseButton",
			"actual_type": "null",
		}
	)
	if multiplayer_button == null:
		ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Missing required main menu presentation node",
		{},
		{
			"subsystem": "main_menu",
			"failure_mode": "missing_node",
			"node_name": "MultiplayerButton",
			"resource_kind": "button",
			"expected_type": "BaseButton",
			"actual_type": "null",
		}
	)
	if multiplayer_label == null:
		ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Missing required main menu presentation node",
		{},
		{
			"subsystem": "main_menu",
			"failure_mode": "missing_node",
			"node_name": "Multi-player",
			"resource_kind": "label",
			"expected_type": "Label",
			"actual_type": "null",
		}
	)
	if sign_in_label == null:
		ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Missing required main menu presentation node",
		{},
		{
			"subsystem": "main_menu",
			"failure_mode": "missing_node",
			"node_name": "Sign-in",
			"resource_kind": "label",
			"expected_type": "Label",
			"actual_type": "null",
		}
	)
	if single_player_button != null:
		single_player_button.pressed.connect(_on_single_player_pressed)

	if multiplayer_button != null:
		multiplayer_button.pressed.connect(_on_multiplayer_pressed)

	if logout_button != null:
		logout_button.pressed.connect(_on_logout_pressed)

	if quit_button == null:
		ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Missing required main menu presentation node",
		{},
		{
			"subsystem": "main_menu",
			"failure_mode": "missing_node",
			"node_name": "QuitButton",
			"resource_kind": "button",
			"expected_type": "BaseButton",
			"actual_type": "null",
		}
	)
	else:
		quit_button.pressed.connect(_on_quit_pressed)

	show_signed_out()


func _on_single_player_pressed() -> void:
	single_player_requested.emit()


func _on_multiplayer_pressed() -> void:
	multiplayer_requested.emit()


func _on_logout_pressed() -> void:
	logout_requested.emit()


func _on_quit_pressed() -> void:
	get_tree().quit()


func show_signed_out() -> void:
	_signed_in = false
	if login_status_label != null:
		login_status_label.text = "Not Signed In"
	if logout_button != null:
		logout_button.visible = false
	if multiplayer_label != null:
		multiplayer_label.visible = true
	if sign_in_label != null:
		sign_in_label.visible = false


func show_signed_in(display_name: String) -> void:
	_signed_in = true
	if login_status_label != null:
		login_status_label.text = display_name
	if logout_button != null:
		logout_button.visible = true
	if multiplayer_label != null:
		multiplayer_label.visible = true
	if sign_in_label != null:
		sign_in_label.visible = false
