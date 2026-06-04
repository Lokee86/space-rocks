extends RefCounted

const Packets = preload("res://scripts/networking/packets/packets.gd")

var effects
var visual_position_for_server_position: Callable


func configure(effects_object, visual_position_converter: Callable) -> void:
	effects = effects_object
	visual_position_for_server_position = visual_position_converter


func apply_server_events(server_events: Array, self_id: String, apply_self_death_event: Callable) -> void:
	for event in server_events:
		if event.get(Packets.FIELD_TYPE, "") == Packets.TYPE_BULLET_BLAST:
			apply_bullet_blast(event)
		elif event.get(Packets.FIELD_TYPE, "") == Packets.TYPE_SHIP_DEATH:
			if event[Packets.FIELD_PLAYER_ID] == self_id:
				apply_self_death_event.call(event)
			apply_ship_death(event)


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
