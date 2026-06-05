extends RefCounted

var players: int = 0
var enemies: int = 0
var asteroids: int = 0
var pickups: int = 0
var total_asteroids: int = 0
var bullets: int = 0
var server_sent_msec: int = -1
var server_clock_offset_ms: int = -1
var latest_packet_arrival_msec: int = -1
var previous_packet_arrival_msec: int = -1
var packet_interval_ms: int = -1
var previous_packet_interval_ms: int = -1
var jitter_ms: int = -1


func reset() -> void:
	players = 0
	enemies = 0
	asteroids = 0
	pickups = 0
	total_asteroids = 0
	bullets = 0
	server_sent_msec = -1
	server_clock_offset_ms = -1
	latest_packet_arrival_msec = -1
	previous_packet_arrival_msec = -1
	packet_interval_ms = -1
	previous_packet_interval_ms = -1
	jitter_ms = -1


func snapshot() -> Dictionary:
	var packet_staleness_ms: int = Time.get_ticks_msec() - latest_packet_arrival_msec if latest_packet_arrival_msec >= 0 else -1
	var packet_age_ms: int = -1
	if server_sent_msec > 0 and server_clock_offset_ms >= 0:
		var estimated_client_sent_msec := server_sent_msec - server_clock_offset_ms
		packet_age_ms = max(Time.get_ticks_msec() - estimated_client_sent_msec, 0)
	return {
		"players": players,
		"enemies": enemies,
		"asteroids": asteroids,
		"pickups": pickups,
		"total_asteroids": total_asteroids,
		"bullets": bullets,
		"server_sent_msec": server_sent_msec,
		"server_clock_offset_ms": server_clock_offset_ms,
		"packet_interval_ms": packet_interval_ms,
		"jitter_ms": jitter_ms,
		"packet_staleness_ms": packet_staleness_ms,
		"packet_age_ms": packet_age_ms,
	}


func set_network_metrics(metrics: Dictionary) -> void:
	if metrics.has("server_clock_offset_ms"):
		var offset_ms := int(metrics["server_clock_offset_ms"])
		if offset_ms >= 0:
			server_clock_offset_ms = offset_ms


func apply_gameplay_state(state: Dictionary) -> void:
	server_sent_msec = int(state.get("server_sent_msec", -1))

	previous_packet_arrival_msec = latest_packet_arrival_msec
	latest_packet_arrival_msec = Time.get_ticks_msec()
	if previous_packet_arrival_msec >= 0:
		previous_packet_interval_ms = packet_interval_ms
		packet_interval_ms = latest_packet_arrival_msec - previous_packet_arrival_msec
		jitter_ms = abs(packet_interval_ms - previous_packet_interval_ms) if previous_packet_interval_ms >= 0 else -1
	else:
		packet_interval_ms = -1
		previous_packet_interval_ms = -1
		jitter_ms = -1

	var server_players_variant: Variant = state.get("server_players", null)
	players = server_players_variant.size() if server_players_variant is Dictionary else 0

	var server_asteroids_variant: Variant = state.get("server_asteroids", null)
	asteroids = server_asteroids_variant.size() if server_asteroids_variant is Dictionary else 0

	var server_pickups_variant: Variant = state.get("server_pickups", null)
	pickups = server_pickups_variant.size() if server_pickups_variant is Dictionary else 0

	total_asteroids = int(state.get("total_asteroids", 0))

	var server_bullets_variant: Variant = state.get("server_bullets", null)
	bullets = server_bullets_variant.size() if server_bullets_variant is Dictionary else 0

	var server_enemies_variant: Variant = state.get("server_enemies", null)
	if server_enemies_variant is Dictionary:
		enemies = server_enemies_variant.size()
		return

	var enemies_variant: Variant = state.get("enemies", null)
	enemies = enemies_variant.size() if enemies_variant is Dictionary else 0
