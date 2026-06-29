extends RefCounted

const SessionLaneState = preload("res://scripts/protocol/realtime/session_lane_state.gd")
const BaselineTracker = preload("res://scripts/protocol/realtime/baseline_tracker.gd")

func apply_session_full(session_lane_state: SessionLaneState, baseline_tracker: BaselineTracker, lane: String, session_packet: Dictionary) -> void:
	var baseline_id = session_packet.get("baseline_id")
	var sequence = session_packet.get("sequence")
	var snapshot_id = session_packet.get("snapshot_id")
	var chunk_index: int = int(session_packet.get("chunk_index", 0))
	var chunk_count: int = int(session_packet.get("chunk_count", 1))
	var is_final_chunk: bool = bool(session_packet.get("is_final_chunk", true))

	session_lane_state.apply_full_session(session_packet)

	if is_final_chunk:
		baseline_tracker.record_full_packet(lane, baseline_id, sequence, snapshot_id, chunk_index, chunk_count, true)
	else:
		baseline_tracker.record_full_chunk(lane, baseline_id, sequence, snapshot_id, chunk_index, chunk_count, false)

func apply_session_delta(session_lane_state: SessionLaneState, baseline_tracker: BaselineTracker, lane: String, session_packet: Dictionary) -> bool:
	var baseline_id = session_packet.get("baseline_id")
	var sequence = session_packet.get("sequence")

	if not baseline_tracker.record_delta(lane, baseline_id, sequence):
		return false

	session_lane_state.apply_session_delta(session_packet)
	return true

