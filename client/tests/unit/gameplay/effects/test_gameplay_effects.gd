extends GutTest

const Constants = preload("res://scripts/generated/constants/constants.gd")
const GameplayEffects = preload("res://scripts/gameplay/effects/gameplay_effects.gd")

var owner_node: Node2D
var effects: GameplayEffects


func before_each() -> void:
	owner_node = Node2D.new()
	add_child(owner_node)

	effects = GameplayEffects.new()
	effects.configure(owner_node, null)


func after_each() -> void:
	effects = null
	if owner_node != null:
		owner_node.free()
		owner_node = null


func test_spawn_torpedo_explosion_adds_scaled_effect_node() -> void:
	effects.spawn_torpedo_explosion(Vector2(10, 20))

	assert_eq(owner_node.get_child_count(), 1)

	var explosion_node := owner_node.get_child(0) as Node2D
	assert_not_null(explosion_node)
	assert_eq(explosion_node.name, "TorpedoExplosion")
	assert_eq(explosion_node.global_position, Vector2(10, 20))
	assert_eq(explosion_node.z_index, Constants.EFFECT_Z_INDEX)

	var sprite := explosion_node.get_node_or_null("AnimatedSprite2D") as AnimatedSprite2D
	assert_not_null(sprite)
	var texture: Texture2D = sprite.sprite_frames.get_frame_texture("torpedo_explosion", 5)
	assert_not_null(texture)
	var target_diameter: float = float(Constants.TORPEDO_RADIAL_ZONE_COUNT * Constants.TORPEDO_RADIAL_ZONE_WIDTH) * 2.0
	var source_diameter: float = float(max(texture.get_width(), texture.get_height()))
	var expected_scale: Vector2 = Vector2.ONE * (target_diameter / source_diameter)
	assert_eq(sprite.scale, expected_scale)

	var sound := explosion_node.get_node_or_null("TorpedoExplosionSound") as AudioStreamPlayer2D
	assert_not_null(sound)
