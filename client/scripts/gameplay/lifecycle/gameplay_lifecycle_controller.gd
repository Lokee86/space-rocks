extends RefCounted

const Packets = preload("res://scripts/networking/packets.gd")
const RESPAWN_RETRY_SECONDS := 0.25

var hud_controller
var effects
var player
var stop_spectating: Callable
var resume_gameplay_pause: Callable
var hide_game_menu: Callable
var show_game_menu: Callable
var refresh_cycle_view_hint: Callable
var disarm_open_menu_input: Callable
var transition_to_dead: Callable
var transition_to_game_over: Callable
var respawn_retry_remaining := 0.0
var awaiting_respawn_confirmation := false


func configure(
	hud_controller_object,
	effects_object,
	player_node,
	stop_spectating_callback: Callable,
	resume_gameplay_pause_callback: Callable,
	hide_game_menu_callback: Callable,
	show_game_menu_callback: Callable,
	refresh_cycle_view_hint_callback: Callable,
	disarm_open_menu_input_callback: Callable,
	set_dead_state_callback: Callable,
	set_game_over_state_callback: Callable
) -> void:
	hud_controller = hud_controller_object
	effects = effects_object
	player = player_node
	stop_spectating = stop_spectating_callback
	resume_gameplay_pause = resume_gameplay_pause_callback
	hide_game_menu = hide_game_menu_callback
	show_game_menu = show_game_menu_callback
	refresh_cycle_view_hint = refresh_cycle_view_hint_callback
	disarm_open_menu_input = disarm_open_menu_input_callback
	transition_to_dead = set_dead_state_callback
	transition_to_game_over = set_game_over_state_callback


func tick_respawn_retry(delta: float) -> void:
	respawn_retry_remaining = max(0.0, respawn_retry_remaining - delta)


func clear_awaiting_respawn_confirmation() -> void:
	awaiting_respawn_confirmation = false


func is_awaiting_respawn_confirmation() -> bool:
	return awaiting_respawn_confirmation


func set_awaiting_respawn_confirmation(value: bool) -> void:
	awaiting_respawn_confirmation = value


func can_request_respawn() -> bool:
	return respawn_retry_remaining <= 0.0


func mark_respawn_requested() -> void:
	respawn_retry_remaining = RESPAWN_RETRY_SECONDS
	awaiting_respawn_confirmation = true


func send_respawn_request_if_ready(network_client) -> void:
	if hud_controller.can_respawn && Input.is_key_pressed(KEY_R) && can_request_respawn():
		mark_respawn_requested()
		network_client.send_packet(Packets.respawn_packet())


func apply_self_death_event(event: Dictionary) -> void:
	var lives := int(event.get(Packets.FIELD_LIVES, 0))
	hud_controller.set_lives(lives)
	if lives <= 0:
		transition_to_game_over.call()
		return

	if event.has(Packets.FIELD_RESPAWN_DELAY):
		transition_to_dead.call(float(event[Packets.FIELD_RESPAWN_DELAY]))
	else:
		push_warning("Ship death event missing respawn delay")
		transition_to_dead.call(0.0)


func set_alive_state() -> void:
	clear_awaiting_respawn_confirmation()
	stop_spectating.call(false)
	resume_gameplay_pause.call()
	disarm_open_menu_input.call()
	hide_game_menu.call()
	hud_controller.set_alive()
	refresh_cycle_view_hint.call()
	effects.reset_game_over_sound()


func set_dead_state(respawn_delay: float) -> void:
	clear_awaiting_respawn_confirmation()
	stop_spectating.call(false)
	resume_gameplay_pause.call()
	disarm_open_menu_input.call()
	hide_game_menu.call()
	player.set_afterburner_active(false)
	hud_controller.set_dead(respawn_delay)
	refresh_cycle_view_hint.call()
	effects.stop_game_over_sound()


func set_game_over_state() -> void:
	clear_awaiting_respawn_confirmation()
	resume_gameplay_pause.call()
	disarm_open_menu_input.call()
	hide_game_menu.call()
	player.set_afterburner_active(false)
	hud_controller.set_game_over()
	show_game_menu.call()
	refresh_cycle_view_hint.call()
	effects.play_game_over_sound_after_delay()
