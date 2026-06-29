extends RefCounted

var player_sessions := {}
var player_lifecycle := {}
var total_asteroids = null

func clear_session() -> void:
	player_sessions.clear()
	player_lifecycle.clear()
	total_asteroids = null

func apply_full_session(session_packet: Dictionary) -> void:
	clear_session()
	_apply_session_fields(session_packet)
	_replace_records(player_sessions, session_packet.get("player_sessions", []))
	_replace_records(player_lifecycle, session_packet.get("player_lifecycle", []))

func apply_session_delta(session_packet: Dictionary) -> void:
	_apply_session_fields(session_packet)
	_apply_creates(player_sessions, session_packet.get("player_session_creates", []))
	_apply_updates(player_sessions, session_packet.get("player_session_updates", []))
	_apply_deletes(player_sessions, session_packet.get("player_session_deletes", []))
	_apply_creates(player_lifecycle, session_packet.get("player_lifecycle_creates", []))
	_apply_updates(player_lifecycle, session_packet.get("player_lifecycle_updates", []))
	_apply_deletes(player_lifecycle, session_packet.get("player_lifecycle_deletes", []))

func _apply_session_fields(session_packet: Dictionary) -> void:
	if session_packet.has("total_asteroids"):
		total_asteroids = session_packet.get("total_asteroids")

func _replace_records(target: Dictionary, records: Array) -> void:
	target.clear()
	_apply_creates(target, records)

func _apply_creates(target: Dictionary, records: Array) -> void:
	for record in records:
		_apply_upsert(target, record)

func _apply_updates(target: Dictionary, records: Array) -> void:
	for record in records:
		_apply_upsert(target, record)

func _apply_deletes(target: Dictionary, records: Array) -> void:
	for record in records:
		var id = record.get("id")
		if id != null:
			target.erase(id)

func _apply_upsert(target: Dictionary, record: Dictionary) -> void:
	var id = record.get("id")
	if id == null:
		return
	target[id] = record.duplicate(true)

