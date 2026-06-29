extends RefCounted

func apply_event_batch_output(event_flow, event_batch_applier, self_id: String) -> void:
	if event_flow == null or event_batch_applier == null:
		return
	if not event_flow.has_method("apply_server_events"):
		return
	var events: Array = []
	if event_batch_applier.has_method("get_applied_events"):
		events = event_batch_applier.get_applied_events()
	if events.is_empty():
		return
	event_flow.apply_server_events(events, self_id)