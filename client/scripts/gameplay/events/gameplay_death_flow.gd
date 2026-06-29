extends RefCounted
class_name GameplayDeathFlow

const Packets = preload("res://scripts/generated/networking/packets/packets.gd")

var hud_flow
var match_end_flow
var player


func configure(hud_flow_ref, match_end_flow_ref, player_ref = null) -> void:
	hud_flow = hud_flow_ref
	match_end_flow = match_end_flow_ref
	player = player_ref


func apply_self_death_event(event: Dictionary) -> void:
	if player != null:
		player.stop_transient_effects()

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

	if match_end_flow != null && match_end_flow.has_method("handle_local_player_eliminated"):
		match_end_flow.handle_local_player_eliminated(event)


