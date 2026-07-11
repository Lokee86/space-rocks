class_name CompactWireDescriptorIDs
extends RefCounted

const DescriptorIndex := preload("res://scripts/protocol/realtime/compact_wire_descriptor_index.gd")

static func expand_codec(codec_id: String, value):
	var codec := DescriptorIndex.codec_by_id(codec_id)
	if codec.is_empty():
		return value
	if value is Array:
		if value.size() != 2 or str(value[0]) != str(codec.get("tag", "")):
			return value
		return _expand_suffix(codec, value[1])
	return _expand_suffix(codec, value)

static func expand_selector(selector_id: String, selector_value, value):
	var selector := DescriptorIndex.selector_by_id(selector_id)
	if selector.is_empty():
		return value
	for mapping in selector.get("mappings", []):
		if str(mapping.get("value", "")) == str(selector_value):
			return expand_codec(str(mapping.get("codec_id", "")), value)
	if selector.get("fallback_tagged", false):
		return expand_tagged(value)
	return value

static func expand_tagged(value):
	if not (value is Array and value.size() == 2):
		return value
	var codec := DescriptorIndex.codec_by_tag(str(value[0]))
	if codec.is_empty():
		return value
	return _expand_suffix(codec, value[1])

static func _expand_suffix(codec: Dictionary, value):
	var prefix := str(codec.get("prefix", ""))
	if value == null or prefix == "":
		return value
	if value is String:
		if value.begins_with(prefix):
			return value
		if value.is_valid_int() and codec.get("numeric_suffix", true):
			return prefix + value
		return value
	if value is int:
		if codec.get("numeric_suffix", true):
			return "%s%d" % [prefix, value]
		return value
	if value is float:
		if codec.get("numeric_suffix", true) and is_finite(value) and is_equal_approx(value, floor(value)):
			return "%s%d" % [prefix, int(value)]
		return value
	return value
