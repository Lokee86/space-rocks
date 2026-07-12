extends RefCounted
class_name LifecycleChunkAssembler

const INCOMPLETE := "incomplete"
const COMPLETE := "complete"
const ERROR := "error"

var _identity := {}
var _chunks := []

func accept(packet: Dictionary, creates_key: String, deletes_key: String) -> Dictionary:
	var index := int(packet.get("chunk_index", 0))
	var count := int(packet.get("chunk_count", 1))
	var final := bool(packet.get("is_final_chunk", true))
	if count < 1 or index < 0 or index >= count or final != (index == count - 1):
		return _fail("invalid_lifecycle_chunk_metadata")
	var identity := {"match_id": packet.get("match_id"), "lane": packet.get("lane"), "sequence": packet.get("sequence"), "baseline_id": packet.get("baseline_id"), "snapshot_id": packet.get("snapshot_id"), "snapshot_kind": packet.get("snapshot_kind"), "chunk_count": count}
	if _chunks.size() > 0 and index == 0:
		return _fail("interrupted_lifecycle_chunk_series")
	if _chunks.is_empty():
		_identity = identity
	elif identity != _identity:
		return _fail("mismatched_lifecycle_chunk_identity")
	if index != _chunks.size():
		return _fail("non_contiguous_lifecycle_chunk_index")
	for key in [creates_key, deletes_key]:
		if packet.has(key) and not packet[key] is Array:
			return _fail("invalid_lifecycle_record_array")
	_chunks.append(packet.duplicate(true))
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

func _fail(reason: String) -> Dictionary:
	reset()
	return {"status": ERROR, "reason": reason}
