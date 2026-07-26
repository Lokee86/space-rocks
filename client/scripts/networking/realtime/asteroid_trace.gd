extends RefCounted
class_name AsteroidTrace

const ClientLogger := preload("res://scripts/logging/logger.gd")

const ENV_NAME := "SPACE_ROCKS_ASTEROID_TRACE"
const MAX_ENTRIES := 96
const DUMP_ENTRIES := 12
const HOT_SAMPLE_INTERVAL := 30
const ANOMALY_THROTTLE_MSEC := 1000

static var _entries: Array = []
static var _enabled_cache := -1
static var _last_anomaly_code := ""
static var _last_anomaly_msec := -ANOMALY_THROTTLE_MSEC


static func enabled() -> bool:
	if _enabled_cache >= 0:
		return _enabled_cache == 1
	var configured := OS.get_environment(ENV_NAME).strip_edges().to_lower()
	if configured.is_empty():
		_enabled_cache = 1 if OS.is_debug_build() and not _is_test_process() else 0
	else:
		_enabled_cache = 1 if configured in ["1", "true", "yes", "on"] else 0
	return _enabled_cache == 1


static func configure() -> void:
	if not enabled():
		return
	ClientLogger.set_category_level(ClientLogger.CATEGORY_PACKETS, ClientLogger.LEVEL_DEBUG)
	ClientLogger.set_category_level(ClientLogger.CATEGORY_WORLD_SYNC, ClientLogger.LEVEL_DEBUG)
	ClientLogger.packets_event(
		ClientLogger.LEVEL_INFO,
		"asteroid_trace_enabled",
		"Asteroid realtime tracing enabled",
		{"ring_capacity": MAX_ENTRIES, "environment_variable": ENV_NAME}
	)


static func reset() -> void:
	_entries.clear()
	_last_anomaly_code = ""
	_last_anomaly_msec = -ANOMALY_THROTTLE_MSEC


static func record_packet(packet: Dictionary, asteroid_count: int) -> void:
	if not enabled() or not _is_asteroid_packet(packet):
		return
	var summary := {
		"stage": "client_receive",
		"time_msec": Time.get_ticks_msec(),
		"type": str(packet.get("type", "")),
		"lane": str(packet.get("lane", "")),
		"sequence": packet.get("sequence", -1),
		"baseline_id": str(packet.get("baseline_id", "")),
		"baseline_sequence": packet.get("baseline_sequence", -1),
		"chunk_index": packet.get("chunk_index", 0),
		"chunk_count": packet.get("chunk_count", 1),
		"creates": _record_ids(packet.get("asteroid_creates", [])),
		"updates": _record_ids(packet.get("asteroid_updates", [])),
		"deletes": _delete_ids(packet.get("asteroid_deletes", [])),
		"state_count": asteroid_count,
	}
	_append(summary)
	var packet_type := str(summary["type"])
	var sequence := int(summary["sequence"])
	if packet_type == "asteroids_lifecycle" or sequence % HOT_SAMPLE_INTERVAL == 0:
		_emit_debug("asteroid_packet_received", summary)


static func record_event(operation: String, fields: Dictionary = {}) -> void:
	if not enabled():
		return
	var entry := fields.duplicate(true)
	entry["stage"] = operation
	entry["time_msec"] = Time.get_ticks_msec()
	_append(entry)
	_emit_debug("asteroid_trace", entry)


static func anomaly(code: String, fields: Dictionary = {}) -> void:
	if not enabled():
		return
	var entry := fields.duplicate(true)
	entry["stage"] = "anomaly"
	entry["failure_mode"] = code
	entry["time_msec"] = Time.get_ticks_msec()
	_append(entry)
	var now := Time.get_ticks_msec()
	if code == _last_anomaly_code and now - _last_anomaly_msec < ANOMALY_THROTTLE_MSEC:
		return
	_last_anomaly_code = code
	_last_anomaly_msec = now
	var event_fields := entry.duplicate(true)
	event_fields["recent_trace"] = JSON.stringify(_recent_entries())
	ClientLogger.packets_event(
		ClientLogger.LEVEL_WARN,
		"asteroid_trace_anomaly",
		"Asteroid realtime anomaly detected",
		event_fields
	)


static func _append(entry: Dictionary) -> void:
	_entries.append(entry)
	while _entries.size() > MAX_ENTRIES:
		_entries.pop_front()


static func _recent_entries() -> Array:
	var start := maxi(0, _entries.size() - DUMP_ENTRIES)
	return _entries.slice(start, _entries.size())


static func _emit_debug(event_name: String, fields: Dictionary) -> void:
	ClientLogger.packets_event(ClientLogger.LEVEL_DEBUG, event_name, "", fields)


static func _is_test_process() -> bool:
	for argument in OS.get_cmdline_args():
		if str(argument).contains("gut_cmdln.gd"):
			return true
	return false


static func _is_asteroid_packet(packet: Dictionary) -> bool:
	var packet_type := str(packet.get("type", ""))
	return packet_type in ["asteroid_delta", "asteroids_lifecycle", "world_full"]


static func _record_ids(value) -> Array:
	var ids: Array = []
	if not value is Array:
		return ids
	for record in value:
		if record is Dictionary and record.has("id"):
			ids.append(str(record["id"]))
	return ids.slice(0, mini(ids.size(), 12))


static func _delete_ids(value) -> Array:
	var ids: Array = []
	if not value is Array:
		return ids
	for id in value:
		ids.append(str(id))
	return ids.slice(0, mini(ids.size(), 12))
