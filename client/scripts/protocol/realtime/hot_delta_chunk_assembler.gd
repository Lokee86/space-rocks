extends RefCounted
class_name HotDeltaChunkAssembler

const LaneMetadata := preload("res://scripts/protocol/realtime/lane_metadata.gd")
const RealtimeReceiveLimits := preload("res://scripts/protocol/realtime/realtime_receive_limits.gd")

const PENDING := "pending"
const COMPLETE := "complete"
const REJECTED := "rejected"

var _latest_sequence := -1
var _active_identity := {}
var _chunks := {}
var _record_count := 0
var _estimated_bytes := 0


func accept(packet: Dictionary, updates_key: String) -> Dictionary:
	var sequence = LaneMetadata.parse_non_negative_integer(packet.get("sequence"))
	var chunk_index = LaneMetadata.parse_non_negative_integer(packet.get("chunk_index", 0))
	var chunk_count = LaneMetadata.parse_non_negative_integer(packet.get("chunk_count", 1))
	if sequence == null or chunk_index == null or chunk_count == null:
		return _rejected("invalid_chunk_metadata")
	if (
		chunk_count < 1
		or chunk_count > RealtimeReceiveLimits.MAX_HOT_DELTA_CHUNKS_PER_ASSEMBLY
		or chunk_index >= chunk_count
	):
		return _rejected(
			"chunk_limit"
			if chunk_count > RealtimeReceiveLimits.MAX_HOT_DELTA_CHUNKS_PER_ASSEMBLY
			else "invalid_chunk_metadata"
		)

	var updates = packet.get(updates_key, [])
	if not updates is Array:
		return _rejected("invalid_update_array")

	var identity := {
		"match_id": packet.get("match_id"),
		"lane": packet.get("lane"),
		"sequence": sequence,
		"baseline_id": packet.get("baseline_id"),
		"snapshot_id": packet.get("snapshot_id"),
		"chunk_count": chunk_count,
	}
	var superseded := {}
	if not _active_identity.is_empty():
		var active_sequence: int = int(_active_identity.get("sequence", -1))
		if sequence < active_sequence:
			return _rejected("stale_sequence")
		if sequence > active_sequence:
			superseded = _incomplete_summary(sequence)
			_clear_active()
		elif identity != _active_identity:
			return _rejected("mismatched_chunk_identity")
	elif sequence <= _latest_sequence:
		return _rejected("stale_or_completed_sequence")

	if _active_identity.is_empty():
		_active_identity = identity
		_latest_sequence = sequence

	if _chunks.has(chunk_index):
		return _rejected("duplicate_chunk")

	var estimated_bytes := JSON.stringify(packet).to_utf8_buffer().size()
	if _record_count + updates.size() > RealtimeReceiveLimits.MAX_HOT_DELTA_RECORDS_PER_ASSEMBLY:
		return _rejected("record_limit")
	if _estimated_bytes + estimated_bytes > RealtimeReceiveLimits.MAX_HOT_DELTA_ESTIMATED_BYTES_PER_ASSEMBLY:
		return _rejected("byte_limit")

	_chunks[chunk_index] = packet.duplicate(true)
	_record_count += updates.size()
	_estimated_bytes += estimated_bytes
	if _chunks.size() < chunk_count:
		var pending := {"status": PENDING}
		if not superseded.is_empty():
			pending["superseded"] = superseded
		return pending

	var assembled := packet.duplicate(true)
	assembled[updates_key] = []
	for index in range(chunk_count):
		if not _chunks.has(index):
			return _rejected("missing_chunk")
		assembled[updates_key].append_array(_chunks[index].get(updates_key, []))
	assembled["chunk_index"] = 0
	assembled["chunk_count"] = 1
	assembled["is_final_chunk"] = true
	_clear_active()
	var complete := {"status": COMPLETE, "packet": assembled}
	if not superseded.is_empty():
		complete["superseded"] = superseded
	return complete


func reset() -> void:
	_latest_sequence = -1
	_clear_active()


func _incomplete_summary(replacement_sequence: int) -> Dictionary:
	return {
		"sequence": int(_active_identity.get("sequence", -1)),
		"replacement_sequence": replacement_sequence,
		"received_chunks": _chunks.size(),
		"expected_chunks": int(_active_identity.get("chunk_count", 0)),
	}


func _clear_active() -> void:
	_active_identity = {}
	_chunks.clear()
	_record_count = 0
	_estimated_bytes = 0


func _rejected(reason: String) -> Dictionary:
	return {"status": REJECTED, "reason": reason}
