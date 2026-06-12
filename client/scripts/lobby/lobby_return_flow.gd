extends RefCounted

var lobby_flow
var multiplayer_lobby_presenter
var main_menu: Control
var return_to_menu_callback: Callable
var return_destination_callback: Callable


func _init(
	lobby_flow_ref,
	multiplayer_lobby_presenter_ref,
	main_menu_ref: Control,
	return_to_menu_callable: Callable
) -> void:
	lobby_flow = lobby_flow_ref
	multiplayer_lobby_presenter = multiplayer_lobby_presenter_ref
	main_menu = main_menu_ref
	return_to_menu_callback = return_to_menu_callable


func configure_return_destination(callback: Callable) -> void:
	return_destination_callback = callback


func return_to_main_menu() -> void:
	return_after_leave()


func return_after_leave() -> void:
	if lobby_flow != null:
		lobby_flow.clear()
	multiplayer_lobby_presenter.clear_lobby()
	if !return_to_menu_callback.is_null():
		return_to_menu_callback.call()
	if !return_destination_callback.is_null() and return_destination_callback.is_valid():
		return_destination_callback.call()
	elif main_menu != null:
		main_menu.show()
