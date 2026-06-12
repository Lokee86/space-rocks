extends Control

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
		push_error("Missing input: EmailInput")

	if password_input != null:
		password_input.editable = false
	else:
		push_error("Missing input: PasswordInput")

	if sign_in_button != null:
		sign_in_button.disabled = true
	else:
		push_error("Missing button: SignInButton")

	if google_login_button != null:
		google_login_button.disabled = true
	else:
		push_error("Missing button: GoogleLoginButton")

	if back_button != null:
		back_button.pressed.connect(_on_back_pressed)
	else:
		push_error("Missing button: BackButton")

	if discord_login_button != null:
		discord_login_button.pressed.connect(_on_discord_login_pressed)
	else:
		push_error("Missing button: DiscordLoginButton")


func _on_back_pressed() -> void:
	back_requested.emit()


func _on_discord_login_pressed() -> void:
	discord_login_requested.emit()
