extends RefCounted

const DescriptorDecoder := preload("res://scripts/protocol/realtime/compact_wire_descriptor_decoder.gd")
const DescriptorIDs := preload("res://scripts/protocol/realtime/compact_wire_descriptor_ids.gd")

static func expand_packet(packet: Dictionary) -> Dictionary:
	return DescriptorDecoder.expand_packet(packet)

static func expand_compact_asteroid_id(value):
	return DescriptorIDs.expand_codec("asteroid_id", value)

static func expand_compact_pickup_id(value):
	return DescriptorIDs.expand_codec("pickup_id", value)

static func expand_compact_table_id(value):
	return DescriptorIDs.expand_codec("table_id", value)

static func expand_compact_ship_id(value):
	return DescriptorIDs.expand_codec("ship_id", value)

static func expand_compact_bullet_id(value):
	return DescriptorIDs.expand_codec("bullet_id", value)

static func expand_compact_presentation_event_id(value):
	return DescriptorIDs.expand_codec("presentation_event_id", value)

static func expand_compact_event_batch_id(value):
	return DescriptorIDs.expand_codec("event_batch_id", value)

static func expand_compact_player_id(value):
	return DescriptorIDs.expand_codec("player_id", value)

static func expand_compact_tagged_id(value):
	return DescriptorIDs.expand_tagged(value)

static func expand_compact_prefixed_id(value, prefix):
	if value is int:
		return "%s%d" % [prefix, value]
	if value is float:
		if is_equal_approx(value, floor(value)):
			return "%s%d" % [prefix, int(value)]
		return value
	if value is String:
		if value.begins_with(prefix):
			return value
		if value.is_valid_int():
			return "%s%s" % [prefix, value]
	return value
