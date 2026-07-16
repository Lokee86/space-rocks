extends Control

const ClientLogger := preload("res://scripts/logging/logger.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")

signal back_requested
signal discord_login_requested


func _ready() -> void:
	var email_input := get_node_or_null("%EmailInput") as LineEdit
	var password_input := get_node_or_null("%PasswordInput") as LineEdit
	var sign_in_button := get_node_or_null("%SignInButton") as BaseButton
	var back_button := get_node_or_null("%BackButton") as BaseButton
	var google_login_button := get_node_or_null("%GoogleLoginButton") as BaseButton
	var discord_login_button := get_node_or_null("%DiscordLoginButton") as BaseButton

	if email_input != null:
		email_input.editable = false
	else:
		ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Missing required sign-in presentation node",
		{},
		{
			"subsystem": "sign_in",
			"failure_mode": "missing_node",
			"node_name": "EmailInput",
			"resource_kind": "input",
			"expected_type": "LineEdit",
			"actual_type": "null",
		}
	)

	if password_input != null:
		password_input.editable = false
	else:
		ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Missing required sign-in presentation node",
		{},
		{
			"subsystem": "sign_in",
			"failure_mode": "missing_node",
			"node_name": "PasswordInput",
			"resource_kind": "input",
			"expected_type": "LineEdit",
			"actual_type": "null",
		}
	)

	if sign_in_button != null:
		sign_in_button.disabled = true
	else:
		ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Missing required sign-in presentation node",
		{},
		{
			"subsystem": "sign_in",
			"failure_mode": "missing_node",
			"node_name": "SignInButton",
			"resource_kind": "button",
			"expected_type": "BaseButton",
			"actual_type": "null",
		}
	)

	if google_login_button != null:
		google_login_button.disabled = true
	else:
		ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Missing required sign-in presentation node",
		{},
		{
			"subsystem": "sign_in",
			"failure_mode": "missing_node",
			"node_name": "GoogleLoginButton",
			"resource_kind": "button",
			"expected_type": "BaseButton",
			"actual_type": "null",
		}
	)

	if back_button != null:
		back_button.pressed.connect(_on_back_pressed)
	else:
		ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Missing required sign-in presentation node",
		{},
		{
			"subsystem": "sign_in",
			"failure_mode": "missing_node",
			"node_name": "BackButton",
			"resource_kind": "button",
			"expected_type": "BaseButton",
			"actual_type": "null",
		}
	)

	if discord_login_button != null:
		discord_login_button.pressed.connect(_on_discord_login_pressed)
	else:
		ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Missing required sign-in presentation node",
		{},
		{
			"subsystem": "sign_in",
			"failure_mode": "missing_node",
			"node_name": "DiscordLoginButton",
			"resource_kind": "button",
			"expected_type": "BaseButton",
			"actual_type": "null",
		}
	)


func _on_back_pressed() -> void:
	back_requested.emit()


func _on_discord_login_pressed() -> void:
	discord_login_requested.emit()
