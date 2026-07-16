extends RefCounted


var kill_player_route: Callable


func configure(kill_player_route_ref: Callable = Callable()) -> void:
	kill_player_route = kill_player_route_ref


func configure_kill_player_route(route: Callable) -> void:
	kill_player_route = route


func process() -> void:
	if Input.is_action_just_pressed("DevToggle5") && !kill_player_route.is_null():
		kill_player_route.call("", "")
