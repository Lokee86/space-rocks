extends GutTest

const PlayerHuePresenter := preload("res://scripts/gameplay/presentation/player_hue_presenter.gd")


func test_remote_hue_for_player_is_deterministic() -> void:
	var sync := PlayerHuePresenter.new()

	assert_eq(sync.remote_hue_for_player("remote-player"), sync.remote_hue_for_player("remote-player"))


func test_remote_hue_avoids_local_default_hue() -> void:
	var sync := PlayerHuePresenter.new()
	var hue: float = sync.remote_hue_for_player("remote-player")

	assert_false(sync.hues_similar(hue, sync.local_player_hue))


func test_get_remote_player_hues_filters_current_self_id() -> void:
	var sync := PlayerHuePresenter.new()
	sync.remote_player_hues = {
		"local-player": 0.1,
		"remote-player": 0.58,
	}

	var hues := sync.remote_player_hues_without("local-player")

	assert_false(hues.has("local-player"))
	assert_eq(hues["remote-player"], 0.58)
