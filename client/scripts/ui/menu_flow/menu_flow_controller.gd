class_name MenuFlowController
extends RefCounted

const PregameMenuScene := preload("res://scenes/ui/pregame_menu.tscn")
const LoginWindowScene := preload("res://scenes/ui/dialogs/login_window.tscn")
const JoinDialogScene := preload("res://scenes/ui/dialogs/join_dialog.tscn")
const MenuRoute := preload("res://scripts/ui/menu_flow/menu_route.gd")
const PregameMenuMode := preload("res://scripts/ui/menu_flow/pregame_menu_mode.gd")
const PregameMenuFlow := preload("res://scripts/ui/menu_flow/pregame_menu_flow.gd")
const TransmissionFlow := preload("res://scripts/ui/menu_flow/transmission_flow.gd")
const ProfileContextProvider := preload("res://scripts/profile/profile_context_provider.gd")
const ProfileStatsProvider := preload("res://scripts/profile/profile_stats_provider.gd")
const ProfileFlow := preload("res://scripts/profile/profile_flow.gd")
const JoinDialogFlow := preload("res://scripts/ui/lobby/join_dialog_flow.gd")
const SignInFlow := preload("res://scripts/ui/sign_in/sign_in_flow.gd")

var canvas_layer: CanvasLayer
var main_menu: Control
var pregame_menu: Control
var sign_in_screen: Control
var join_dialog: Control
var pregame_menu_flow
var join_dialog_flow
var sign_in_flow
var auth_session_controller
var profile_context_provider
var profile_stats_provider
var profile_stats_provider_is_shared := false
var start_single_player_callable: Callable
var request_discord_sign_in_callable: Callable
var create_room_callable: Callable
var join_room_callable: Callable
var logout_callable: Callable
var current_route := ""


func configure(
		canvas_layer_ref: CanvasLayer,
		main_menu_ref: Control,
		start_single_player_callable_ref: Callable = Callable(),
		request_discord_sign_in_callable_ref: Callable = Callable(),
		create_room_callable_ref: Callable = Callable(),
		join_room_callable_ref: Callable = Callable(),
		logout_callable_ref: Callable = Callable(),
		auth_session_controller_ref = null,
		profile_stats_provider_ref = null) -> void:
	canvas_layer = canvas_layer_ref
	main_menu = main_menu_ref
	start_single_player_callable = start_single_player_callable_ref
	request_discord_sign_in_callable = request_discord_sign_in_callable_ref
	create_room_callable = create_room_callable_ref
	join_room_callable = join_room_callable_ref
	logout_callable = logout_callable_ref
	auth_session_controller = auth_session_controller_ref
	profile_stats_provider_is_shared = profile_stats_provider_ref != null
	if profile_context_provider == null:
		profile_context_provider = ProfileContextProvider.new()
	if profile_context_provider != null and profile_context_provider.has_method("configure"):
		profile_context_provider.configure(auth_session_controller)
	if profile_stats_provider_ref != null:
		profile_stats_provider = profile_stats_provider_ref
	elif profile_stats_provider == null:
		profile_stats_provider = ProfileStatsProvider.new()
		if profile_stats_provider != null and profile_stats_provider.has_method("configure"):
			profile_stats_provider.configure(auth_session_controller)
	elif !profile_stats_provider_is_shared and profile_stats_provider != null and profile_stats_provider.has_method("configure"):
		profile_stats_provider.configure(auth_session_controller)
	current_route = MenuRoute.MAIN_MENU
	if main_menu != null:
		main_menu.show()


func show_main_menu() -> void:
	_clear_pregame_menu()
	_clear_sign_in_screen()
	_clear_join_dialog()
	pregame_menu_flow = null
	if main_menu != null:
		main_menu.show()
	current_route = MenuRoute.MAIN_MENU


func show_single_player_pregame() -> void:
	_show_pregame()
	if pregame_menu_flow != null:
		pregame_menu_flow.show_single_player()


func show_multiplayer_pregame() -> void:
	_show_pregame()
	if pregame_menu_flow != null:
		pregame_menu_flow.show_multiplayer()


func show_sign_in_screen() -> void:
	_clear_pregame_menu()
	if main_menu != null:
		main_menu.hide()

	if sign_in_screen == null or not is_instance_valid(sign_in_screen):
		sign_in_screen = LoginWindowScene.instantiate()
		if canvas_layer != null:
			canvas_layer.add_child(sign_in_screen)

	sign_in_flow = SignInFlow.new()
	if sign_in_flow != null:
		sign_in_flow.configure(sign_in_screen, Callable(self, "show_main_menu"), request_discord_sign_in_callable)
	current_route = MenuRoute.SIGN_IN_SCREEN
	if sign_in_screen != null:
		sign_in_screen.show()


func show_join_dialog() -> void:
	if canvas_layer == null:
		return

	_clear_sign_in_screen()
	if main_menu != null:
		main_menu.hide()

	if join_dialog == null or not is_instance_valid(join_dialog):
		join_dialog = JoinDialogScene.instantiate()
		canvas_layer.add_child(join_dialog)

	if join_dialog_flow == null:
		join_dialog_flow = JoinDialogFlow.new()

	join_dialog_flow.configure(
		join_dialog,
		Callable(self, "close_join_dialog"),
		join_room_callable,
		Callable(self, "clear_for_room_transition"))
	current_route = MenuRoute.JOIN_DIALOG
	if join_dialog != null:
		join_dialog.show()


func close_join_dialog() -> void:
	_clear_join_dialog()
	if pregame_menu != null and is_instance_valid(pregame_menu):
		current_route = MenuRoute.PREGAME_MENU
		pregame_menu.show()
	elif main_menu != null:
		current_route = MenuRoute.MAIN_MENU
		main_menu.show()


func clear_for_gameplay() -> void:
	_clear_pregame_menu()
	_clear_sign_in_screen()
	_clear_join_dialog()
	pregame_menu_flow = null
	if main_menu != null:
		main_menu.hide()
	current_route = ""


func clear_for_room_transition() -> void:
	_clear_pregame_menu()
	_clear_sign_in_screen()
	_clear_join_dialog()
	if main_menu != null:
		main_menu.hide()
	current_route = ""


func _show_pregame() -> void:
	_clear_sign_in_screen()
	_clear_join_dialog()
	if main_menu != null:
		main_menu.hide()

	if pregame_menu == null or not is_instance_valid(pregame_menu):
		pregame_menu = PregameMenuScene.instantiate()
		if canvas_layer != null:
			canvas_layer.add_child(pregame_menu)
		pregame_menu_flow = PregameMenuFlow.new()

	if pregame_menu_flow == null:
		pregame_menu_flow = PregameMenuFlow.new()

	var transmission_flow := TransmissionFlow.new()
	transmission_flow.configure(pregame_menu)

	if profile_context_provider == null:
		profile_context_provider = ProfileContextProvider.new()
	if profile_context_provider != null and profile_context_provider.has_method("configure"):
		profile_context_provider.configure(auth_session_controller)

	var profile_flow := ProfileFlow.new()
	profile_flow.configure(profile_context_provider, profile_stats_provider, transmission_flow)

	pregame_menu_flow.configure(
		pregame_menu,
		Callable(self, "show_main_menu"),
		start_single_player_callable,
		create_room_callable,
		Callable(self, "show_join_dialog"),
		logout_callable,
		Callable(self, "clear_for_room_transition"),
		profile_context_provider,
		profile_flow,
		transmission_flow)
	current_route = MenuRoute.PREGAME_MENU
	if pregame_menu != null:
		pregame_menu.show()


func _clear_pregame_menu() -> void:
	if pregame_menu != null and is_instance_valid(pregame_menu):
		pregame_menu.queue_free()
	pregame_menu = null
	pregame_menu_flow = null


func _clear_sign_in_screen() -> void:
	if sign_in_screen != null and is_instance_valid(sign_in_screen):
		sign_in_screen.queue_free()
	sign_in_screen = null
	sign_in_flow = null


func _clear_join_dialog() -> void:
	if join_dialog != null and is_instance_valid(join_dialog):
		join_dialog.queue_free()
	join_dialog = null
	join_dialog_flow = null


func get_current_route() -> String:
	return current_route


func get_pregame_menu() -> Control:
	return pregame_menu


func get_sign_in_screen() -> Control:
	return sign_in_screen


func get_join_dialog() -> Control:
	return join_dialog


func get_single_player_profile_context() -> Dictionary:
	if pregame_menu_flow != null and pregame_menu_flow.has_method("get_single_player_profile_context"):
		return pregame_menu_flow.get_single_player_profile_context()

	if profile_context_provider != null and profile_context_provider.has_method("context_for_mode"):
		return profile_context_provider.context_for_mode(PregameMenuMode.SINGLE_PLAYER)

	return {
		"play_mode": PregameMenuMode.SINGLE_PLAYER,
		"identity_kind": "guest",
		"callsign": "Guest",
	}


func get_single_player_context() -> Dictionary:
	return get_single_player_profile_context()
