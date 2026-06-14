class_name PregameMenuFlow
extends RefCounted

const PregameMenuMode := preload("res://scripts/ui/menu_flow/pregame_menu_mode.gd")
const LocalPilotFlow := preload("res://scripts/ui/menu_flow/local_pilot_flow.gd")

var pregame_menu: Control
var return_to_main_menu: Callable
var start_single_player_callable: Callable
var create_room_callable: Callable
var show_join_dialog_callable: Callable
var logout_callable: Callable
var clear_for_room_transition_callable: Callable
var profile_context_provider
var profile_flow
var transmission_flow
var local_pilot_flow
var current_mode := ""


func configure(
		pregame_menu_ref: Control,
		return_to_main_menu_callable: Callable,
		start_single_player_callable_ref: Callable = Callable(),
		create_room_callable_ref: Callable = Callable(),
		show_join_dialog_callable_ref: Callable = Callable(),
		logout_callable_ref: Callable = Callable(),
		clear_for_room_transition_callable_ref: Callable = Callable(),
		profile_context_provider_ref = null,
		profile_flow_ref = null,
		transmission_flow_ref = null) -> void:
	pregame_menu = pregame_menu_ref
	return_to_main_menu = return_to_main_menu_callable
	start_single_player_callable = start_single_player_callable_ref
	create_room_callable = create_room_callable_ref
	show_join_dialog_callable = show_join_dialog_callable_ref
	logout_callable = logout_callable_ref
	clear_for_room_transition_callable = clear_for_room_transition_callable_ref
	profile_context_provider = profile_context_provider_ref
	profile_flow = profile_flow_ref
	transmission_flow = transmission_flow_ref
	local_pilot_flow = LocalPilotFlow.new()
	local_pilot_flow.configure(transmission_flow, Callable(pregame_menu, "set_callsign"))

	if pregame_menu != null and pregame_menu.has_signal("back_requested"):
		if not pregame_menu.back_requested.is_connected(_on_back_requested):
			pregame_menu.back_requested.connect(_on_back_requested)
	if pregame_menu != null and pregame_menu.has_signal("play_endless_requested"):
		if not pregame_menu.play_endless_requested.is_connected(_on_play_endless_requested):
			pregame_menu.play_endless_requested.connect(_on_play_endless_requested)
	if pregame_menu != null and pregame_menu.has_signal("create_game_requested"):
		if not pregame_menu.create_game_requested.is_connected(_on_create_game_requested):
			pregame_menu.create_game_requested.connect(_on_create_game_requested)
	if pregame_menu != null and pregame_menu.has_signal("join_game_requested"):
		if not pregame_menu.join_game_requested.is_connected(_on_join_game_requested):
			pregame_menu.join_game_requested.connect(_on_join_game_requested)
	if pregame_menu != null and pregame_menu.has_signal("logout_requested"):
		if not pregame_menu.logout_requested.is_connected(_on_logout_requested):
			pregame_menu.logout_requested.connect(_on_logout_requested)
	if pregame_menu != null and pregame_menu.has_signal("profile_requested"):
		if not pregame_menu.profile_requested.is_connected(_on_profile_requested):
			pregame_menu.profile_requested.connect(_on_profile_requested)
	if pregame_menu != null and pregame_menu.has_signal("select_pilot_requested"):
		if not pregame_menu.select_pilot_requested.is_connected(_on_select_pilot_requested):
			pregame_menu.select_pilot_requested.connect(_on_select_pilot_requested)


func show_single_player() -> void:
	current_mode = PregameMenuMode.SINGLE_PLAYER
	if pregame_menu != null and pregame_menu.has_method("show_single_player_mode"):
		pregame_menu.show_single_player_mode()
	_update_callsign_indicator()


func show_multiplayer() -> void:
	current_mode = PregameMenuMode.MULTIPLAYER
	if pregame_menu != null and pregame_menu.has_method("show_multiplayer_mode"):
		pregame_menu.show_multiplayer_mode()
	_update_callsign_indicator()


func _on_back_requested() -> void:
	if transmission_flow != null and transmission_flow.has_active_transmission():
		transmission_flow.clear()
		return
	if return_to_main_menu.is_valid():
		return_to_main_menu.call()


func _on_play_endless_requested() -> void:
	if current_mode != PregameMenuMode.SINGLE_PLAYER:
		return
	if start_single_player_callable.is_valid():
		start_single_player_callable.call()


func _on_create_game_requested() -> void:
	if current_mode != PregameMenuMode.MULTIPLAYER:
		return
	if clear_for_room_transition_callable.is_valid():
		clear_for_room_transition_callable.call()
	if create_room_callable.is_valid():
		create_room_callable.call()


func _on_join_game_requested() -> void:
	if current_mode != PregameMenuMode.MULTIPLAYER:
		return
	if show_join_dialog_callable.is_valid():
		show_join_dialog_callable.call()


func _on_logout_requested() -> void:
	if current_mode != PregameMenuMode.MULTIPLAYER:
		return
	if logout_callable.is_valid():
		logout_callable.call()
	if return_to_main_menu.is_valid():
		return_to_main_menu.call()


func _on_profile_requested() -> void:
	if profile_flow != null and profile_flow.has_method("show_profile"):
		await profile_flow.show_profile(current_mode)


func _on_select_pilot_requested() -> void:
	if current_mode != PregameMenuMode.SINGLE_PLAYER:
		return
	if local_pilot_flow != null and local_pilot_flow.has_method("show_selector"):
		local_pilot_flow.show_selector()


func _update_callsign_indicator() -> void:
	if pregame_menu == null or !pregame_menu.has_method("set_callsign"):
		return
	if profile_context_provider == null or !profile_context_provider.has_method("context_for_mode"):
		return

	var context: Dictionary = profile_context_provider.context_for_mode(current_mode)
	pregame_menu.set_callsign(str(context.get("callsign", "Guest")))
