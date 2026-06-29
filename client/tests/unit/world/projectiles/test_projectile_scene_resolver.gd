extends GutTest

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
const WorldLaneState := preload("res://scripts/protocol/realtime/world_lane_state.gd")
const ProjectileSceneResolver := preload("res://scripts/world/projectiles/projectile_scene_resolver.gd")
const BulletScene := preload("res://scenes/bullet.tscn")
const TorpedoScene := preload("res://scenes/projectiles/torpedo.tscn")


func test_scene_for_state_defaults_to_bullet_when_missing() -> void:
	assert_eq(ProjectileSceneResolver.scene_for_state({}), BulletScene)


func test_scene_for_state_defaults_to_bullet_when_empty() -> void:
	assert_eq(
		ProjectileSceneResolver.scene_for_state({
			Packets.FIELD_PROJECTILE_TYPE: "",
		}),
		BulletScene
	)


func test_scene_for_state_returns_bullet_scene_for_bullet_type() -> void:
	assert_eq(
		ProjectileSceneResolver.scene_for_state({
			Packets.FIELD_PROJECTILE_TYPE: "bullet",
		}),
		BulletScene
	)


func test_scene_for_state_returns_torpedo_scene_for_torpedo_type() -> void:
	assert_eq(
		ProjectileSceneResolver.scene_for_state({
			Packets.FIELD_PROJECTILE_TYPE: "torpedo",
		}),
		TorpedoScene
	)


func test_scene_for_state_returns_torpedo_scene_for_world_lane_torpedo_bullet_state() -> void:
	var world_lane_state := WorldLaneState.new()
	world_lane_state.upsert_bullet({
		Packets.FIELD_ID: "bullet-1",
		Packets.FIELD_X: 10.0,
		Packets.FIELD_Y: 20.0,
		Packets.FIELD_VELOCITY_X: 0.0,
		Packets.FIELD_VELOCITY_Y: 0.0,
		Packets.FIELD_ROTATION: 0.0,
		Packets.FIELD_OWNER_ID: "player-1",
		Packets.FIELD_LIFESPAN_SECONDS: 1.0,
		Packets.FIELD_WEAPON_ID: "torpedo",
		Packets.FIELD_PROJECTILE_TYPE: "torpedo",
	})

	assert_eq(ProjectileSceneResolver.scene_for_state(world_lane_state.bullets["bullet-1"]), TorpedoScene)


func test_scene_for_state_defaults_to_bullet_for_unknown_type() -> void:
	assert_eq(
		ProjectileSceneResolver.scene_for_state({
			Packets.FIELD_PROJECTILE_TYPE: "mystery",
		}),
		BulletScene
	)
