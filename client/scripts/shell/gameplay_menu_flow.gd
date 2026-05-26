extends RefCounted
class_name GameplayMenuFlow

var hud: Control
var game_menu


func configure(hud_ref: Control) -> void:
	hud = hud_ref
	if hud != null:
		game_menu = hud.get_node_or_null("GameMenu")


func hide_menu() -> void:
	if game_menu != null:
		game_menu.hide()


func show_menu() -> void:
	if game_menu != null:
		game_menu.show()


func is_menu_visible() -> bool:
	return game_menu != null && game_menu.visible
