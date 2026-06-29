extends RefCounted

const SHIP_FIELDS := ["id", "x", "y", "rotation", "velocity_x", "velocity_y", "thrusting", "health", "shields"]
const BULLET_FIELDS := ["id", "x", "y", "velocity_x", "velocity_y", "rotation", "owner_id", "lifespan_seconds"]
const ASTEROID_FIELDS := ["id", "x", "y", "velocity_x", "velocity_y", "rotation", "size", "health"]
const PICKUP_FIELDS := ["id", "x", "y", "pickup_type"]

var ships := {}
var bullets := {}
var asteroids := {}
var pickups := {}

func clear_world() -> void:
	ships.clear()
	bullets.clear()
	asteroids.clear()
	pickups.clear()

func apply_full_lane(world_state: Dictionary) -> void:
	clear_world()
	_replace_records(ships, world_state.get("ships", []), SHIP_FIELDS)
	_replace_records(bullets, world_state.get("bullets", []), BULLET_FIELDS)
	_replace_records(asteroids, world_state.get("asteroids", []), ASTEROID_FIELDS)
	_replace_records(pickups, world_state.get("pickups", []), PICKUP_FIELDS)

func replace_ships(records: Array) -> void:
	_replace_records(ships, records, SHIP_FIELDS)

func replace_bullets(records: Array) -> void:
	_replace_records(bullets, records, BULLET_FIELDS)

func replace_asteroids(records: Array) -> void:
	_replace_records(asteroids, records, ASTEROID_FIELDS)

func replace_pickups(records: Array) -> void:
	_replace_records(pickups, records, PICKUP_FIELDS)

func upsert_ship(record: Dictionary) -> void:
	_upsert_record(ships, record, SHIP_FIELDS)

func upsert_bullet(record: Dictionary) -> void:
	_upsert_record(bullets, record, BULLET_FIELDS)

func upsert_asteroid(record: Dictionary) -> void:
	_upsert_record(asteroids, record, ASTEROID_FIELDS)

func upsert_pickup(record: Dictionary) -> void:
	_upsert_record(pickups, record, PICKUP_FIELDS)

func delete_ship(id) -> void:
	ships.erase(id)

func delete_bullet(id) -> void:
	bullets.erase(id)

func delete_asteroid(id) -> void:
	asteroids.erase(id)

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

func _narrow_record(record: Dictionary, fields: Array) -> Dictionary:
	var narrowed := {}
	for field in fields:
		if record.has(field):
			narrowed[field] = record[field]
	return narrowed
