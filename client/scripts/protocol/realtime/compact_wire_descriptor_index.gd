class_name CompactWireDescriptorIndex
extends RefCounted



static var _initialized := false
static var _records_by_id: Dictionary = {}
static var _packets_by_readable_id: Dictionary = {}
static var _packets_by_compact_id: Dictionary = {}
static var _bindings_by_key: Dictionary = {}
static var _scalar_bindings_by_field: Dictionary = {}
static var _codecs_by_id: Dictionary = {}
static var _codecs_by_tag: Dictionary = {}
static var _selectors_by_id: Dictionary = {}
static var _events_by_readable: Dictionary = {}
static var _events_by_compact: Dictionary = {}

static func _ensure_initialized() -> void:
	if _initialized:
		return
	_initialized = true
	var contract: Dictionary = RealtimeWireGenerated.CONTRACT
	for record in contract.get("records", []):
		_records_by_id[record.get("id", "")] = record
	for packet in contract.get("packets", []):
		_packets_by_readable_id[packet.get("id", "")] = packet
		_packets_by_compact_id[packet.get("compact", "")] = packet
	for binding_key in RealtimeWireGenerated.PACKET_FIELD_RECORD_IDS:
		var record_ids: Array = RealtimeWireGenerated.PACKET_FIELD_RECORD_IDS[binding_key]
		_bindings_by_key[binding_key] = record_ids
		if record_ids.size() != 1:
			continue
		var parts: PackedStringArray = str(binding_key).split(".")
		if parts.is_empty():
			continue
		var record: Dictionary = _records_by_id.get(str(record_ids[0]), {})
		if record.get("encoding", "") == "scalar_id":
			var field := parts[parts.size() - 1]
			if not _scalar_bindings_by_field.has(field):
				_scalar_bindings_by_field[field] = []
			_scalar_bindings_by_field[field].append(str(record_ids[0]))
	for codec in contract.get("id_codecs", []):
		_codecs_by_id[codec.get("id", "")] = codec
		var tag = codec.get("tag")
		if tag != null and str(tag) != "":
			_codecs_by_tag[str(tag)] = codec
	for selector in contract.get("id_selectors", []):
		_selectors_by_id[selector.get("id", "")] = selector
	for event_type in RealtimeWireGenerated.EVENTS_BY_READABLE:
		_events_by_readable[event_type] = RealtimeWireGenerated.EVENTS_BY_READABLE[event_type]
	for event_type in RealtimeWireGenerated.EVENTS_BY_COMPACT:
		_events_by_compact[event_type] = RealtimeWireGenerated.EVENTS_BY_COMPACT[event_type]

static func record_by_id(record_id: String) -> Dictionary:
	_ensure_initialized()
	return _records_by_id.get(record_id, {})

static func packet_by_readable_id(packet_id: String) -> Dictionary:
	_ensure_initialized()
	return _packets_by_readable_id.get(packet_id, {})

static func packet_by_compact_id(packet_id: String) -> Dictionary:
	_ensure_initialized()
	return _packets_by_compact_id.get(packet_id, {})

static func binding_record_ids(packet_id: String, readable_field: String) -> Array:
	_ensure_initialized()
	return _bindings_by_key.get(packet_id + "." + readable_field, [])

static func scalar_binding_record_ids(readable_field: String) -> Array:
	_ensure_initialized()
	return _scalar_bindings_by_field.get(readable_field, [])

static func codec_by_id(codec_id: String) -> Dictionary:
	_ensure_initialized()
	return _codecs_by_id.get(codec_id, {})

static func codec_by_tag(tag: String) -> Dictionary:
	_ensure_initialized()
	return _codecs_by_tag.get(tag, {})

static func selector_by_id(selector_id: String) -> Dictionary:
	_ensure_initialized()
	return _selectors_by_id.get(selector_id, {})

static func event_by_readable_type(event_type: String) -> Dictionary:
	_ensure_initialized()
	return _events_by_readable.get(event_type, {})

static func event_by_compact_type(event_type: String) -> Dictionary:
	_ensure_initialized()
	return _events_by_compact.get(event_type, {})

static func readable_key(compact_key_value: String) -> String:
	return RealtimeWireGenerated.KEY_READABLE_BY_COMPACT.get("wire." + compact_key_value, compact_key_value)

static func compact_key(readable_key_value: String) -> String:
	return RealtimeWireGenerated.KEY_COMPACT_BY_READABLE.get("wire." + readable_key_value, readable_key_value)

static func readable_value(domain: String, compact_value_value: String) -> String:
	return RealtimeWireGenerated.VALUE_READABLE_BY_COMPACT.get(domain, {}).get(compact_value_value, compact_value_value)

static func compact_value(domain: String, readable_value_value: String) -> String:
	return RealtimeWireGenerated.VALUE_COMPACT_BY_READABLE.get(domain, {}).get(readable_value_value, readable_value_value)

static func readable_packet_type(compact_type: String) -> String:
	return readable_value("packet_type", compact_type)

static func compact_packet_type(readable_type: String) -> String:
	return compact_value("packet_type", readable_type)

static func readable_event_type(compact_type: String) -> String:
	return readable_value("event_type", compact_type)

static func compact_event_type(readable_type: String) -> String:
	return compact_value("event_type", readable_type)

static func readable_lane(compact_lane_value: String) -> String:
	return readable_value("lane", compact_lane_value)

static func compact_lane(readable_lane_value: String) -> String:
	return compact_value("lane", readable_lane_value)

static func readable_snapshot_kind(compact_kind: String) -> String:
	return readable_value("snapshot_kind", compact_kind)

static func compact_snapshot_kind(readable_kind: String) -> String:
	return compact_value("snapshot_kind", readable_kind)
