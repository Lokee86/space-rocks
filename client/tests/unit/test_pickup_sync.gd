extends GutTest

const PickupSync := preload("res://scripts/world/pickup_sync.gd")


class FakePickupNode:
	extends Node2D


func test_pickup_position_entries_exposes_positions_and_metadata() -> void:
	var pickup_sync: PickupSync = PickupSync.new()
	var pickup_node := FakePickupNode.new()

	pickup_sync.pickup_nodes = {
		"pickup-1": pickup_node,
	}
	pickup_sync.pickup_types = {
		"pickup-1": "1_up",
	}
	pickup_sync.pickup_classes = {
		"pickup-1": "powerup",
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
	assert_eq(entry["pickup_class"], "powerup")
	assert_eq(entry["node"], pickup_node)
	pickup_node.free()


func test_pickup_position_entries_uses_empty_pickup_type_and_null_node_when_missing() -> void:
	var pickup_sync: PickupSync = PickupSync.new()
	pickup_sync.target_pickup_positions = {
		"pickup-2": Vector2(50, 60),
	}

	var entries: Dictionary = pickup_sync.pickup_position_entries()
	assert_true(entries.has("pickup-2"))

	var entry = entries["pickup-2"]
	assert_eq(entry["visual_position"], Vector2(50, 60))
	assert_eq(entry["server_position"], Vector2(50, 60))
	assert_eq(entry["pickup_type"], "")
	assert_eq(entry["pickup_class"], "")
	assert_null(entry["node"])


func test_powerup_pickup_class_creates_powerup_pickup_scene() -> void:
	var pickup_sync: PickupSync = PickupSync.new()
	var pickups_layer: Node2D = Node2D.new()
	add_child_autofree(pickups_layer)
	pickup_sync.configure(pickups_layer)

	var server_pickups := {
		"pickup-1": {
			"type": "1_up",
			"pickup_class": "powerup",
			"x": 12.0,
			"y": 34.0,
			"age_seconds": 1.5,
			"lifespan_seconds": 9.0,
		},
	}

	pickup_sync.apply(server_pickups, Vector2(100, 200), Vector2(100, 200))

	var pickup_node: Node = pickup_sync.pickup_nodes.get("pickup-1")
	assert_not_null(pickup_node)
	assert_true(pickup_node is Node2D)

	var badge: Node = pickup_node.get_node_or_null("Badge")
	assert_not_null(badge)

	var pickup_icon: CanvasItem = badge.get_node_or_null("1_up")
	assert_not_null(pickup_icon)
	assert_true(pickup_icon.visible)


func test_weapon_pickup_class_creates_weapon_pickup_scene() -> void:
	var pickup_sync: PickupSync = PickupSync.new()
	var pickups_layer: Node2D = Node2D.new()
	add_child_autofree(pickups_layer)
	pickup_sync.configure(pickups_layer)

	var server_pickups := {
		"pickup-2": {
			"type": "torpedo",
			"pickup_class": "weapon",
			"x": 24.0,
			"y": 48.0,
			"age_seconds": 2.0,
			"lifespan_seconds": 11.0,
		},
	}

	pickup_sync.apply(server_pickups, Vector2(100, 200), Vector2(100, 200))

	var pickup_node: Node = pickup_sync.pickup_nodes.get("pickup-2")
	assert_not_null(pickup_node)
	assert_true(pickup_node is Node2D)

	var badge: Node = pickup_node.get_node_or_null("Badge")
	assert_not_null(badge)

	var pickup_icon: CanvasItem = badge.get_node_or_null("torpedo")
	assert_not_null(pickup_icon)
	assert_true(pickup_icon.visible)
