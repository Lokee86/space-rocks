extends GutTest

const HudScene := preload("res://scenes/ui/hud.tscn")
const GameplayHudFlow := preload("res://scripts/shell/gameplay_hud_flow.gd")


func test_lives_icon_uses_presented_player_hue() -> void:
	var hud := HudScene.instantiate()
	add_child_autofree(hud)
	await get_tree().process_frame
	var flow := GameplayHudFlow.new()
	flow.configure(hud)

	flow.apply_player_hue(0.58)

	var icon := hud.get_node("%LivesIcon") as TextureRect
	var material := icon.material as ShaderMaterial
	assert_not_null(material)
	assert_almost_eq(float(material.get_shader_parameter("hue_shift")), 0.58, 0.0001)
