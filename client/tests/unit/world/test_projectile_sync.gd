extends GutTest

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
const ProjectileSync := preload("res://scripts/world/projectile_sync.gd")
const BulletPresentation := preload("res://scripts/entities/bullet.gd")
const TorpedoPresentation := preload("res://scripts/entities/torpedo.gd")
const BulletScene := preload("res://scenes/bullet.tscn")
const TorpedoScene := preload("res://scenes/projectiles/torpedo.tscn")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const Contract := preload("res://scripts/generated/observability/contract_generated.gd")
const EventCapture := preload("res://tests/unit/logging/presentation_event_capture.gd")

func test_invalid_projectile_root_emits_canonical_contract_violation() -> void:
	var capture := _begin_capture()
	var projectile_sync := ProjectileSync.new()
	var invalid_root := Control.new()

	projectile_sync._contract_violation("bullet", invalid_root)
	assert_push_error_count(1)

	var record := capture.last_record()
	assert_eq(record["event"], Contract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION)
	assert_eq(record["fields"]["entity_kind"], "projectile")
	assert_eq(record["fields"]["failure_mode"], "wrong_scene_root")
	assert_eq(record["fields"]["expected_type"], "BulletPresentation")
	assert_eq(record["fields"]["actual_type"], "Control")
	invalid_root.free()


func test_bullet_scene_root_satisfies_bullet_presentation_contract() -> void:
	var node := BulletScene.instantiate()
	add_child_autofree(node)
	assert_true(node is Node2D)
	assert_true(node is BulletPresentation)


func test_torpedo_scene_root_satisfies_torpedo_presentation_contract() -> void:
	var node := TorpedoScene.instantiate()
	add_child_autofree(node)
	assert_true(node is Node2D)
	assert_true(node is TorpedoPresentation)

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


func test_unknown_hot_projectile_update_does_not_create_node() -> void:
	var projectile_sync := _new_projectile_sync()

	projectile_sync.apply_projectile(
		"bullet-unknown",
		{
			Packets.FIELD_X: 320.0,
			Packets.FIELD_Y: 340.0,
			Packets.FIELD_ROTATION: 0.75,
			Packets.FIELD_PROJECTILE_TYPE: "torpedo",
		},
		Vector2.ZERO,
		Vector2.ZERO,
		false
	)

	assert_false(projectile_sync.projectile_nodes.has("bullet-unknown"))
	assert_false(projectile_sync.projectile_node_types.has("bullet-unknown"))
	assert_false(projectile_sync.initialized_projectiles.has("bullet-unknown"))
	assert_eq(projectile_sync.created_projectile_node_count, 0)
	assert_eq(projectile_sync.pool_size(), 0)


func test_hot_projectile_update_after_lifecycle_delete_is_ignored() -> void:
	var projectile_sync := _new_projectile_sync()
	var projectile_id := "torpedo-1"

	projectile_sync.apply_projectile(
		projectile_id,
		{
			Packets.FIELD_X: 10.0,
			Packets.FIELD_Y: 20.0,
			Packets.FIELD_ROTATION: 0.5,
			Packets.FIELD_PROJECTILE_TYPE: "torpedo",
		},
		Vector2.ZERO,
		Vector2.ZERO
	)
	projectile_sync.remove_projectile(projectile_id)
	projectile_sync.apply_projectile(
		projectile_id,
		{
			Packets.FIELD_X: 320.0,
			Packets.FIELD_Y: 340.0,
			Packets.FIELD_ROTATION: 0.75,
			Packets.FIELD_PROJECTILE_TYPE: "torpedo",
		},
		Vector2.ZERO,
		Vector2.ZERO,
		false
	)

	assert_false(projectile_sync.projectile_nodes.has(projectile_id))
	assert_false(projectile_sync.projectile_node_types.has(projectile_id))
	assert_true(projectile_sync.is_deleted(projectile_id))
	assert_eq(projectile_sync.pool_size_for_type("torpedo"), 1)
	assert_eq(projectile_sync.pool_size_for_type("bullet"), 0)


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
	old_node.modulate = Color(0.25, 0.5, 0.75, 0.9)
	old_node.rotation = 1.25
	old_node.scale = Vector2(2.0, 3.0)
	projectile_sync.remove_projectile("bullet-1")

	projectile_sync.apply_projectile(
		"bullet-2",
		{
			Packets.FIELD_X: 30.0,
			Packets.FIELD_Y: 40.0,
			Packets.FIELD_ROTATION: 0.0,
			Packets.FIELD_PROJECTILE_TYPE: "bullet",
		},
		Vector2.ZERO,
		Vector2.ZERO
	)

	var new_node = projectile_sync.projectile_nodes["bullet-2"]

	assert_eq(new_node, old_node)
	assert_eq(new_node.modulate, Color.WHITE)
	assert_eq(new_node.rotation, 0.0)
	assert_eq(new_node.scale, Vector2.ONE)
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
			Packets.FIELD_ROTATION: 0.0,
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
			Packets.FIELD_ROTATION: 0.0,
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
			Packets.FIELD_ROTATION: 0.0,
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

func test_deleted_projectile_tombstones_are_bounded_and_recreated() -> void:
	var projectile_sync := _new_projectile_sync()
	for index in range(ProjectileSync.DELETED_PROJECTILE_ID_CAP + 1):
		projectile_sync.remove_projectile("bullet-%d" % index)

	assert_eq(projectile_sync.deleted_projectile_ids.size(), ProjectileSync.DELETED_PROJECTILE_ID_CAP)
	assert_false(projectile_sync.deleted_projectile_ids.has("bullet-0"))
	assert_true(projectile_sync.deleted_projectile_ids.has("bullet-%d" % ProjectileSync.DELETED_PROJECTILE_ID_CAP))

	var retained_id := "bullet-%d" % ProjectileSync.DELETED_PROJECTILE_ID_CAP
	var state := {
		Packets.FIELD_X: 10.0,
		Packets.FIELD_Y: 20.0,
		Packets.FIELD_ROTATION: 0.5,
		Packets.FIELD_PROJECTILE_TYPE: "bullet",
	}
	projectile_sync.apply_projectile(retained_id, state, Vector2.ZERO, Vector2.ZERO)

	assert_false(projectile_sync.deleted_projectile_ids.has(retained_id))
	assert_false(projectile_sync._deleted_projectile_id_order.has(retained_id))
	assert_true(projectile_sync.has_projectile(retained_id))

	projectile_sync.remove_projectile(retained_id)
	assert_true(projectile_sync.deleted_projectile_ids.has(retained_id))
	assert_true(projectile_sync._deleted_projectile_id_order.has(retained_id))
	assert_eq(projectile_sync.deleted_projectile_ids.size(), ProjectileSync.DELETED_PROJECTILE_ID_CAP)

func test_reset_clears_active_pooled_and_interpolation_state() -> void:
	var projectile_sync := _new_projectile_sync()
	var state := {
		Packets.FIELD_X: 10.0,
		Packets.FIELD_Y: 20.0,
		Packets.FIELD_ROTATION: 0.5,
		Packets.FIELD_PROJECTILE_TYPE: "bullet",
	}
	projectile_sync.apply_projectile("bullet-1", state, Vector2.ZERO, Vector2.ZERO)
	var active_node = projectile_sync.projectile_nodes["bullet-1"]
	projectile_sync.remove_projectile("bullet-1")
	var pooled_node = projectile_sync.pooled_projectile_nodes_by_type["bullet"][0]
	var torpedo_state := state.duplicate()
	torpedo_state[Packets.FIELD_PROJECTILE_TYPE] = "torpedo"
	projectile_sync.apply_projectile("torpedo-1", torpedo_state, Vector2.ZERO, Vector2.ZERO)
	var second_active_node = projectile_sync.projectile_nodes["torpedo-1"]
	projectile_sync.deleted_projectile_ids["stale-id"] = true

	projectile_sync.reset()

	assert_true(active_node.is_queued_for_deletion())
	assert_true(pooled_node.is_queued_for_deletion())
	assert_true(second_active_node.is_queued_for_deletion())
	assert_true(projectile_sync.projectile_nodes.is_empty())
	assert_true(projectile_sync.projectile_node_types.is_empty())
	assert_true(projectile_sync.initialized_projectiles.is_empty())
	assert_true(projectile_sync.pooled_projectile_nodes_by_type.is_empty())
	assert_true(projectile_sync.target_projectile_positions.is_empty())
	assert_true(projectile_sync.target_projectile_rotations.is_empty())
	assert_true(projectile_sync.deleted_projectile_ids.is_empty())
	assert_true(projectile_sync._deleted_projectile_id_order.is_empty())

func test_reset_allows_reusing_a_projectile_id_in_a_new_match() -> void:
	var projectile_sync := _new_projectile_sync()
	var state := {
		Packets.FIELD_X: 10.0,
		Packets.FIELD_Y: 20.0,
		Packets.FIELD_ROTATION: 0.5,
		Packets.FIELD_PROJECTILE_TYPE: "bullet",
	}
	projectile_sync.apply_projectile("bullet-1", state, Vector2.ZERO, Vector2.ZERO)
	projectile_sync.remove_projectile("bullet-1")
	projectile_sync.reset()
	projectile_sync.apply_projectile("bullet-1", state, Vector2.ZERO, Vector2.ZERO)

	assert_true(projectile_sync.has_projectile("bullet-1"))
	assert_false(projectile_sync.is_deleted("bullet-1"))

func _begin_capture() -> EventCapture:
	var capture := EventCapture.new()
	ClientLogger._set_file_writer_for_tests(capture)
	return capture


func after_each() -> void:
	ClientLogger.reset_for_tests()


func _new_projectile_sync() -> ProjectileSync:
	var projectile_sync := ProjectileSync.new()
	var bullets_layer := Node2D.new()
	add_child_autofree(bullets_layer)
	projectile_sync.configure(bullets_layer)
	return projectile_sync
