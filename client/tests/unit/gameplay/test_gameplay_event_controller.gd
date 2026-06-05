extends GutTest

const GameplayEventController = preload("res://scripts/gameplay/events/gameplay_event_controller.gd")
const Packets = preload("res://scripts/generated/networking/packets/packets.gd")


class FakeEffects:
	var spawn_pickup_collected_call_count := 0
	var last_visual_position := Vector2.ZERO

	func spawn_pickup_collected(visual_position: Vector2) -> void:
		spawn_pickup_collected_call_count += 1
		last_visual_position = visual_position


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
			Packets.FIELD_X: 12,
			Packets.FIELD_Y: 34,
		}
	]

	controller.apply_server_events(server_events, "player-1", Callable())

	assert_eq(effects.spawn_pickup_collected_call_count, 1)
	assert_eq(effects.last_visual_position, Vector2(22, 54))
