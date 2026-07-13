extends RefCounted
class_name EventPresentationAdapter

const ClientLogger := preload("res://scripts/logging/logger.gd")
const GameplayEventLifecycleFlow := preload("res://scripts/gameplay/events/gameplay_event_lifecycle_flow.gd")
const EventBatchApplier := preload("res://scripts/protocol/realtime/event_batch_applier.gd")

func apply_event_batch_output(event_flow: GameplayEventLifecycleFlow, event_batch_applier: EventBatchApplier, self_id: String) -> void:
	if event_flow == null or event_batch_applier == null:
		return
	var events: Array = event_batch_applier.drain_applied_events()
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

