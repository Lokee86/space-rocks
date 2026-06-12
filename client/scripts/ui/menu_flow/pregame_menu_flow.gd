class_name PregameMenuFlow
extends RefCounted

const PregameMenuMode := preload("res://scripts/ui/menu_flow/pregame_menu_mode.gd")

var pregame_menu: Control
var return_to_main_menu: Callable
var current_mode := ""


func configure(pregame_menu_ref: Control, return_to_main_menu_callable: Callable) -> void:
	pregame_menu = pregame_menu_ref
	return_to_main_menu = return_to_main_menu_callable

	if pregame_menu != null and pregame_menu.has_signal("back_requested"):
		if not pregame_menu.back_requested.is_connected(_on_back_requested):
			pregame_menu.back_requested.connect(_on_back_requested)


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
