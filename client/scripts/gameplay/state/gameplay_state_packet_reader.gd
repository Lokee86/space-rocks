extends RefCounted

const Packets = preload("res://scripts/generated/networking/packets/packets.gd")
const PlayerLifecycle = preload("res://scripts/gameplay/lifecycle/player_lifecycle.gd")
const FIELD_DEBUG_STATUS := "debug_status"
const FIELD_DEBUG_STATUSES := "debug_statuses"
const FIELD_SERVER_SENT_MSEC := "server_sent_msec"
const FIELD_PLAYER_SESSIONS := "player_sessions"
const FIELD_DEBUG_COLLISION_BODIES := "debug_collision_bodies"


static func read(data: Dictionary) -> Dictionary:
	var server_events: Array = []
	var events_data = data.get(Packets.FIELD_EVENTS, [])
	if events_data is Array:
		server_events = events_data

	var has_lives := data.has(Packets.FIELD_LIVES)
	var lives := 0
	if has_lives:
		lives = int(data[Packets.FIELD_LIVES])

	var debug_status = data.get(FIELD_DEBUG_STATUS, {})
	if !(debug_status is Dictionary):
		debug_status = {}

	var debug_statuses = data.get(FIELD_DEBUG_STATUSES, {})
	if !(debug_statuses is Dictionary):
		debug_statuses = {}
	var player_sessions := {}
	var player_sessions_value = data.get(FIELD_PLAYER_SESSIONS, {})
	if player_sessions_value is Dictionary:
		player_sessions = player_sessions_value

	var server_pickups_value = data.get(Packets.FIELD_PICKUPS, {})
	var server_pickups := {}
	if server_pickups_value is Dictionary:
		server_pickups = server_pickups_value

	var debug_collision_bodies: Variant = data.get(FIELD_DEBUG_COLLISION_BODIES, [])
	if !(debug_collision_bodies is Array):
		debug_collision_bodies = []

	return {
		"self_id": data[Packets.FIELD_SELF_ID],
		"server_players": data[Packets.FIELD_PLAYERS],
		"player_sessions": player_sessions,
		"player_lifecycle": PlayerLifecycle.from_state(data),
		"server_bullets": data.get(Packets.FIELD_BULLETS, {}),
		"server_asteroids": data.get(Packets.FIELD_ASTEROIDS, {}),
		"server_pickups": server_pickups,
		"total_asteroids": int(data.get(Packets.FIELD_TOTAL_ASTEROIDS, 0)),
		"server_events": server_events,
		"server_sent_msec": int(data.get(FIELD_SERVER_SENT_MSEC, -1)),
		"debug_status": debug_status,
		"debug_statuses": debug_statuses,
		"debug_collision_bodies": debug_collision_bodies,
		"has_lives": has_lives,
		"lives": lives,
	}
