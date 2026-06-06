extends GutTest

const PickupSync := preload("res://scripts/world/pickup_sync.gd")


class FakePickupNode:
	extends Node2D


func test_pickup_position_entries_exposes_positions_and_metadata() -> void:
	var pickup_sync = PickupSync.new()
	var pickup_node := FakePickupNode.new()

	pickup_sync.pickup_nodes = {
		"pickup-1": pickup_node,
	}
	pickup_sync.pickup_types = {
		"pickup-1": "1_up",
	}
	pickup_sync.target_pickup_positions = {
		"pickup-1": Vector2(10, 20),
	}
	pickup_sync.pickup_server_positions = {
		"pickup-1": Vector2(30, 40),
	}

	var entries: Dictionary = pickup_sync.pickup_position_entries()
	assert_true(entries.has("pickup-1"))

	var entry = entries["pickup-1"]
	assert_eq(entry["visual_position"], Vector2(10, 20))
	assert_eq(entry["server_position"], Vector2(30, 40))
	assert_eq(entry["pickup_type"], "1_up")
	assert_eq(entry["node"], pickup_node)
	pickup_node.free()


func test_pickup_position_entries_uses_empty_pickup_type_and_null_node_when_missing() -> void:
	var pickup_sync = PickupSync.new()
	pickup_sync.target_pickup_positions = {
		"pickup-2": Vector2(50, 60),
	}

	var entries: Dictionary = pickup_sync.pickup_position_entries()
	assert_true(entries.has("pickup-2"))

	var entry = entries["pickup-2"]
	assert_eq(entry["visual_position"], Vector2(50, 60))
	assert_eq(entry["server_position"], Vector2(50, 60))
	assert_eq(entry["pickup_type"], "")
	assert_null(entry["node"])
