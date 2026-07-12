extends RefCounted
class_name WorldFullChunkAssembler

const INCOMPLETE := "incomplete"
const COMPLETE := "complete"
const ERROR := "error"

var _identity := {}
var _chunks := []

func accept(packet: Dictionary) -> Dictionary:
	var identity := {
		"match_id": packet.get("match_id"),
		"lane": packet.get("lane"),
		"sequence": packet.get("sequence"),
		"baseline_id": packet.get("baseline_id"),
		"snapshot_id": packet.get("snapshot_id"),
		"snapshot_kind": packet.get("snapshot_kind"),
		"chunk_count": packet.get("chunk_count", 1),
	}
	var index := int(packet.get("chunk_index", 0))
	var count := int(packet.get("chunk_count", 1))
	var final := bool(packet.get("is_final_chunk", true))
	if count < 1 or index < 0 or index >= count or final != (index == count - 1):
		return _fail("invalid_chunk_metadata")
	if _chunks.size() > 0 and index == 0:
		return _fail("interrupted_chunk_series")
	if _chunks.size() == 0:
		_identity = identity
		_chunks = []
	elif identity != _identity:
		return _fail("mismatched_chunk_identity")
	if index != _chunks.size():
		return _fail("non_contiguous_chunk_index")
	_chunks.append(packet.duplicate(true))
	if not final:
		return {"status": INCOMPLETE}
	var assembled := packet.duplicate(true)
	for field in ["ships", "bullets", "asteroids", "pickups"]:
		assembled[field] = []
	for chunk in _chunks:
		for field in ["ships", "bullets", "asteroids", "pickups"]:
			var records = chunk.get(field, [])
			if not records is Array:
				return _fail("invalid_record_array")
			assembled[field].append_array(records)
	reset()
	return {"status": COMPLETE, "packet": assembled}

func reset() -> void:
	_identity = {}
	_chunks.clear()

func _fail(reason: String) -> Dictionary:
	reset()
	return {"status": ERROR, "reason": reason}
