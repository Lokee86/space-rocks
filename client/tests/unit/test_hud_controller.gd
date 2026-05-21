extends GutTest

const HudControllerScript := preload("res://scripts/ui/hud_controller.gd")
const HudScene := preload("res://scenes/ui/hud.tscn")

var hud_scene: Control
var hud_controller: HudController


func before_each() -> void:
	hud_scene = HudScene.instantiate()
	add_child(hud_scene)
	hud_controller = HudControllerScript.new()
	hud_controller.configure(hud_scene)


func after_each() -> void:
	hud_controller = null
	if hud_scene != null:
		hud_scene.free()
		hud_scene = null


func test_set_lives_shows_three_lives() -> void:
	hud_controller.set_lives(3)

	assert_eq(_lives_label().text, "3 x ")


func test_set_lives_shows_two_lives() -> void:
	hud_controller.set_lives(2)

	assert_eq(_lives_label().text, "2 x ")


func test_set_lives_shows_zero_lives() -> void:
	hud_controller.set_lives(0)

	assert_eq(_lives_label().text, "0 x ")


func test_set_dead_shows_death_overlay_and_countdown() -> void:
	hud_controller.set_dead(3.2)

	assert_true(hud_controller.death_overlay.visible)
	assert_false(hud_controller.game_over_overlay.visible)
	assert_eq(hud_controller.respawn_timer_label.text, "Respawn in 4")
	assert_true(hud_controller.respawn_timer_label.visible)
	assert_false(hud_controller.respawn_tell_label.visible)


func test_set_alive_hides_death_overlay() -> void:
	hud_controller.set_dead(3.0)
	hud_controller.set_alive()

	assert_false(hud_controller.death_overlay.visible)
	assert_false(hud_controller.game_over_overlay.visible)
	assert_false(hud_controller.is_dead)
	assert_false(hud_controller.can_respawn)


func test_update_shows_manual_respawn_prompt_when_countdown_finishes() -> void:
	hud_controller.set_dead(1.0)
	hud_controller.update(1.0)

	assert_true(hud_controller.death_overlay.visible)
	assert_true(hud_controller.can_respawn)
	assert_eq(hud_controller.respawn_timer_label.text, "")
	assert_true(hud_controller.respawn_tell_label.visible)


func _lives_label() -> Label:
	return hud_scene.find_child("LivesCount", true, false) as Label
