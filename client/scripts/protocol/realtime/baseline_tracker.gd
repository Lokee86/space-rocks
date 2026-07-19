extends RefCounted

const LaneMetadata = preload("res://scripts/protocol/realtime/lane_metadata.gd")
const LaneSyncState = preload("res://scripts/protocol/realtime/lane_sync_state.gd")
const ResyncState = preload("res://scripts/protocol/realtime/resync_state.gd")


const LANES := [LaneMetadata.LANE_WORLD, LaneMetadata.LANE_OVERLAY, LaneMetadata.LANE_SESSION]

signal resync_required(lane, baseline_id, last_accepted_sequence, reason)

var _lane_state := {}
var _readiness: GameplayReadiness

func _init() -> void:
	for lane in LANES:
		_lane_state[lane] = _make_lane_state()

func bind_readiness(readiness_ref: GameplayReadiness) -> void:
	_readiness = readiness_ref
	_refresh_readiness()

func get_lane_state(lane: String) -> Dictionary:
	return _lane_state.get(lane, _make_lane_state()).duplicate(true)

func is_lane_synced(lane: String) -> bool:
	if not _is_supported_lane(lane):
		return false
	return _ensure_lane_state(lane).sync_state == LaneSyncState.SYNCED

func get_active_baseline_id(lane: String) -> String:
	if not _is_supported_lane(lane):
		return ""
	var value = _ensure_lane_state(lane).baseline_id
	return value if value is String else ""

func get_last_accepted_sequence(lane: String):
	if not _is_supported_lane(lane):
		return null
	return _ensure_lane_state(lane).sequence

func record_full_packet(lane: String, baseline_id, sequence, snapshot_id = null, chunk_index = 0, chunk_count = 1, is_final_chunk = true) -> bool:
	if not _is_supported_lane(lane):
		return false
	var state := _ensure_lane_state(lane)
	var normalized_sequence = LaneMetadata.parse_non_negative_integer(sequence)
	if normalized_sequence == null:
		_request_resync(lane, state, ResyncState.REASON_STALE_OR_INVALID_SEQUENCE)
		return false
	if not LaneMetadata.is_valid_id(baseline_id):
		_request_resync(lane, state, ResyncState.REASON_WRONG_BASELINE)
		return false
	var normalized_chunks := _normalize_chunk_metadata(chunk_index, chunk_count, is_final_chunk)
	if normalized_chunks.is_empty():
		_request_resync(lane, state, ResyncState.REASON_STALE_OR_INVALID_SEQUENCE)
		return false
	if not _accept_full_start(state, baseline_id, normalized_sequence, normalized_chunks.chunk_index):
		return false
	_apply_full_chunk(state, baseline_id, normalized_sequence, snapshot_id, normalized_chunks.chunk_index, normalized_chunks.chunk_count, normalized_chunks.is_final_chunk)
	_refresh_readiness_for_lane(lane)
	return true

func record_full_chunk(lane: String, baseline_id, sequence, snapshot_id = null, chunk_index = 0, chunk_count = 1, is_final_chunk = false) -> bool:
	if not _is_supported_lane(lane):
		return false
	var state := _ensure_lane_state(lane)
	var normalized_sequence = LaneMetadata.parse_non_negative_integer(sequence)
	if normalized_sequence == null:
		_request_resync(lane, state, ResyncState.REASON_STALE_OR_INVALID_SEQUENCE)
		return false
	if not LaneMetadata.is_valid_id(baseline_id):
		_request_resync(lane, state, ResyncState.REASON_WRONG_BASELINE)
		return false
	var normalized_chunks := _normalize_chunk_metadata(chunk_index, chunk_count, is_final_chunk)
	if normalized_chunks.is_empty():
		_request_resync(lane, state, ResyncState.REASON_STALE_OR_INVALID_SEQUENCE)
		return false
	if not state.pending_full:
		if not _accept_full_start(state, baseline_id, normalized_sequence, normalized_chunks.chunk_index):
			return false
	else:
		if baseline_id != state.baseline_id or normalized_sequence != state.sequence:
			return false
		if normalized_chunks.chunk_count != state.chunk_count or normalized_chunks.chunk_index != state.chunk_index + 1:
			return false
	_apply_full_chunk(state, baseline_id, normalized_sequence, snapshot_id, normalized_chunks.chunk_index, normalized_chunks.chunk_count, normalized_chunks.is_final_chunk)
	_refresh_readiness_for_lane(lane)
	return true

func record_delta(lane: String, baseline_id, sequence, snapshot_id = null) -> bool:
	if not _is_supported_lane(lane):
		return false
	var state := _ensure_lane_state(lane)
	var normalized_sequence = LaneMetadata.parse_non_negative_integer(sequence)
	if normalized_sequence == null:
		_request_resync(lane, state, ResyncState.REASON_STALE_OR_INVALID_SEQUENCE)
		return false
	if not LaneMetadata.is_valid_id(baseline_id):
		_request_resync(lane, state, ResyncState.REASON_WRONG_BASELINE)
		return false
	if not state.full_packet_seen or state.baseline_id == null or str(state.baseline_id).is_empty():
		_request_resync(lane, state, ResyncState.REASON_MISSING_BASELINE)
		return false
	if state.sync_state == LaneSyncState.UNSYNCED or _is_stale_sequence(state, normalized_sequence):
		return false
	if state.baseline_id != baseline_id:
		_request_resync(lane, state, ResyncState.REASON_WRONG_BASELINE)
		return false
	state.sequence = normalized_sequence
	state.snapshot_id = snapshot_id
	return true

func mark_lane_unsynced(lane: String) -> void:
	if not _is_supported_lane(lane):
		return
	var state := _ensure_lane_state(lane)
	state.sync_state = LaneSyncState.UNSYNCED
	state.resync_needed = true
	_refresh_readiness_for_lane(lane)

func needs_resync(lane: String) -> bool:
	if not _is_supported_lane(lane):
		return false
	return _ensure_lane_state(lane).resync_needed

func request_resync_for_lane(lane: String, reason: String) -> void:
	if not _is_supported_lane(lane):
		return
	_request_resync(lane, _ensure_lane_state(lane), reason)

func _is_supported_lane(lane: String) -> bool:
	return LaneMetadata.is_supported_baseline_lane(lane)

func _ensure_lane_state(lane: String) -> Dictionary:
	return _lane_state[lane]

func _make_lane_state() -> Dictionary:
	return {"baseline_id": null, "snapshot_id": null, "sequence": null, "chunk_index": -1, "chunk_count": 0, "is_final_chunk": false, "sync_state": LaneSyncState.UNSYNCED, "resync_needed": false, "pending_full": false, "full_packet_seen": false}

func _is_stale_sequence(state: Dictionary, sequence) -> bool:
	return state.sequence != null and sequence <= state.sequence

func _normalize_chunk_metadata(chunk_index, chunk_count, is_final_chunk) -> Dictionary:
	var normalized_index = LaneMetadata.parse_non_negative_integer(chunk_index)
	var normalized_count = LaneMetadata.parse_non_negative_integer(chunk_count)
	if normalized_index == null or normalized_count == null or normalized_count < 1 or normalized_index >= normalized_count or typeof(is_final_chunk) != TYPE_BOOL:
		return {}
	if is_final_chunk != (normalized_index == normalized_count - 1):
		return {}
	return {"chunk_index": normalized_index, "chunk_count": normalized_count, "is_final_chunk": is_final_chunk}

func _accept_full_start(state: Dictionary, baseline_id, sequence, chunk_index: int) -> bool:
	if chunk_index != 0 or state.pending_full:
		return false
	if state.baseline_id != null and baseline_id != state.baseline_id:
		return state.resync_needed
	return state.sequence == null or sequence > state.sequence

func _apply_full_chunk(state: Dictionary, baseline_id, sequence, snapshot_id, chunk_index: int, chunk_count: int, is_final_chunk: bool) -> void:
	state.baseline_id = baseline_id
	state.snapshot_id = snapshot_id
	state.sequence = sequence
	state.chunk_index = chunk_index
	state.chunk_count = chunk_count
	state.is_final_chunk = is_final_chunk
	state.pending_full = not is_final_chunk
	state.full_packet_seen = true
	state.sync_state = LaneSyncState.SYNCED if is_final_chunk else LaneSyncState.SYNCING
	if is_final_chunk:
		state.resync_needed = false

func _request_resync(lane: String, state: Dictionary, reason: String) -> void:
	var was_pending: bool = state.resync_needed
	var previous_baseline = state.baseline_id if state.baseline_id is String else ""
	var previous_sequence = state.sequence
	state.sync_state = LaneSyncState.UNSYNCED
	state.resync_needed = true
	state.pending_full = false
	_refresh_readiness_for_lane(lane)
	if not was_pending:
		resync_required.emit(lane, previous_baseline, previous_sequence, reason)

func _refresh_readiness() -> void:
	for lane in LANES:
		_refresh_readiness_for_lane(lane)

func _refresh_readiness_for_lane(lane: String) -> void:
	if _readiness != null:
		_readiness.set_lane_synced(lane, is_lane_synced(lane))
