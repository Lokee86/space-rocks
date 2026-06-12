extends RefCounted

var login_window: Control
var show_main_menu_callable: Callable
var request_discord_sign_in_callable: Callable


func configure(login_window_ref: Control, show_main_menu_callable_ref: Callable, request_discord_sign_in_callable_ref: Callable) -> void:
	login_window = login_window_ref
	show_main_menu_callable = show_main_menu_callable_ref
	request_discord_sign_in_callable = request_discord_sign_in_callable_ref

	if login_window == null:
		return

	if login_window.has_signal("back_requested") and not login_window.is_connected("back_requested", Callable(self, "_on_back_requested")):
		login_window.connect("back_requested", Callable(self, "_on_back_requested"))

	if login_window.has_signal("discord_login_requested") and not login_window.is_connected("discord_login_requested", Callable(self, "_on_discord_login_requested")):
		login_window.connect("discord_login_requested", Callable(self, "_on_discord_login_requested"))


func _on_back_requested() -> void:
	if show_main_menu_callable.is_valid():
		show_main_menu_callable.call()


func _on_discord_login_requested() -> void:
	if request_discord_sign_in_callable.is_valid():
		request_discord_sign_in_callable.call()
