extends Control

const PregameModePresenter := preload("res://scripts/ui/menus/pregame_mode_presenter.gd")
const PregameMenuMode := preload("res://scripts/ui/menu_flow/pregame_menu_mode.gd")

signal back_requested
signal play_endless_requested
signal create_game_requested
signal join_game_requested
signal logout_requested
signal profile_requested
signal select_pilot_requested

var mode_presenter: PregameModePresenter
var current_mode: String = ""


func _ready() -> void:
	mode_presenter = PregameModePresenter.new()
	var back_button := get_node_or_null("%BackButton") as BaseButton
	if back_button != null:
		back_button.pressed.connect(_on_back_pressed)
	var endless_create_button := get_node_or_null("%EndlessCreateButton") as BaseButton
	if endless_create_button != null:
		endless_create_button.pressed.connect(_on_endless_create_pressed)
	var campaign_join_button := get_node_or_null("%CampaignJoinButton") as BaseButton
	if campaign_join_button != null:
		campaign_join_button.pressed.connect(_on_campaign_join_pressed)
	var profile_button := get_node_or_null("%ProfileButton") as BaseButton
	if profile_button != null:
		profile_button.pressed.connect(_on_profile_pressed)
	var select_pilot_logout_button := get_node_or_null("%SelectPilotLogoutButton") as BaseButton
	if select_pilot_logout_button != null:
		select_pilot_logout_button.pressed.connect(_on_select_pilot_logout_pressed)
	set_callsign("Guest")
	show_single_player_mode()


func show_single_player_mode() -> void:
	current_mode = PregameMenuMode.SINGLE_PLAYER
	mode_presenter.apply_mode(self, current_mode)
	set_callsign("Guest")


func show_multiplayer_mode() -> void:
	current_mode = PregameMenuMode.MULTIPLAYER
	mode_presenter.apply_mode(self, current_mode)
	set_callsign("Guest")


func set_callsign(callsign: String) -> void:
	var callsign_label := get_node_or_null("%CallsignLabel") as Label
	if callsign_label != null:
		callsign_label.text = "CALLSIGN:\n" + callsign


func _on_back_pressed() -> void:
	back_requested.emit()


func _on_endless_create_pressed() -> void:
	if current_mode == PregameMenuMode.SINGLE_PLAYER:
		play_endless_requested.emit()
	elif current_mode == PregameMenuMode.MULTIPLAYER:
		create_game_requested.emit()


func _on_campaign_join_pressed() -> void:
	if current_mode == PregameMenuMode.MULTIPLAYER:
		join_game_requested.emit()


func _on_profile_pressed() -> void:
	profile_requested.emit()


func _on_select_pilot_logout_pressed() -> void:
	if current_mode == PregameMenuMode.MULTIPLAYER:
		logout_requested.emit()
	elif current_mode == PregameMenuMode.SINGLE_PLAYER:
		select_pilot_requested.emit()
