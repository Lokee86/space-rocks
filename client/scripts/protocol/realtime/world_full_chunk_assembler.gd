extends RefCounted
class_name WorldFullChunkAssembler

const LaneMetadata := preload("res://scripts/protocol/realtime/lane_metadata.gd")
const RealtimeReceiveLimits := preload("res://scripts/protocol/realtime/realtime_receive_limits.gd")

const INCOMPLETE := "incomplete"
const COMPLETE := "complete"
const ERROR := "error"

var _identity := {}
var _chunks := []
var _started_at := 0
var _record_count := 0
var _estimated_bytes := 0
var _clock: Callable

func _init(clock: Callable = Callable(Time, "get_ticks_msec")) -> void:
	_clock = clock

func accept(packet: Dictionary) -> Dictionary:
	if not _chunks.is_empty() and _now_msec() - _started_at >= RealtimeReceiveLimits.WORLD_FULL_ASSEMBLY_LIFETIME_MSEC:
		return _fail("expired")
	var sequence = LaneMetadata.parse_non_negative_integer(packet.get("sequence"))
	var chunk_index = LaneMetadata.parse_non_negative_integer(packet.get("chunk_index", 0))
	var chunk_count = LaneMetadata.parse_non_negative_integer(packet.get("chunk_count", 1))
	var baseline_id = packet.get("baseline_id")
	var final = packet.get("is_final_chunk", true)
	if sequence == null or chunk_index == null or chunk_count == null or not LaneMetadata.is_valid_id(baseline_id) or typeof(final) != TYPE_BOOL:
		return _fail("invalid_chunk_metadata")
	if chunk_count < 1 or chunk_count > RealtimeReceiveLimits.MAX_WORLD_FULL_CHUNKS_PER_ASSEMBLY or chunk_index >= chunk_count or final != (chunk_index == chunk_count - 1):
		return _fail("invalid_chunk_metadata" if chunk_count <= RealtimeReceiveLimits.MAX_WORLD_FULL_CHUNKS_PER_ASSEMBLY else "chunk_limit")
	var identity := {"match_id": packet.get("match_id"), "lane": packet.get("lane"), "sequence": sequence, "baseline_id": baseline_id, "snapshot_id": packet.get("snapshot_id"), "snapshot_kind": packet.get("snapshot_kind"), "chunk_count": chunk_count}
	if not _chunks.is_empty() and chunk_index == 0:
		return _fail("interrupted_chunk_series")
	if _chunks.is_empty():
		_identity = identity
		_started_at = _now_msec()
	elif identity != _identity:
		return _fail("mismatched_chunk_identity")
	if chunk_index != _chunks.size():
		return _fail("non_contiguous_chunk_index")
	var records := 0
	for field in ["ships", "bullets", "asteroids", "pickups"]:
		var field_records = packet.get(field, [])
		if not field_records is Array:
			return _fail("invalid_record_array")
		records += field_records.size()
	var estimated_bytes := JSON.stringify(packet).to_utf8_buffer().size()
	if _record_count + records > RealtimeReceiveLimits.MAX_WORLD_FULL_RECORDS_PER_ASSEMBLY:
		return _fail("record_limit")
	if _estimated_bytes + estimated_bytes > RealtimeReceiveLimits.MAX_WORLD_FULL_ESTIMATED_BYTES_PER_ASSEMBLY:
		return _fail("byte_limit")
	_chunks.append(packet.duplicate(true))
	_record_count += records
	_estimated_bytes += estimated_bytes
	if not final:
		return {"status": INCOMPLETE}
	var assembled := packet.duplicate(true)
	for field in ["ships", "bullets", "asteroids", "pickups"]:
		assembled[field] = []
	for chunk in _chunks:
		for field in ["ships", "bullets", "asteroids", "pickups"]:
			assembled[field].append_array(chunk.get(field, []))
	reset()
	return {"status": COMPLETE, "packet": assembled}

func reset() -> void:
	_identity = {}
	_chunks.clear()
	_started_at = 0
	_record_count = 0
	_estimated_bytes = 0

func _now_msec() -> int:
	return int(_clock.call())

func _fail(reason: String) -> Dictionary:
	reset()
	return {"status": ERROR, "reason": reason}
