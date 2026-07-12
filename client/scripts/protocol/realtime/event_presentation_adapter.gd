extends RefCounted

const ClientLogger := preload("res://scripts/logging/logger.gd")

func apply_event_batch_output(event_flow, event_batch_applier, self_id: String) -> void:
	if event_flow == null or event_batch_applier == null:
		return
	if not event_flow.has_method("apply_server_events"):
		return
	var events: Array = []
	if event_batch_applier.has_method("drain_applied_events"):
		events = event_batch_applier.drain_applied_events()
	elif event_batch_applier.has_method("get_applied_events"):
		events = event_batch_applier.get_applied_events()
	if events.is_empty():
		return
	var event_types := []
	for event in events:
		event_types.append(str(event.get("type", "")))
	ClientLogger.packets_event(
		ClientLogger.LEVEL_DEBUG,
		"event_batch_forwarded",
		"Forwarding applied events to lifecycle",
		{"count": events.size(), "self_id": self_id, "event_types": event_types}
	)
	event_flow.apply_server_events(events, self_id)

