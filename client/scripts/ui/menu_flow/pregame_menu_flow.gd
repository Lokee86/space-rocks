class_name PregameMenuFlow
extends RefCounted

const PregameMenuMode := preload("res://scripts/ui/menu_flow/pregame_menu_mode.gd")

var pregame_menu: Control
var return_to_main_menu: Callable
var start_single_player_callable: Callable
var current_mode := ""


func configure(pregame_menu_ref: Control, return_to_main_menu_callable: Callable, start_single_player_callable_ref: Callable = Callable()) -> void:
	pregame_menu = pregame_menu_ref
	return_to_main_menu = return_to_main_menu_callable
	start_single_player_callable = start_single_player_callable_ref

	if pregame_menu != null and pregame_menu.has_signal("back_requested"):
		if not pregame_menu.back_requested.is_connected(_on_back_requested):
			pregame_menu.back_requested.connect(_on_back_requested)
	if pregame_menu != null and pregame_menu.has_signal("play_endless_requested"):
		if not pregame_menu.play_endless_requested.is_connected(_on_play_endless_requested):
			pregame_menu.play_endless_requested.connect(_on_play_endless_requested)


func show_single_player() -> void:
	current_mode = PregameMenuMode.SINGLE_PLAYER
	if pregame_menu != null and pregame_menu.has_method("show_single_player_mode"):
		pregame_menu.show_single_player_mode()


func show_multiplayer() -> void:
	current_mode = PregameMenuMode.MULTIPLAYER
	if pregame_menu != null and pregame_menu.has_method("show_multiplayer_mode"):
		pregame_menu.show_multiplayer_mode()


func _on_back_requested() -> void:
	if return_to_main_menu.is_valid():
		return_to_main_menu.call()


func _on_play_endless_requested() -> void:
	if current_mode != PregameMenuMode.SINGLE_PLAYER:
		return
	if start_single_player_callable.is_valid():
		start_single_player_callable.call()
