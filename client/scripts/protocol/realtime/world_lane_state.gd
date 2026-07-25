extends RefCounted
class_name WorldLaneState

const SHIP_FIELDS := ["id", "x", "y", "rotation", "velocity_x", "velocity_y", "thrusting", "health", "shields", "ship_type", "target_kind", "target_id"]
const BULLET_FIELDS := ["id", "x", "y", "velocity_x", "velocity_y", "rotation", "owner_id", "lifespan_seconds", "weapon_id", "projectile_type"]
const ASTEROID_FIELDS := ["id", "x", "y", "velocity_x", "velocity_y", "rotation", "size", "health", "scale", "variant"]
const PICKUP_FIELDS := ["id", "type", "pickup_class", "x", "y", "health", "age_seconds", "lifespan_seconds"]
const DELETED_SHIP_ID_CAP := 4096
const PENDING_SHIP_UPDATE_CAP := 2048
const DELETED_BULLET_ID_CAP := 4096
const PENDING_BULLET_UPDATE_CAP := 2048

var ships := {}
var pending_ship_updates := {}
var deleted_ship_ids := {}
var _pending_ship_update_order := []
var _deleted_ship_id_order := []
var bullets := {}
var pending_bullet_updates := {}
var deleted_bullet_ids := {}
var _pending_bullet_update_order := []
var _deleted_bullet_id_order := []
var dirty_bullet_ids := {}
var removed_bullet_ids := {}
var bullet_full_sync_required := false
var dirty_asteroid_ids := {}
var removed_asteroid_ids := {}
var asteroid_full_sync_required := false
var asteroids := {}
var pickups := {}
var latest_ship_delta_sequence := -1
var latest_asteroid_delta_sequence := -1
var latest_bullet_delta_sequence := -1
var ship_delta_chunk_count := 1
var asteroid_delta_chunk_count := 1
var bullet_delta_chunk_count := 1
var ship_delta_received_chunks := {}
var asteroid_delta_received_chunks := {}
var bullet_delta_received_chunks := {}

func clear_world() -> void:
	ships.clear()
	pending_ship_updates.clear()
	_pending_ship_update_order.clear()
	deleted_ship_ids.clear()
	_deleted_ship_id_order.clear()
	latest_ship_delta_sequence = -1
	ship_delta_chunk_count = 1
	ship_delta_received_chunks.clear()
	bullets.clear()
	pending_bullet_updates.clear()
	_pending_bullet_update_order.clear()
	deleted_bullet_ids.clear()
	_deleted_bullet_id_order.clear()
	dirty_bullet_ids.clear()
	removed_bullet_ids.clear()
	bullet_full_sync_required = true
	latest_bullet_delta_sequence = -1
	bullet_delta_chunk_count = 1
	bullet_delta_received_chunks.clear()
	clear_asteroid_change_sets()
	asteroid_full_sync_required = true
	latest_asteroid_delta_sequence = -1
	asteroid_delta_chunk_count = 1
	asteroid_delta_received_chunks.clear()
	asteroids.clear()
	pickups.clear()

func apply_full_lane(world_state: Dictionary) -> void:
	clear_world()
	_replace_records(ships, world_state.get("ships", []), SHIP_FIELDS)
	_replace_records(bullets, world_state.get("bullets", []), BULLET_FIELDS)
	_replace_records(asteroids, world_state.get("asteroids", []), ASTEROID_FIELDS)
	asteroid_full_sync_required = true
	_replace_records(pickups, world_state.get("pickups", []), PICKUP_FIELDS)

func replace_ships(records: Array) -> void:
	_replace_records(ships, records, SHIP_FIELDS)

func replace_bullets(records: Array) -> void:
	_replace_records(bullets, records, BULLET_FIELDS)
	dirty_bullet_ids.clear()
	removed_bullet_ids.clear()
	bullet_full_sync_required = true

func replace_asteroids(records: Array) -> void:
	_replace_records(asteroids, records, ASTEROID_FIELDS)
	dirty_asteroid_ids.clear()
	removed_asteroid_ids.clear()
	asteroid_full_sync_required = true

func replace_pickups(records: Array) -> void:
	_replace_records(pickups, records, PICKUP_FIELDS)

func mark_bullet_dirty(id) -> void:
	if id == null or id == "":
		return
	removed_bullet_ids.erase(id)
	dirty_bullet_ids[id] = true

func mark_bullet_removed(id) -> void:
	if id == null or id == "":
		return
	dirty_bullet_ids.erase(id)
	removed_bullet_ids[id] = true

func clear_bullet_change_sets() -> void:
	dirty_bullet_ids.clear()
	removed_bullet_ids.clear()
	bullet_full_sync_required = false

func mark_asteroid_dirty(id) -> void:
	if id == null or id == "":
		return
	removed_asteroid_ids.erase(id)
	dirty_asteroid_ids[id] = true

func mark_asteroid_removed(id) -> void:
	if id == null or id == "":
		return
	dirty_asteroid_ids.erase(id)
	removed_asteroid_ids[id] = true

func clear_asteroid_change_sets() -> void:
	dirty_asteroid_ids.clear()
	removed_asteroid_ids.clear()
	asteroid_full_sync_required = false

func upsert_ship(record: Dictionary) -> void:
	var id = record.get("id")
	if id == null or id == "":
		return
	_clear_deleted_ship_id(id)
	_upsert_record(ships, record, SHIP_FIELDS)
	apply_pending_ship_update(id)

func merge_ship_update(record: Dictionary) -> void:
	var id = record.get("id")
	if id == null or not ships.has(id):
		return
	_merge_record_update(ships, record, SHIP_FIELDS)

func apply_pending_ship_update(id) -> void:
	if not pending_ship_updates.has(id):
		return
	if ships.has(id):
		merge_ship_update(pending_ship_updates[id])
	pending_ship_updates.erase(id)
	_clear_pending_ship_update_order(id)

func merge_or_buffer_ship_update(record: Dictionary) -> void:
	var id = record.get("id")
	if id == null or id == "":
		return
	if ships.has(id):
		merge_ship_update(record)
		return
	if deleted_ship_ids.has(id):
		return
	if not pending_ship_updates.has(id):
		_pending_ship_update_order.append(id)
	pending_ship_updates[id] = record.duplicate(true)
	while _pending_ship_update_order.size() > PENDING_SHIP_UPDATE_CAP:
		pending_ship_updates.erase(_pending_ship_update_order.pop_front())

func clear_pending_ship_updates() -> void:
	pending_ship_updates.clear()
	_pending_ship_update_order.clear()

func clear_pending_ship_update(id) -> void:
	pending_ship_updates.erase(id)
	_clear_pending_ship_update_order(id)

func upsert_bullet(record: Dictionary) -> void:
	var id = record.get("id")
	if id != null:
		_clear_deleted_bullet_id(id)
	_upsert_record(bullets, record, BULLET_FIELDS)
	apply_pending_bullet_update(id)
	mark_bullet_dirty(id)

func merge_bullet_update(record: Dictionary) -> void:
	var id = record.get("id")
	if id == null:
		return
	if not bullets.has(id):
		return
	_merge_record_update(bullets, record, BULLET_FIELDS)
	mark_bullet_dirty(id)

func apply_pending_bullet_update(id) -> void:
	if not pending_bullet_updates.has(id):
		return
	if bullets.has(id):
		merge_bullet_update(pending_bullet_updates[id])
		mark_bullet_dirty(id)
	pending_bullet_updates.erase(id)
	_clear_pending_bullet_update_order(id)

func merge_or_buffer_bullet_update(record: Dictionary) -> void:
	var id = record.get("id")
	if id == null or id == "":
		return
	if bullets.has(id):
		merge_bullet_update(record)
		return
	if deleted_bullet_ids.has(id):
		return
	if not pending_bullet_updates.has(id):
		_pending_bullet_update_order.append(id)
	pending_bullet_updates[id] = record.duplicate(true)
	while _pending_bullet_update_order.size() > PENDING_BULLET_UPDATE_CAP:
		pending_bullet_updates.erase(_pending_bullet_update_order.pop_front())

func clear_pending_bullet_updates() -> void:
	pending_bullet_updates.clear()
	_pending_bullet_update_order.clear()

func clear_pending_bullet_update(id) -> void:
	pending_bullet_updates.erase(id)
	_clear_pending_bullet_update_order(id)

func accept_ship_delta_sequence(sequence, chunk_index = 0, chunk_count = 1) -> bool:
	return _accept_hot_delta(sequence, chunk_index, chunk_count, "ship")

func accept_asteroid_delta_sequence(sequence, chunk_index = 0, chunk_count = 1) -> bool:
	return _accept_hot_delta(sequence, chunk_index, chunk_count, "asteroid")

func accept_bullet_delta_sequence(sequence, chunk_index = 0, chunk_count = 1) -> bool:
	return _accept_hot_delta(sequence, chunk_index, chunk_count, "bullet")

func _accept_hot_delta(sequence, chunk_index, chunk_count, lane: String) -> bool:
	var parsed_sequence = _parse_hot_delta_sequence(sequence)
	var parsed_index = _parse_hot_delta_chunk_index(chunk_index)
	var parsed_count = _parse_hot_delta_chunk_count(chunk_count)
	if parsed_sequence == null or parsed_index == null or parsed_count == null or parsed_index >= parsed_count:
		return false

	var latest_sequence = latest_ship_delta_sequence if lane == "ship" else (latest_asteroid_delta_sequence if lane == "asteroid" else latest_bullet_delta_sequence)
	var tracked_count = ship_delta_chunk_count if lane == "ship" else (asteroid_delta_chunk_count if lane == "asteroid" else bullet_delta_chunk_count)
	var received_chunks = ship_delta_received_chunks if lane == "ship" else (asteroid_delta_received_chunks if lane == "asteroid" else bullet_delta_received_chunks)
	if parsed_sequence < latest_sequence:
		return false
	if parsed_sequence > latest_sequence:
		received_chunks.clear()
		tracked_count = parsed_count
		if lane == "ship":
			latest_ship_delta_sequence = parsed_sequence
			ship_delta_chunk_count = tracked_count
		elif lane == "asteroid":
			latest_asteroid_delta_sequence = parsed_sequence
			asteroid_delta_chunk_count = tracked_count
		else:
			latest_bullet_delta_sequence = parsed_sequence
			bullet_delta_chunk_count = tracked_count
	elif parsed_count != tracked_count or received_chunks.has(parsed_index):
		return false
	received_chunks[parsed_index] = true
	return true

func _parse_hot_delta_sequence(sequence):
	if typeof(sequence) != TYPE_INT and typeof(sequence) != TYPE_FLOAT:
		return null
	if sequence < 0 or (typeof(sequence) == TYPE_FLOAT and (not is_finite(sequence) or sequence != floor(sequence))):
		return null
	return int(sequence)

func _parse_hot_delta_chunk_index(chunk_index):
	if typeof(chunk_index) != TYPE_INT and typeof(chunk_index) != TYPE_FLOAT:
		return null
	var parsed = int(chunk_index)
	if parsed != chunk_index or parsed < 0:
		return null
	return parsed

func _parse_hot_delta_chunk_count(chunk_count):
	if typeof(chunk_count) != TYPE_INT and typeof(chunk_count) != TYPE_FLOAT:
		return null
	var parsed = int(chunk_count)
	if parsed != chunk_count or parsed < 1:
		return null
	return parsed

func upsert_asteroid(record: Dictionary) -> void:
	var id = record.get("id")
	if id == null:
		return
	_upsert_record(asteroids, record, ASTEROID_FIELDS)
	mark_asteroid_dirty(id)

func merge_asteroid_update(record: Dictionary) -> void:
	var id = record.get("id")
	if id == null:
		return
	if not asteroids.has(id):
		return
	_merge_record_update(asteroids, record, ASTEROID_FIELDS)
	mark_asteroid_dirty(id)

func upsert_pickup(record: Dictionary) -> void:
	_upsert_record(pickups, record, PICKUP_FIELDS)

func merge_pickup_update(record: Dictionary) -> void:
	_merge_record_update(pickups, record, PICKUP_FIELDS)

func delete_ship(id) -> void:
	ships.erase(id)
	clear_pending_ship_update(id)
	if not deleted_ship_ids.has(id):
		deleted_ship_ids[id] = true
		_deleted_ship_id_order.append(id)
		while _deleted_ship_id_order.size() > DELETED_SHIP_ID_CAP:
			deleted_ship_ids.erase(_deleted_ship_id_order.pop_front())

func _clear_deleted_ship_id(id) -> void:
	if not deleted_ship_ids.has(id):
		return
	deleted_ship_ids.erase(id)
	_deleted_ship_id_order.erase(id)

func _clear_pending_ship_update_order(id) -> void:
	_pending_ship_update_order.erase(id)

func delete_bullet(id) -> void:
	bullets.erase(id)
	clear_pending_bullet_update(id)
	if not deleted_bullet_ids.has(id):
		deleted_bullet_ids[id] = true
		_deleted_bullet_id_order.append(id)
		while _deleted_bullet_id_order.size() > DELETED_BULLET_ID_CAP:
			deleted_bullet_ids.erase(_deleted_bullet_id_order.pop_front())
	mark_bullet_removed(id)

func _clear_deleted_bullet_id(id) -> void:
	if not deleted_bullet_ids.has(id):
		return
	deleted_bullet_ids.erase(id)
	_deleted_bullet_id_order.erase(id)

func _clear_pending_bullet_update_order(id) -> void:
	_pending_bullet_update_order.erase(id)

func delete_asteroid(id) -> void:
	asteroids.erase(id)
	mark_asteroid_removed(id)

func delete_pickup(id) -> void:
	pickups.erase(id)

func _replace_records(target: Dictionary, records: Array, fields: Array) -> void:
	target.clear()
	for record in records:
		_upsert_record(target, record, fields)

func _upsert_record(target: Dictionary, record: Dictionary, fields: Array) -> void:
	var id = record.get("id")
	if id == null:
		return
	target[id] = _narrow_record(record, fields)

func _merge_record_update(target: Dictionary, record: Dictionary, fields: Array) -> void:
	var id = record.get("id")
	if id == null:
		return
	if not target.has(id):
		return

	var merged: Dictionary = target[id].duplicate(true)
	var narrowed: Dictionary = _narrow_record(record, fields)
	for field in narrowed:
		merged[field] = narrowed[field]
	target[id] = merged

func _narrow_record(record: Dictionary, fields: Array) -> Dictionary:
	var narrowed := {}
	for field in fields:
		if record.has(field):
			narrowed[field] = record[field]
	return narrowed



