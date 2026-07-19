class_name CompactWireDescriptorDecoder
extends RefCounted

const DescriptorIndex := preload("res://scripts/protocol/realtime/compact_wire_descriptor_index.gd")
const DescriptorRecords := preload("res://scripts/protocol/realtime/compact_wire_descriptor_records.gd")

static func expand_packet(packet: Dictionary) -> Dictionary:
	var packet_type := _readable_packet_type(packet)
	var descriptor := DescriptorIndex.packet_by_readable_id(packet_type)
	var expanded: Dictionary = {}
	for raw_key in packet:
		var readable_key := DescriptorIndex.readable_key(str(raw_key))
		var value = packet[raw_key]
		var record_ids: Array = DescriptorIndex.binding_record_ids(packet_type, readable_key)
		if not descriptor.is_empty() and not record_ids.is_empty():
			expanded[readable_key] = DescriptorRecords.expand_bound(value, record_ids)
		else:
			expanded[readable_key] = DescriptorRecords.expand_generic(value, readable_key)
	_normalize_metadata(expanded, descriptor)
	return expanded

static func _readable_packet_type(packet: Dictionary) -> String:
	var raw_type = packet.get("type", packet.get("t"))
	return DescriptorIndex.readable_packet_type(str(raw_type))

static func _normalize_metadata(packet: Dictionary, descriptor: Dictionary) -> void:
	if descriptor.is_empty() or not descriptor.get("runtime", false):
		return
	var lane = packet.get("lane")
	if lane == null and descriptor.get("infer_lane", false):
		lane = str(descriptor.get("lane", ""))
		if lane != "":
			packet["lane"] = lane
	var snapshot_kind = packet.get("snapshot_kind")
	if snapshot_kind == null and descriptor.get("infer_snapshot_kind", false):
		snapshot_kind = str(descriptor.get("snapshot_kind", ""))
		if snapshot_kind != "":
			packet["snapshot_kind"] = snapshot_kind
	var sequence = packet.get("sequence")
	if lane != null and sequence != null:
		if packet.get("snapshot_id") == null and descriptor.get("infer_snapshot_id", false):
			var snapshot_id := _default_snapshot_id(str(lane), str(snapshot_kind), sequence)
			if snapshot_id != "":
				packet["snapshot_id"] = snapshot_id
		if packet.get("baseline_id") == null and (descriptor.get("infer_baseline_id", false) or descriptor.get("use_baseline_sequence", false)):
			var baseline_id := _default_baseline_id(str(lane), str(snapshot_kind), sequence, packet.get("baseline_sequence"))
			if baseline_id != "":
				packet["baseline_id"] = baseline_id
	if packet.get("chunk_index") == null:
		packet["chunk_index"] = 0
	if packet.get("chunk_count") == null:
		packet["chunk_count"] = 1
	if packet.get("is_final_chunk") == null:
		var chunk_index: int = int(packet.get("chunk_index", 0))
		var chunk_count: int = int(packet.get("chunk_count", 1))
		packet["is_final_chunk"] = chunk_count <= 1 or chunk_index == chunk_count - 1

static func _default_snapshot_id(lane: String, snapshot_kind: String, sequence) -> String:
	if snapshot_kind == "full":
		return "%s-baseline-%s" % [lane, _sequence_suffix(sequence)]
	if snapshot_kind == "delta":
		return "%s-snapshot-%s" % [lane, _sequence_suffix(sequence)]
	return ""

static func _default_baseline_id(lane: String, snapshot_kind: String, sequence, baseline_sequence) -> String:
	if snapshot_kind == "full":
		return "%s-baseline-%s" % [lane, _sequence_suffix(sequence)]
	if snapshot_kind == "delta" and baseline_sequence != null:
		return "%s-baseline-%s" % [lane, _sequence_suffix(baseline_sequence)]
	return ""

static func _sequence_suffix(value) -> String:
	if value is int:
		return str(value)
	if value is float and is_finite(value) and is_equal_approx(value, floor(value)):
		return str(int(value))
	return str(value)
