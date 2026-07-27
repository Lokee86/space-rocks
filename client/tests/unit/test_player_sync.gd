extends GutTest

const PlayerHuePresenter := preload("res://scripts/gameplay/presentation/player_hue_presenter.gd")


func test_remote_hue_for_player_is_deterministic() -> void:
	var sync := PlayerHuePresenter.new()

	assert_eq(sync.remote_hue_for_player("remote-player"), sync.remote_hue_for_player("remote-player"))


func test_remote_hue_avoids_local_default_hue() -> void:
	var sync := PlayerHuePresenter.new()
	var hue: float = sync.remote_hue_for_player("remote-player")

	assert_false(sync.hues_similar(hue, sync.local_player_hue))


func test_team_id_forces_shared_authoritative_hue() -> void:
	var sync := PlayerHuePresenter.new()

	var first := sync.remote_hue_for_player("player-1", "team_3")
	var second := sync.remote_hue_for_player("player-2", "team_3")
	var opposing := sync.remote_hue_for_player("player-3", "team_4")

	assert_eq(first, second)
	assert_ne(first, opposing)


func test_remote_hue_survives_death_and_respawn_without_slot_shift() -> void:
	var sync := PlayerHuePresenter.new()
	sync.set_remote_player_order(["player-2", "player-3"])
	var original_hue := sync.remote_hue_for_player("player-3")
	sync.remote_player_hues["player-3"] = original_hue

	sync.set_remote_player_order(["player-3"])
	sync.remove_player("player-3")

	assert_eq(sync.remote_hue_for_player("player-3"), original_hue)


func test_get_remote_player_hues_filters_current_self_id() -> void:
	var sync := PlayerHuePresenter.new()
	sync.remote_player_hues = {
		"local-player": 0.1,
		"remote-player": 0.58,
	}

	var hues := sync.remote_player_hues_without("local-player")

	assert_false(hues.has("local-player"))
	assert_eq(hues["remote-player"], 0.58)
