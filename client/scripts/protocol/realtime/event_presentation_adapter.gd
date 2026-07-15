extends RefCounted
class_name EventPresentationAdapter

const GameplayEventLifecycleFlow := preload("res://scripts/gameplay/events/gameplay_event_lifecycle_flow.gd")
const EventBatchApplier := preload("res://scripts/protocol/realtime/event_batch_applier.gd")

func apply_event_batch_output(event_flow: GameplayEventLifecycleFlow, event_batch_applier: EventBatchApplier, self_id: String) -> void:
	if event_flow == null or event_batch_applier == null:
		return
	var events: Array = event_batch_applier.drain_applied_events()
	if events.is_empty():
		return
	event_flow.apply_server_events(events, self_id)

