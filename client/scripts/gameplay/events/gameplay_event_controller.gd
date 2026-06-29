extends RefCounted

const Packets = preload("res://scripts/generated/networking/packets/packets.gd")

var effects
var visual_position_for_server_position: Callable
var _logged_ship_death_diagnostics := false


func configure(effects_object, visual_position_converter: Callable) -> void:
	effects = effects_object
	visual_position_for_server_position = visual_position_converter


func apply_server_events(server_events: Array, self_id: String, apply_self_death_event: Callable) -> void:
	for event in server_events:
		if event.get(Packets.FIELD_TYPE, "") == Packets.TYPE_BULLET_BLAST:
			apply_bullet_blast(event)
		elif event.get(Packets.FIELD_TYPE, "") == Packets.TYPE_SHIP_DEATH:
			var event_player_id := str(event.get(Packets.FIELD_PLAYER_ID, ""))
			var self_id_string := str(self_id)
			var is_self_death := event_player_id == self_id_string
			if !_logged_ship_death_diagnostics:
				print(
					"Gameplay ship_death diagnostics: event.player_id=%s self_id=%s is_self_death=%s lives=%s respawn_delay=%s" % [
						event_player_id,
						self_id_string,
						str(is_self_death),
						str(event.get(Packets.FIELD_LIVES, "")),
						str(event.get(Packets.FIELD_RESPAWN_DELAY, ""))
					]
				)
				_logged_ship_death_diagnostics = true
			if is_self_death:
				apply_self_death_event.call(event)
			apply_ship_death(event)
		elif event.get(Packets.FIELD_TYPE, "") == "radial_effect_started":
			apply_radial_effect_started(event)
		elif event.get(Packets.FIELD_TYPE, "") == "pickup_collected":
			apply_pickup_collected(event)
		elif event.get(Packets.FIELD_TYPE, "") == "pickup_effect_applied":
			pass


func apply_bullet_blast(event: Dictionary) -> void:
	var visual_position: Vector2 = _visual_position_for_event(event, "bullet blast")
	if visual_position == null:
		return
	effects.spawn_bullet_blast(visual_position)


func apply_ship_death(event: Dictionary) -> void:
	var visual_position: Vector2 = _visual_position_for_event(event, "ship death")
	if visual_position == null:
		return
	effects.spawn_ship_death(visual_position)


func apply_pickup_collected(event: Dictionary) -> void:
	var visual_position: Vector2 = _visual_position_for_event(event, "pickup collected")
	if visual_position == null:
		return
	effects.spawn_pickup_collected(visual_position)


func apply_radial_effect_started(event: Dictionary) -> void:
	var visual_position: Vector2 = _visual_position_for_event(event, "radial effect started")
	if visual_position == null:
		return
	effects.spawn_torpedo_explosion(visual_position)


func _visual_position_for_event(event: Dictionary, event_name: String):
	if visual_position_for_server_position.is_null():
		push_warning("Cannot convert %s event position without a visual position converter." % event_name)
		return null

	var event_x = event.get(Packets.FIELD_X)
	var event_y = event.get(Packets.FIELD_Y)
	if event_x is Callable || event_y is Callable:
		push_warning("Cannot convert %s event position because event coordinates contain a Callable." % event_name)
		return null

	var event_position := Vector2(event_x, event_y)
	return visual_position_for_server_position.call(event_position)

