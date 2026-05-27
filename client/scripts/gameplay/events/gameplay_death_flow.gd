extends RefCounted
class_name GameplayDeathFlow

const Packets = preload("res://scripts/networking/packets/packets.gd")

var hud_flow
var menu_flow
var event_flow


func configure(hud_flow_ref, menu_flow_ref, event_flow_ref) -> void:
	hud_flow = hud_flow_ref
	menu_flow = menu_flow_ref
	event_flow = event_flow_ref


func apply_self_death_event(event: Dictionary) -> void:
	var lives := int(event.get(Packets.FIELD_LIVES, 0))
	if hud_flow != null:
		hud_flow.apply_lives(lives)
	if lives > 0:
		var respawn_delay := 0.0
		if event.has(Packets.FIELD_RESPAWN_DELAY):
			respawn_delay = float(event[Packets.FIELD_RESPAWN_DELAY])
		if hud_flow != null:
			hud_flow.set_dead(respawn_delay)
		return

	if hud_flow != null:
		hud_flow.set_game_over()
	if menu_flow != null:
		menu_flow.set_game_over()
	if event_flow != null && event_flow.has_method("play_game_over_sound_after_delay"):
		event_flow.play_game_over_sound_after_delay()
