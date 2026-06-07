extends GutTest

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
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


func test_scene_for_state_defaults_to_bullet_for_unknown_type() -> void:
	assert_eq(
		ProjectileSceneResolver.scene_for_state({
			Packets.FIELD_PROJECTILE_TYPE: "mystery",
		}),
		BulletScene
	)
