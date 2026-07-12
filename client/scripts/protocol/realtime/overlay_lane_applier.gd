extends RefCounted

const OverlayLaneState = preload("res://scripts/protocol/realtime/overlay_lane_state.gd")
const BaselineTracker = preload("res://scripts/protocol/realtime/baseline_tracker.gd")

func apply_overlay_full(overlay_lane_state: OverlayLaneState, baseline_tracker: BaselineTracker, lane: String, overlay_packet: Dictionary) -> bool:
	var baseline_id = overlay_packet.get("baseline_id")
	var sequence = overlay_packet.get("sequence")
	var snapshot_id = overlay_packet.get("snapshot_id")
	var chunk_index = overlay_packet.get("chunk_index", 0)
	var chunk_count = overlay_packet.get("chunk_count", 1)
	var is_final_chunk = overlay_packet.get("is_final_chunk", true)

	var accepted: bool
	if is_final_chunk:
		accepted = baseline_tracker.record_full_packet(lane, baseline_id, sequence, snapshot_id, chunk_index, chunk_count, true)
	else:
		accepted = baseline_tracker.record_full_chunk(lane, baseline_id, sequence, snapshot_id, chunk_index, chunk_count, false)
	if accepted:
		overlay_lane_state.apply_full_overlay(overlay_packet)
	return accepted

func apply_overlay_delta(overlay_lane_state: OverlayLaneState, baseline_tracker: BaselineTracker, lane: String, overlay_packet: Dictionary) -> bool:
	var baseline_id = overlay_packet.get("baseline_id")
	var sequence = overlay_packet.get("sequence")
	var snapshot_id = overlay_packet.get("snapshot_id")

	if not baseline_tracker.record_delta(lane, baseline_id, sequence, snapshot_id):
		return false

	overlay_lane_state.apply_overlay_delta(overlay_packet)
	return true

