extends GutTest

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
const ProjectileSync := preload("res://scripts/world/projectile_sync.gd")

func test_remove_projectile_returns_node_to_pool() -> void:
	var projectile_sync := _new_projectile_sync()

	projectile_sync.apply_projectile(
		"bullet-1",
		{
			Packets.FIELD_X: 10.0,
			Packets.FIELD_Y: 20.0,
			Packets.FIELD_ROTATION: 0.5,
			Packets.FIELD_PROJECTILE_TYPE: "bullet",
		},
		Vector2.ZERO,
		Vector2.ZERO
	)

	var node = projectile_sync.projectile_nodes["bullet-1"]

	projectile_sync.remove_projectile("bullet-1")

	assert_false(projectile_sync.projectile_nodes.has("bullet-1"))
	assert_eq(projectile_sync.pool_size(), 1)
	assert_eq(projectile_sync.pool_size_for_type("bullet"), 1)
	assert_false(node.visible)

func test_apply_projectile_reuses_pooled_node() -> void:
	var projectile_sync := _new_projectile_sync()

	projectile_sync.apply_projectile(
		"bullet-1",
		{
			Packets.FIELD_X: 10.0,
			Packets.FIELD_Y: 20.0,
			Packets.FIELD_ROTATION: 0.5,
			Packets.FIELD_PROJECTILE_TYPE: "bullet",
		},
		Vector2.ZERO,
		Vector2.ZERO
	)

	var old_node = projectile_sync.projectile_nodes["bullet-1"]
	projectile_sync.remove_projectile("bullet-1")

	projectile_sync.apply_projectile(
		"bullet-2",
		{
			Packets.FIELD_X: 30.0,
			Packets.FIELD_Y: 40.0,
			Packets.FIELD_ROTATION: 1.0,
			Packets.FIELD_PROJECTILE_TYPE: "bullet",
		},
		Vector2.ZERO,
		Vector2.ZERO
	)

	var new_node = projectile_sync.projectile_nodes["bullet-2"]

	assert_eq(new_node, old_node)
	assert_eq(projectile_sync.pool_size(), 0)
	assert_true(new_node.visible)
	assert_false(projectile_sync.projectile_nodes.has("bullet-1"))
	assert_true(projectile_sync.projectile_nodes.has("bullet-2"))

func test_projectile_pool_metrics_track_create_release_reuse() -> void:
	var projectile_sync := _new_projectile_sync()

	projectile_sync.apply_projectile(
		"bullet-1",
		{
			Packets.FIELD_X: 10.0,
			Packets.FIELD_Y: 20.0,
			Packets.FIELD_ROTATION: 0.5,
			Packets.FIELD_PROJECTILE_TYPE: "bullet",
		},
		Vector2.ZERO,
		Vector2.ZERO
	)
	projectile_sync.remove_projectile("bullet-1")
	projectile_sync.apply_projectile(
		"bullet-2",
		{
			Packets.FIELD_X: 30.0,
			Packets.FIELD_Y: 40.0,
			Packets.FIELD_ROTATION: 1.0,
			Packets.FIELD_PROJECTILE_TYPE: "bullet",
		},
		Vector2.ZERO,
		Vector2.ZERO
	)

	var metrics = projectile_sync.metrics_snapshot()

	assert_eq(metrics["active_projectile_nodes"], 1)
	assert_eq(metrics["pooled_projectile_nodes"], 0)
	assert_eq(metrics["created_projectile_node_count"], 1)
	assert_eq(metrics["released_projectile_node_count"], 1)
	assert_eq(metrics["reused_projectile_node_count"], 1)

func test_torpedo_does_not_reuse_bullet_pool_node() -> void:
	var projectile_sync := _new_projectile_sync()

	projectile_sync.apply_projectile(
		"bullet-1",
		{
			Packets.FIELD_X: 10.0,
			Packets.FIELD_Y: 20.0,
			Packets.FIELD_ROTATION: 0.5,
			Packets.FIELD_PROJECTILE_TYPE: "bullet",
		},
		Vector2.ZERO,
		Vector2.ZERO
	)

	var bullet_node = projectile_sync.projectile_nodes["bullet-1"]
	projectile_sync.remove_projectile("bullet-1")
	assert_eq(projectile_sync.pool_size_for_type("bullet"), 1)

	projectile_sync.apply_projectile(
		"torpedo-1",
		{
			Packets.FIELD_X: 30.0,
			Packets.FIELD_Y: 40.0,
			Packets.FIELD_ROTATION: 1.0,
			Packets.FIELD_PROJECTILE_TYPE: "torpedo",
		},
		Vector2.ZERO,
		Vector2.ZERO
	)

	var torpedo_node = projectile_sync.projectile_nodes["torpedo-1"]

	assert_ne(torpedo_node, bullet_node)
	assert_eq(projectile_sync.pool_size_for_type("bullet"), 1)
	assert_eq(projectile_sync.pool_size_for_type("torpedo"), 0)
	assert_eq(projectile_sync.projectile_node_types["torpedo-1"], "torpedo")

func test_torpedo_node_returns_to_torpedo_pool() -> void:
	var projectile_sync := _new_projectile_sync()

	projectile_sync.apply_projectile(
		"torpedo-1",
		{
			Packets.FIELD_X: 10.0,
			Packets.FIELD_Y: 20.0,
			Packets.FIELD_ROTATION: 0.5,
			Packets.FIELD_PROJECTILE_TYPE: "torpedo",
		},
		Vector2.ZERO,
		Vector2.ZERO
	)

	var torpedo_node = projectile_sync.projectile_nodes["torpedo-1"]
	projectile_sync.remove_projectile("torpedo-1")

	assert_false(projectile_sync.projectile_nodes.has("torpedo-1"))
	assert_eq(projectile_sync.pool_size_for_type("torpedo"), 1)
	assert_eq(projectile_sync.pool_size_for_type("bullet"), 0)
	assert_false(projectile_sync.projectile_node_types.has("torpedo-1"))

func test_torpedo_reuses_torpedo_pool_node() -> void:
	var projectile_sync := _new_projectile_sync()

	projectile_sync.apply_projectile(
		"torpedo-1",
		{
			Packets.FIELD_X: 10.0,
			Packets.FIELD_Y: 20.0,
			Packets.FIELD_ROTATION: 0.5,
			Packets.FIELD_PROJECTILE_TYPE: "torpedo",
		},
		Vector2.ZERO,
		Vector2.ZERO
	)

	var old_torpedo_node = projectile_sync.projectile_nodes["torpedo-1"]
	projectile_sync.remove_projectile("torpedo-1")

	projectile_sync.apply_projectile(
		"torpedo-2",
		{
			Packets.FIELD_X: 30.0,
			Packets.FIELD_Y: 40.0,
			Packets.FIELD_ROTATION: 1.0,
			Packets.FIELD_PROJECTILE_TYPE: "torpedo",
		},
		Vector2.ZERO,
		Vector2.ZERO
	)

	var new_torpedo_node = projectile_sync.projectile_nodes["torpedo-2"]

	assert_eq(new_torpedo_node, old_torpedo_node)
	assert_eq(projectile_sync.pool_size_for_type("torpedo"), 0)
	assert_eq(projectile_sync.projectile_node_types["torpedo-2"], "torpedo")
	assert_false(projectile_sync.projectile_nodes.has("torpedo-1"))
	assert_true(projectile_sync.projectile_nodes.has("torpedo-2"))

func test_new_torpedo_uses_torpedo_scene_when_pool_empty() -> void:
	var projectile_sync := _new_projectile_sync()

	projectile_sync.apply_projectile(
		"torpedo-1",
		{
			Packets.FIELD_X: 10.0,
			Packets.FIELD_Y: 20.0,
			Packets.FIELD_ROTATION: 0.5,
			Packets.FIELD_PROJECTILE_TYPE: "torpedo",
		},
		Vector2.ZERO,
		Vector2.ZERO
	)

	var node = projectile_sync.projectile_nodes["torpedo-1"]

	assert_eq(projectile_sync.projectile_node_types["torpedo-1"], "torpedo")
	assert_not_null(node)
	assert_eq(node.name, "Torpedo")

func _new_projectile_sync() -> ProjectileSync:
	var projectile_sync := ProjectileSync.new()
	var bullets_layer := Node2D.new()
	add_child_autofree(bullets_layer)
	projectile_sync.configure(bullets_layer)
	return projectile_sync