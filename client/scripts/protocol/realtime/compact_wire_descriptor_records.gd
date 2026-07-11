class_name CompactWireDescriptorRecords
extends RefCounted

const DescriptorIndex := preload("res://scripts/protocol/realtime/compact_wire_descriptor_index.gd")
const DescriptorIDs := preload("res://scripts/protocol/realtime/compact_wire_descriptor_ids.gd")

static func expand_generic(value, parent_key = ""):
	if value is Array:
		var expanded_array: Array = []
		for item in value:
			expanded_array.append(expand_generic(item, parent_key))
		return expanded_array
	if not (value is Dictionary):
		if parent_key != "":
			var scalar_record_ids: Array = DescriptorIndex.scalar_binding_record_ids(str(parent_key))
			if not scalar_record_ids.is_empty():
				return expand_bound(value, scalar_record_ids)
		if parent_key == "type":
			return _expand_type(value)
		if parent_key == "lane":
			return DescriptorIndex.readable_lane(str(value))
		if parent_key == "snapshot_kind":
			return DescriptorIndex.readable_snapshot_kind(str(value))
		return value
	var expanded: Dictionary = {}
	for raw_key in value:
		var readable_key := DescriptorIndex.readable_key(str(raw_key))
		expanded[readable_key] = expand_generic(value[raw_key], readable_key)
	return expanded

static func expand_bound(value, record_ids: Array):
	if record_ids.is_empty():
		return expand_generic(value)
	var record: Dictionary = DescriptorIndex.record_by_id(str(record_ids[0]))
	if record.is_empty():
		return expand_generic(value)
	var mode := str(record.get("encoding", ""))
	if mode == "discriminated_event_tuple":
		if not (value is Array):
			return expand_generic(value)
		var events: Array = []
		for item in value:
			events.append(_expand_event(item))
		return events
	var records: Array = []
	for record_id in record_ids:
		var candidate: Dictionary = DescriptorIndex.record_by_id(str(record_id))
		if not candidate.is_empty():
			records.append(candidate)
	var scalar_record: Dictionary = _scalar_record(records)
	if not scalar_record.is_empty():
		var scalar_mode := str(scalar_record.get("encoding", ""))
		var scalar_fields: Array = scalar_record.get("fields", [])
		if scalar_mode == "scalar_id" and not scalar_fields.is_empty():
			return _expand_field(value, scalar_fields[0], {})
		if scalar_mode == "scalar_id_list":
			if not (value is Array) or scalar_fields.is_empty():
				return expand_generic(value)
			var ids: Array = []
			for item in value:
				ids.append(_expand_field(item, scalar_fields[0], {}))
			return ids
	if mode == "scalar_list":
		if not (value is Array):
			return expand_generic(value)
		return expand_generic(value)
	if value is Dictionary:
		var map_record: Dictionary = _record_for_value(value, records)
		return expand_record(value, map_record) if not map_record.is_empty() else expand_generic(value)
	if value is Array and _is_record_list(value, records):
		var expanded_records: Array = []
		for item in value:
			var item_record: Dictionary = _record_for_value(item, records)
			expanded_records.append(expand_record(item, item_record) if not item_record.is_empty() else expand_generic(item))
		return expanded_records
	var tuple_record: Dictionary = _record_for_value(value, records)
	return expand_record(value, tuple_record) if not tuple_record.is_empty() else expand_generic(value)

static func _scalar_record(records: Array) -> Dictionary:
	for record in records:
		if record.get("encoding", "") == "scalar_id_list":
			return record
	for record in records:
		if record.get("encoding", "") == "scalar_id":
			return record
	return {}

static func _is_record_list(value: Array, records: Array) -> bool:
	if value.is_empty():
		return true
	for record in records:
		var encoding := str(record.get("encoding", ""))
		if value[0] is Dictionary and encoding not in ["scalar_id", "scalar_id_list", "scalar_list", "discriminated_event_tuple"]:
			return true
		if value[0] is Array and encoding in ["fixed_tuple", "sparse_positional_tuple", "sparse_key_value_tuple"]:
			return true
	return false

static func _record_for_value(value, records: Array) -> Dictionary:
	if value is Dictionary:
		for record in records:
			if record.get("encoding", "") == "map":
				return record
		for record in records:
			if record.get("encoding", "") in ["fixed_tuple", "sparse_positional_tuple", "sparse_key_value_tuple"]:
				return record
	if value is Array:
		for record in records:
			if record.get("encoding", "") in ["fixed_tuple", "sparse_positional_tuple", "sparse_key_value_tuple"]:
				return record
	return {}

static func expand_record(value, record: Dictionary):
	if value is Dictionary:
		return _expand_map(value, record)
	if not (value is Array):
		return expand_generic(value)
	var mode := str(record.get("encoding", ""))
	match mode:
		"fixed_tuple":
			return _expand_fixed_tuple(value, record)
		"sparse_positional_tuple":
			return _expand_sparse_positional(value, record)
		"sparse_key_value_tuple":
			return _expand_sparse_key_value(value, record)
		_:
			return expand_generic(value)

static func _expand_event(value):
	if not (value is Array) or value.is_empty():
		return expand_generic(value)
	var event := DescriptorIndex.event_by_compact_type(str(value[0]))
	if event.is_empty():
		return expand_generic(value)
	var record := DescriptorIndex.record_by_id(str(event.get("record_id", "")))
	if record.is_empty():
		return expand_generic(value)
	return expand_record(value, record)

static func _expand_map(value: Dictionary, record: Dictionary) -> Dictionary:
	var expanded: Dictionary = expand_generic(value)
	for field in record.get("fields", []):
		var field_name := str(field.get("json", ""))
		if expanded.has(field_name):
			expanded[field_name] = _expand_field(expanded[field_name], field, expanded)
	return expanded

static func _expand_fixed_tuple(value: Array, record: Dictionary) -> Dictionary:
	var expanded: Dictionary = {}
	var fields: Array = record.get("fields", [])
	for index in range(fields.size()):
		var field: Dictionary = fields[index]
		var item = value[index] if index < value.size() else null
		expanded[str(field.get("json", ""))] = _expand_field(item, field, expanded)
	return expanded

static func _expand_sparse_positional(value: Array, record: Dictionary) -> Dictionary:
	var expanded: Dictionary = {}
	var fields: Array = record.get("fields", [])
	for index in range(fields.size()):
		if index >= value.size():
			break
		var item = value[index]
		if index > 0 and item == null:
			continue
		var field: Dictionary = fields[index]
		expanded[str(field.get("json", ""))] = _expand_field(item, field, expanded)
	return expanded

static func _expand_sparse_key_value(value: Array, record: Dictionary) -> Dictionary:
	var expanded: Dictionary = {}
	var fields_by_name: Dictionary = {}
	var fields: Array = record.get("fields", [])
	for field in fields:
		fields_by_name[str(field.get("json", ""))] = field
	if not value.is_empty() and not fields.is_empty():
		expanded[str(fields[0].get("json", ""))] = _expand_field(value[0], fields[0], expanded)
	var index := 1
	while index + 1 < value.size():
		var readable_key := DescriptorIndex.readable_key(str(value[index]))
		var field: Dictionary = fields_by_name.get(readable_key, {})
		var item = expand_generic(value[index + 1], readable_key)
		if not field.is_empty():
			item = _expand_field(item, field, expanded)
		expanded[readable_key] = item
		index += 2
	return expanded

static func _expand_field(value, field: Dictionary, siblings: Dictionary):
	var codec_by_value = field.get("id_codec_by")
	if codec_by_value != null and str(codec_by_value) != "":
		var codec_by := str(codec_by_value)
		return DescriptorIDs.expand_selector(codec_by, siblings.get(codec_by), value)
	var codec_id_value = field.get("id_codec")
	if codec_id_value != null and str(codec_id_value) != "":
		var codec_id := str(codec_id_value)
		return DescriptorIDs.expand_codec(codec_id, value)
	var domain_value = field.get("value_domain")
	if domain_value != null and str(domain_value) != "":
		var domain := str(domain_value)
		return DescriptorIndex.readable_value(domain, str(value))
	return value

static func _expand_type(value):
	var packet_type := DescriptorIndex.readable_packet_type(str(value))
	if packet_type != str(value):
		return packet_type
	return DescriptorIndex.readable_event_type(str(value))
