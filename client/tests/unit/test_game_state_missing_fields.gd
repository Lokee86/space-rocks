extends GutTest

const EffectsScript := preload("res://scripts/effects.gd")
const GameScript := preload("res://scripts/game.gd")
const HudControllerScript := preload("res://scripts/ui/hud_controller.gd")
const HudScene := preload("res://scenes/ui/hud.tscn")
const Packets := preload("res://scripts/networking/packets.gd")
const PlayerScene := preload("res://scenes/player.tscn")
const WorldStateFixture := preload("res://tests/fixtures/world_state_fixture.gd")
const WorldSyncScript := preload("res://scripts/networking/world_sync.gd")

var game
var hud_scene: Control
var hud_controller: HudController


func before_each() -> void:
	game = GameScript.new()
	game.player = PlayerScene.instantiate()
	game.bullets = Node2D.new()
	game.asteroids = Node2D.new()
	game.add_child(game.player)
	game.add_child(game.bullets)
	game.add_child(game.asteroids)

	hud_scene = HudScene.instantiate()
	hud_controller = HudControllerScript.new()
	hud_controller.configure(hud_scene)
	game.hud_controller = hud_controller

	game.world_sync = WorldSyncScript.new()
	game.world_sync.configure(game, game.player, game.bullets, game.asteroids)
	game.effects = EffectsScript.new()
	game.effects.configure(game, null)


func after_each() -> void:
	hud_controller = null
	if hud_scene != null:
		hud_scene.free()
		hud_scene = null
	if game != null:
		game.free()
		game = null


func test_apply_state_missing_lives_keeps_current_hud_lives_and_does_not_crash() -> void:
	hud_controller.set_lives(7)
	var state := WorldStateFixture.state()
	state.erase(Packets.FIELD_LIVES)
	state[Packets.FIELD_ASTEROIDS] = {}
	state[Packets.FIELD_BULLETS] = {}

	game._apply_state(state)

	assert_eq(_lives_label().text, "7 x ")
	assert_true(game.has_received_state)


func test_self_death_event_missing_respawn_delay_uses_safe_zero_delay() -> void:
	game.self_id = WorldStateFixture.LOCAL_PLAYER_ID
	hud_controller.set_lives(2)
	var event := {
		Packets.FIELD_TYPE: Packets.TYPE_SHIP_DEATH,
		Packets.FIELD_PLAYER_ID: WorldStateFixture.LOCAL_PLAYER_ID,
		Packets.FIELD_X: 100.0,
		Packets.FIELD_Y: 120.0,
		Packets.FIELD_LIVES: 1,
	}

	game._apply_self_death_event(event)

	assert_eq(_lives_label().text, "1 x ")
	assert_true(hud_controller.is_dead)
	assert_eq(hud_controller.respawn_countdown_remaining, 0.0)


func _lives_label() -> Label:
	return hud_scene.find_child("LivesCount", true, false) as Label
