extends RefCounted

var _applied_batch_ids := {}
var _applied_event_ids := {}

func has_applied_batch(batch_id) -> bool:
	return _applied_batch_ids.has(batch_id)

func has_applied_event(event_id) -> bool:
	return _applied_event_ids.has(event_id)

func apply_event_batch(event_batch_packet: Dictionary, event_sink) -> bool:
	var batch_id = event_batch_packet.get("batch_id")
	if batch_id != null and _applied_batch_ids.has(batch_id):
		return false

	var events = event_batch_packet.get("events", [])
	var applied_any := false
	for event in events:
		if not _apply_event(event_sink, event):
			continue
		applied_any = true

	if batch_id != null:
		_applied_batch_ids[batch_id] = true
	return applied_any

func _apply_event(event_sink, event: Dictionary) -> bool:
	var event_id = event.get("event_id")
	if event_id != null and _applied_event_ids.has(event_id):
		return false

	var event_type = event.get("type")
	var payload = event.get("payload", {})
	if event_sink != null and event_sink.has_method("handle_presentation_event"):
		event_sink.handle_presentation_event(event_type, payload, event)
	elif event_sink != null and event_sink.has_method("apply_presentation_event"):
		event_sink.apply_presentation_event(event_type, payload, event)

	if event_id != null:
		_applied_event_ids[event_id] = true
	return true

