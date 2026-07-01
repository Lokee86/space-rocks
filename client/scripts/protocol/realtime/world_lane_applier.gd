extends RefCounted

const WorldLaneState = preload("res://scripts/protocol/realtime/world_lane_state.gd")
const BaselineTracker = preload("res://scripts/protocol/realtime/baseline_tracker.gd")

func apply_world_full(world_lane_state: WorldLaneState, baseline_tracker: BaselineTracker, lane: String, world_packet: Dictionary) -> void:
	var baseline_id = world_packet.get("baseline_id")
	var sequence = world_packet.get("sequence")
	var snapshot_id = world_packet.get("snapshot_id")
	var chunk_index: int = int(world_packet.get("chunk_index", 0))
	var chunk_count: int = int(world_packet.get("chunk_count", 1))
	var is_final_chunk: bool = bool(world_packet.get("is_final_chunk", true))

	world_lane_state.apply_full_lane(world_packet)

	if is_final_chunk:
		baseline_tracker.record_full_packet(lane, baseline_id, sequence, snapshot_id, chunk_index, chunk_count, true)
	else:
		baseline_tracker.record_full_chunk(lane, baseline_id, sequence, snapshot_id, chunk_index, chunk_count, false)

func apply_world_delta(world_lane_state: WorldLaneState, baseline_tracker: BaselineTracker, lane: String, world_packet: Dictionary) -> bool:
	var baseline_id = world_packet.get("baseline_id")
	var sequence = world_packet.get("sequence")

	if not baseline_tracker.record_delta(lane, baseline_id, sequence):
		return false

	_apply_entity_deltas(world_lane_state, _array_field(world_packet, "ship_creates"), _array_field(world_packet, "ship_updates"), _array_field(world_packet, "ship_deletes"), "ship")
	_apply_entity_deltas(world_lane_state, _array_field(world_packet, "bullet_creates"), _array_field(world_packet, "bullet_updates"), _array_field(world_packet, "bullet_deletes"), "bullet")
	_apply_entity_deltas(world_lane_state, _array_field(world_packet, "asteroid_creates"), _array_field(world_packet, "asteroid_updates"), _array_field(world_packet, "asteroid_deletes"), "asteroid")
	_apply_entity_deltas(world_lane_state, _array_field(world_packet, "pickup_creates"), _array_field(world_packet, "pickup_updates"), _array_field(world_packet, "pickup_deletes"), "pickup")
	return true

func _array_field(packet: Dictionary, key: String) -> Array:
	var value = packet.get(key, [])
	if value is Array:
		return value
	return []

func _apply_entity_deltas(world_lane_state: WorldLaneState, creates: Array, updates: Array, deletes: Array, entity_kind: String) -> void:
	for record in creates:
		_apply_entity_create(world_lane_state, record, entity_kind)
	for record in updates:
		_apply_entity_update(world_lane_state, record, entity_kind)
	for id in deletes:
		_apply_entity_delete(world_lane_state, id, entity_kind)

func _apply_entity_create(world_lane_state: WorldLaneState, record: Dictionary, entity_kind: String) -> void:
	match entity_kind:
		"ship":
			world_lane_state.upsert_ship(record)
		"bullet":
			world_lane_state.upsert_bullet(record)
		"asteroid":
			world_lane_state.upsert_asteroid(record)
		"pickup":
			world_lane_state.upsert_pickup(record)

func _apply_entity_update(world_lane_state: WorldLaneState, record: Dictionary, entity_kind: String) -> void:
	match entity_kind:
		"ship":
			world_lane_state.merge_ship_update(record)
		"bullet":
			world_lane_state.merge_bullet_update(record)
		"asteroid":
			world_lane_state.merge_asteroid_update(record)
		"pickup":
			world_lane_state.merge_pickup_update(record)
		_:
			_apply_entity_create(world_lane_state, record, entity_kind)

func _apply_entity_delete(world_lane_state: WorldLaneState, id, entity_kind: String) -> void:
	match entity_kind:
		"ship":
			world_lane_state.delete_ship(id)
		"bullet":
			world_lane_state.delete_bullet(id)
		"asteroid":
			world_lane_state.delete_asteroid(id)
		"pickup":
			world_lane_state.delete_pickup(id)
