extends GutTest

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
const PlayerRenderApiScript := preload("res://scripts/world/player_render/player_render_api.gd")
const PlayerScene := preload("res://scenes/player.tscn")

var game_owner: Node2D
var player: Player
var view_anchor: Node2D
var api


func before_each() -> void:
	game_owner = Node2D.new()
	add_child(game_owner)

	player = PlayerScene.instantiate()
	view_anchor = Node2D.new()
	view_anchor.add_child(Camera2D.new())

	game_owner.add_child(player)
	game_owner.add_child(view_anchor)

	api = PlayerRenderApiScript.new()
	api.configure(game_owner, player, view_anchor)


func after_each() -> void:
	api = null
	if game_owner != null:
		game_owner.free()
		game_owner = null


func test_apply_state_uses_self_position_when_no_view_target_is_set() -> void:
	var self_id := "player-1"
	var server_players := {
		self_id: {
			Packets.FIELD_X: 100.0,
			Packets.FIELD_Y: 200.0,
			Packets.FIELD_ROTATION: 0.0,
		}
	}

	api.apply_state(self_id, server_players)

	assert_eq(api.server_position(), Vector2(100.0, 200.0))
	assert_eq(view_anchor.global_position, api.visual_position())


func test_apply_state_uses_view_target_position_when_view_target_is_set() -> void:
	var self_id := "player-1"
	var target_id := "player-2"
	var server_players := {
		self_id: {
			Packets.FIELD_X: 100.0,
			Packets.FIELD_Y: 200.0,
			Packets.FIELD_ROTATION: 0.0,
		},
		target_id: {
			Packets.FIELD_X: 400.0,
			Packets.FIELD_Y: 500.0,
			Packets.FIELD_ROTATION: 0.0,
		}
	}

	api.set_view_target_player(target_id)
	api.apply_state(self_id, server_players)

	assert_eq(api.server_position(), Vector2(400.0, 500.0))


func test_apply_state_falls_back_to_self_position_when_view_target_missing() -> void:
	var self_id := "player-1"
	var server_players := {
		self_id: {
			Packets.FIELD_X: 100.0,
			Packets.FIELD_Y: 200.0,
			Packets.FIELD_ROTATION: 0.0,
		}
	}

	api.set_view_target_player("missing-player")
	api.apply_state(self_id, server_players)

	assert_eq(api.server_position(), Vector2(100.0, 200.0))

