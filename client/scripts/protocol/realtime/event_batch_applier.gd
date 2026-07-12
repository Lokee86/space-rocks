extends RefCounted

const RealtimeQuantize := preload("res://scripts/protocol/realtime/realtime_quantize.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
var _applied_batch_ids := {}
var _applied_event_ids := {}
var _applied_events := []
var _logged_applied_batch_ids := {}

func has_applied_batch(batch_id) -> bool:
	return _applied_batch_ids.has(batch_id)

func has_applied_event(event_id) -> bool:
	return _applied_event_ids.has(event_id)

func get_applied_events() -> Array:
	return _applied_events.duplicate(true)

func drain_applied_events() -> Array:
	var events := _applied_events.duplicate(true)
	_applied_events.clear()
	return events

func apply_event_batch(event_batch_packet: Dictionary, event_sink) -> bool:
	var batch_id = event_batch_packet.get("batch_id")
	var batch_already_applied := batch_id != null and _applied_batch_ids.has(batch_id)

	var events = event_batch_packet.get("events", [])
	var applied_any := false
	var newly_applied_events := []
	for event in events:
		var decoded_event := RealtimeQuantize.decode_event_record(event)
		var event_id = str(decoded_event.get("event_id", ""))
		if event_id != "" and _applied_event_ids.has(event_id):
			continue
		if batch_already_applied and event_id == "":
			continue
		if not _apply_event(event_sink, decoded_event):
			continue
		applied_any = true
		newly_applied_events.append(decoded_event)

	if batch_id != null:
		_applied_batch_ids[batch_id] = true
	if applied_any and batch_id != null and !_logged_applied_batch_ids.has(batch_id):
		_logged_applied_batch_ids[batch_id] = true
		var applied_event_types := []
		for event in newly_applied_events:
			applied_event_types.append(str(event.get("type", "")))
		ClientLogger.packets_event(
			ClientLogger.LEVEL_DEBUG,
			"event_batch_applied",
			"Applied new server events",
			{"batch_id": str(batch_id), "new_event_count": newly_applied_events.size(), "event_types": applied_event_types}
		)
	return applied_any

func _apply_event(event_sink, event: Dictionary) -> bool:
	var event_id = event.get("event_id")
	if event_id != null and _applied_event_ids.has(event_id):
		return false

	var event_type = event.get("type")
	var payload = event.get("payload", {})
	if event_sink != null and event_sink.has_method("handle_presentation_event"):
		event_sink.handle_presentation_event(event_type, payload, event)


	if event_id != null:
		_event_id_record(event_id)
	_applied_events.append(event.duplicate(true))
	return true

func _event_id_record(event_id) -> void:
	_applied_event_ids[event_id] = true

