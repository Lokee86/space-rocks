extends GutTest

const GameplayEventController = preload("res://scripts/gameplay/events/gameplay_event_controller.gd")
const Packets = preload("res://scripts/generated/networking/packets/packets.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const Contract := preload("res://scripts/generated/observability/contract_generated.gd")
const EventCapture := preload("res://tests/unit/logging/presentation_event_capture.gd")


class FakeEffects:
	var spawn_pickup_collected_call_count := 0
	var spawn_torpedo_explosion_call_count := 0
	var last_visual_position := Vector2.ZERO

	func spawn_pickup_collected(visual_position: Vector2) -> void:
		spawn_pickup_collected_call_count += 1
		last_visual_position = visual_position

	func spawn_torpedo_explosion(visual_position: Vector2) -> void:
		spawn_torpedo_explosion_call_count += 1
		last_visual_position = visual_position


func test_missing_visual_position_converter_emits_canonical_state_invalid() -> void:
	var capture := _begin_capture()
	var controller := GameplayEventController.new()

	controller.apply_bullet_blast({
		Packets.FIELD_X: 12.0,
		Packets.FIELD_Y: 34.0,
	})

	var record := capture.last_record()
	assert_eq(record["event"], Contract.EVENT_CLIENT_PRESENTATION_STATE_INVALID)
	assert_eq(record["fields"]["subsystem"], "gameplay_events")
	assert_eq(record["fields"]["failure_mode"], "missing_visual_position_converter")
	assert_eq(record["fields"]["event_kind"], "bullet blast")
	assert_eq(record["fields"]["field_name"], "visual_position_converter")


func _begin_capture() -> EventCapture:
	var capture := EventCapture.new()
	ClientLogger._set_file_writer_for_tests(capture)
	return capture


func after_each() -> void:
	ClientLogger.reset_for_tests()


class FakeCallableTarget:
	func visual_position_for_server_position(server_position: Vector2) -> Vector2:
		return server_position + Vector2(10, 20)


func test_apply_server_events_routes_pickup_collected_to_effects() -> void:
	var effects := FakeEffects.new()
	var callable_target := FakeCallableTarget.new()
	var controller := GameplayEventController.new()
	controller.configure(effects, Callable(callable_target, "visual_position_for_server_position"))

	var server_events := [
		{
			Packets.FIELD_TYPE: "pickup_collected",
			Packets.FIELD_X: 12.3,
			Packets.FIELD_Y: 34.5,
		}
	]

	controller.apply_server_events(server_events, "player-1", Callable())

	assert_eq(effects.spawn_pickup_collected_call_count, 1)
	assert_eq(effects.last_visual_position, Vector2(22.3, 54.5))


func test_apply_server_events_routes_radial_effect_started_to_torpedo_explosion() -> void:
	var effects := FakeEffects.new()
	var callable_target := FakeCallableTarget.new()
	var controller := GameplayEventController.new()
	controller.configure(effects, Callable(callable_target, "visual_position_for_server_position"))

	var server_events := [
		{
			Packets.FIELD_TYPE: "radial_effect_started",
			Packets.FIELD_X: 40,
			Packets.FIELD_Y: 60,
		}
	]

	controller.apply_server_events(server_events, "player-1", Callable())

	assert_eq(effects.spawn_torpedo_explosion_call_count, 1)
	assert_eq(effects.last_visual_position, Vector2(50, 80))
