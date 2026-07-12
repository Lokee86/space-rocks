extends RefCounted

const RealtimeQuantize = preload("res://scripts/protocol/realtime/realtime_quantize.gd")
const WorldLaneState = preload("res://scripts/protocol/realtime/world_lane_state.gd")
const BaselineTracker = preload("res://scripts/protocol/realtime/baseline_tracker.gd")
const WorldFullChunkAssembler = preload("res://scripts/protocol/realtime/world_full_chunk_assembler.gd")

var _full_assembler := WorldFullChunkAssembler.new()

func apply_world_full(world_lane_state: WorldLaneState, baseline_tracker: BaselineTracker, lane: String, world_packet: Dictionary) -> bool:
	var baseline_id = world_packet.get("baseline_id")
	var sequence = world_packet.get("sequence")
	var snapshot_id = world_packet.get("snapshot_id")
	var chunk_index: int = int(world_packet.get("chunk_index", 0))
	var chunk_count: int = int(world_packet.get("chunk_count", 1))
	var is_final_chunk: bool = bool(world_packet.get("is_final_chunk", true))

	var assembly := _full_assembler.accept(world_packet)
	if assembly.status == WorldFullChunkAssembler.ERROR:
		baseline_tracker.request_resync_for_lane(lane, assembly.reason)
		return false
	var accepted := baseline_tracker.record_full_chunk(lane, baseline_id, sequence, snapshot_id, chunk_index, chunk_count, is_final_chunk)
	if not accepted:
		_full_assembler.reset()
		return false
	if assembly.status != WorldFullChunkAssembler.COMPLETE:
		return true
	var assembled: Dictionary = assembly.packet
	world_lane_state.clear_pending_bullet_updates()
	world_lane_state.apply_full_lane(_decode_world_full_packet(assembled))
	return true

func reset() -> void:
	_full_assembler.reset()

func apply_world_delta(world_lane_state: WorldLaneState, baseline_tracker: BaselineTracker, lane: String, world_packet: Dictionary) -> bool:
	var baseline_id = world_packet.get("baseline_id")
	var sequence = world_packet.get("sequence")
	var snapshot_id = world_packet.get("snapshot_id")

	if not baseline_tracker.record_delta(lane, baseline_id, sequence, snapshot_id):
		return false

	_apply_entity_deltas(world_lane_state, _array_field(world_packet, "ship_creates"), _array_field(world_packet, "ship_updates"), _array_field(world_packet, "ship_deletes"), "ship")
	_apply_entity_deltas(world_lane_state, _array_field(world_packet, "bullet_creates"), _array_field(world_packet, "bullet_updates"), _array_field(world_packet, "bullet_deletes"), "bullet")
	_apply_entity_deltas(world_lane_state, _array_field(world_packet, "asteroid_creates"), _array_field(world_packet, "asteroid_updates"), _array_field(world_packet, "asteroid_deletes"), "asteroid")
	_apply_entity_deltas(world_lane_state, _array_field(world_packet, "pickup_creates"), _array_field(world_packet, "pickup_updates"), _array_field(world_packet, "pickup_deletes"), "pickup")
	return true


func apply_asteroid_delta(world_lane_state: WorldLaneState, lane: String, asteroid_packet: Dictionary) -> void:
	if not world_lane_state.accept_asteroid_delta_sequence(asteroid_packet.get("sequence"), asteroid_packet.get("chunk_index", 0), asteroid_packet.get("chunk_count", 1)):
		return
	_apply_entity_deltas(world_lane_state, [], _array_field(asteroid_packet, "asteroid_updates"), [], "asteroid")

func apply_asteroids_lifecycle(world_lane_state: WorldLaneState, packet: Dictionary) -> bool:
	if not _valid_lifecycle_payload(packet, "asteroid_creates", "asteroid_deletes"):
		return false
	_apply_entity_deltas(world_lane_state, _array_field(packet, "asteroid_creates"), [], _array_field(packet, "asteroid_deletes"), "asteroid")
	return true

func apply_bullet_delta(world_lane_state: WorldLaneState, lane: String, bullet_packet: Dictionary) -> void:
	if not world_lane_state.accept_bullet_delta_sequence(bullet_packet.get("sequence"), bullet_packet.get("chunk_index", 0), bullet_packet.get("chunk_count", 1)):
		return
	_apply_entity_deltas(world_lane_state, [], _array_field(bullet_packet, "bullet_updates"), [], "bullet")

func apply_bullets_lifecycle(world_lane_state: WorldLaneState, packet: Dictionary) -> bool:
	if not _valid_lifecycle_payload(packet, "bullet_creates", "bullet_deletes"):
		return false
	_apply_entity_deltas(world_lane_state, _array_field(packet, "bullet_creates"), [], _array_field(packet, "bullet_deletes"), "bullet")
	return true

func _decode_world_full_packet(world_packet: Dictionary) -> Dictionary:
	var decoded := world_packet.duplicate(true)
	decoded["ships"] = _decode_entity_records(_array_field(world_packet, "ships"), "ship")
	decoded["bullets"] = _decode_entity_records(_array_field(world_packet, "bullets"), "bullet")
	decoded["asteroids"] = _decode_entity_records(_array_field(world_packet, "asteroids"), "asteroid")
	decoded["pickups"] = _decode_entity_records(_array_field(world_packet, "pickups"), "pickup")
	return decoded

func _decode_entity_records(records: Array, entity_kind: String) -> Array:
	var decoded: Array = []
	for record in records:
		if not (record is Dictionary):
			decoded.append(record)
			continue
		match entity_kind:
			"ship":
				decoded.append(RealtimeQuantize.decode_world_ship_record(record))
			"bullet":
				decoded.append(RealtimeQuantize.decode_world_bullet_record(record))
			"asteroid":
				decoded.append(RealtimeQuantize.decode_world_asteroid_record(record))
			"pickup":
				decoded.append(RealtimeQuantize.decode_world_pickup_record(record))
			_:
				decoded.append(record)
	return decoded

func _array_field(packet: Dictionary, key: String) -> Array:
	# Missing sparse delta arrays are intentionally treated as empty no-ops.
	var value = packet.get(key, [])
	if value is Array:
		return value
	return []

func _valid_lifecycle_payload(packet: Dictionary, creates_key: String, deletes_key: String) -> bool:
	for key in [creates_key, deletes_key]:
		if packet.has(key) and not packet[key] is Array:
			return false
	for record in _array_field(packet, creates_key):
		if not record is Dictionary:
			return false
	return true

func _apply_entity_deltas(world_lane_state: WorldLaneState, creates: Array, updates: Array, deletes: Array, entity_kind: String) -> void:
	for record in creates:
		_apply_entity_create(world_lane_state, record, entity_kind)
	for record in updates:
		_apply_entity_update(world_lane_state, record, entity_kind)
	for id in deletes:
		_apply_entity_delete(world_lane_state, id, entity_kind)

func _apply_entity_create(world_lane_state: WorldLaneState, record: Dictionary, entity_kind: String) -> void:
	var decoded := _decode_entity_record(record, entity_kind)
	match entity_kind:
		"ship":
			world_lane_state.upsert_ship(decoded)
		"bullet":
			world_lane_state.upsert_bullet(decoded)
		"asteroid":
			world_lane_state.upsert_asteroid(decoded)
		"pickup":
			world_lane_state.upsert_pickup(decoded)

func _apply_entity_update(world_lane_state: WorldLaneState, record: Dictionary, entity_kind: String) -> void:
	var decoded := _decode_entity_record(record, entity_kind)
	match entity_kind:
		"ship":
			world_lane_state.merge_ship_update(decoded)
		"bullet":
			world_lane_state.merge_or_buffer_bullet_update(decoded)
		"asteroid":
			world_lane_state.merge_asteroid_update(decoded)
		"pickup":
			world_lane_state.merge_pickup_update(decoded)
		_:
			_apply_entity_create(world_lane_state, decoded, entity_kind)

func _decode_entity_record(record: Dictionary, entity_kind: String) -> Dictionary:
	if record == null:
		return {}
	match entity_kind:
		"ship":
			return RealtimeQuantize.decode_world_ship_record(record)
		"bullet":
			return RealtimeQuantize.decode_world_bullet_record(record)
		"asteroid":
			return RealtimeQuantize.decode_world_asteroid_record(record)
		"pickup":
			return RealtimeQuantize.decode_world_pickup_record(record)
		_:
			return record

func _apply_entity_delete(world_lane_state: WorldLaneState, id, entity_kind: String) -> void:
	match entity_kind:
		"ship":
			world_lane_state.delete_ship(id)
		"bullet":
			world_lane_state.delete_bullet(id)
			world_lane_state.clear_pending_bullet_update(id)
		"asteroid":
			world_lane_state.delete_asteroid(id)
		"pickup":
			world_lane_state.delete_pickup(id)
