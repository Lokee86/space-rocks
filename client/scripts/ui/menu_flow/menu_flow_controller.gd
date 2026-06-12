class_name MenuFlowController
extends RefCounted

const PregameMenuScene := preload("res://scenes/ui/pregame_menu.tscn")
const MenuRoute := preload("res://scripts/ui/menu_flow/menu_route.gd")
const PregameMenuFlow := preload("res://scripts/ui/menu_flow/pregame_menu_flow.gd")

var canvas_layer: CanvasLayer
var main_menu: Control
var pregame_menu: Control
var pregame_menu_flow
var start_single_player_callable: Callable
var current_route := ""


func configure(canvas_layer_ref: CanvasLayer, main_menu_ref: Control, start_single_player_callable_ref: Callable = Callable()) -> void:
	canvas_layer = canvas_layer_ref
	main_menu = main_menu_ref
	start_single_player_callable = start_single_player_callable_ref
	current_route = MenuRoute.MAIN_MENU
	if main_menu != null:
		main_menu.show()


func show_main_menu() -> void:
	if pregame_menu != null and is_instance_valid(pregame_menu):
		pregame_menu.queue_free()
	pregame_menu = null
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


func clear_for_gameplay() -> void:
	if pregame_menu != null and is_instance_valid(pregame_menu):
		pregame_menu.queue_free()
	pregame_menu = null
	pregame_menu_flow = null
	if main_menu != null:
		main_menu.hide()
	current_route = ""


func _show_pregame() -> void:
	if main_menu != null:
		main_menu.hide()

	if pregame_menu == null or not is_instance_valid(pregame_menu):
		pregame_menu = PregameMenuScene.instantiate()
		if canvas_layer != null:
			canvas_layer.add_child(pregame_menu)
		pregame_menu_flow = PregameMenuFlow.new()

	if pregame_menu_flow == null:
		pregame_menu_flow = PregameMenuFlow.new()

	pregame_menu_flow.configure(pregame_menu, Callable(self, "show_main_menu"), start_single_player_callable)
	current_route = MenuRoute.PREGAME_MENU
	if pregame_menu != null:
		pregame_menu.show()


func get_current_route() -> String:
	return current_route


func get_pregame_menu() -> Control:
	return pregame_menu
