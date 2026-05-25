extends RefCounted

const Packets = preload("res://scripts/networking/packets.gd")

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
	var event_position := Vector2(event[Packets.FIELD_X], event[Packets.FIELD_Y])
	effects.spawn_bullet_blast(visual_position_for_server_position.call(event_position))


func apply_ship_death(event: Dictionary) -> void:
	var event_position := Vector2(event[Packets.FIELD_X], event[Packets.FIELD_Y])
	effects.spawn_ship_death(visual_position_for_server_position.call(event_position))
