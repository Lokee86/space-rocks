extends RefCounted
class_name LifecycleChunkAssembler

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

func accept(packet: Dictionary, creates_key: String, deletes_key: String) -> Dictionary:
	if not _chunks.is_empty() and _now_msec() - _started_at >= RealtimeReceiveLimits.LIFECYCLE_ASSEMBLY_LIFETIME_MSEC:
		return _fail("expired")
	var sequence = LaneMetadata.parse_non_negative_integer(packet.get("sequence"))
	var index = LaneMetadata.parse_non_negative_integer(packet.get("chunk_index", 0))
	var count = LaneMetadata.parse_non_negative_integer(packet.get("chunk_count", 1))
	var baseline_id = packet.get("baseline_id")
	var final = packet.get("is_final_chunk", true)
	if sequence == null or index == null or count == null or not LaneMetadata.is_valid_id(baseline_id) or typeof(final) != TYPE_BOOL:
		return _fail("invalid_lifecycle_chunk_metadata")
	if count < 1 or count > RealtimeReceiveLimits.MAX_LIFECYCLE_CHUNKS_PER_ASSEMBLY or index >= count or final != (index == count - 1):
		return _fail("chunk_limit" if count > RealtimeReceiveLimits.MAX_LIFECYCLE_CHUNKS_PER_ASSEMBLY else "invalid_lifecycle_chunk_metadata")
	var identity := {"match_id": packet.get("match_id"), "lane": packet.get("lane"), "sequence": sequence, "baseline_id": baseline_id, "snapshot_id": packet.get("snapshot_id"), "snapshot_kind": packet.get("snapshot_kind"), "chunk_count": count}
	if not _chunks.is_empty() and index == 0:
		return _fail("interrupted_lifecycle_chunk_series")
	if _chunks.is_empty():
		_identity = identity
		_started_at = _now_msec()
	elif identity != _identity:
		return _fail("mismatched_lifecycle_chunk_identity")
	if index != _chunks.size():
		return _fail("non_contiguous_lifecycle_chunk_index")
	var creates = packet.get(creates_key, [])
	var deletes = packet.get(deletes_key, [])
	if not creates is Array or not deletes is Array:
		return _fail("invalid_lifecycle_record_array")
	var records: int = creates.size() + deletes.size()
	var estimated_bytes := JSON.stringify(packet).to_utf8_buffer().size()
	if _record_count + records > RealtimeReceiveLimits.MAX_LIFECYCLE_RECORDS_PER_ASSEMBLY:
		return _fail("record_limit")
	if _estimated_bytes + estimated_bytes > RealtimeReceiveLimits.MAX_LIFECYCLE_ESTIMATED_BYTES_PER_ASSEMBLY:
		return _fail("byte_limit")
	_chunks.append(packet.duplicate(true))
	_record_count += records
	_estimated_bytes += estimated_bytes
	if not final:
		return {"status": INCOMPLETE}
	var assembled := packet.duplicate(true)
	assembled[creates_key] = []
	assembled[deletes_key] = []
	for chunk in _chunks:
		assembled[creates_key].append_array(chunk.get(creates_key, []))
		assembled[deletes_key].append_array(chunk.get(deletes_key, []))
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
