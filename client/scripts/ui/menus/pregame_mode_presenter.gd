class_name PregameModePresenter
extends RefCounted

const PregameMenuMode := preload("res://scripts/ui/menu_flow/pregame_menu_mode.gd")


func apply_mode(menu: Control, mode: String) -> void:
	if mode == PregameMenuMode.SINGLE_PLAYER:
		_apply_single_player(menu)
	elif mode == PregameMenuMode.MULTIPLAYER:
		_apply_multiplayer(menu)


func _apply_single_player(menu: Control) -> void:
	_set_label_text(menu, "%ModeLabel", "SINGLE PLAYER")
	_set_label_visible(menu, "%EndlessLabel", true)
	_set_label_visible(menu, "%CreateLabel", false)
	_set_label_visible(menu, "%CampaignLabel", true)
	_set_label_visible(menu, "%JoinLabel", false)
	_set_label_visible(menu, "%SelectPilotLabel", true)
	_set_label_visible(menu, "%LogoutLabel", false)
	_set_button_disabled(menu, "%EndlessCreateButton", false)
	_set_button_disabled(menu, "%CampaignJoinButton", true)
	_set_button_disabled(menu, "%LoadoutButton", true)
	_set_button_disabled(menu, "%ProvisionerButton", true)
	_set_button_disabled(menu, "%BuyOrebitsButton", true)
	_set_button_disabled(menu, "%ProfileButton", false)
	_set_button_disabled(menu, "%SelectPilotLogoutButton", false)


func _apply_multiplayer(menu: Control) -> void:
	_set_label_text(menu, "%ModeLabel", "MULTIPLAYER")
	_set_label_visible(menu, "%EndlessLabel", false)
	_set_label_visible(menu, "%CreateLabel", true)
	_set_label_visible(menu, "%CampaignLabel", false)
	_set_label_visible(menu, "%JoinLabel", true)
	_set_label_visible(menu, "%SelectPilotLabel", false)
	_set_label_visible(menu, "%LogoutLabel", true)
	_set_button_disabled(menu, "%EndlessCreateButton", false)
	_set_button_disabled(menu, "%CampaignJoinButton", false)
	_set_button_disabled(menu, "%LoadoutButton", true)
	_set_button_disabled(menu, "%ProvisionerButton", true)
	_set_button_disabled(menu, "%BuyOrebitsButton", true)
	_set_button_disabled(menu, "%ProfileButton", false)
	_set_button_disabled(menu, "%SelectPilotLogoutButton", false)


func _set_label_text(menu: Control, node_path: String, text: String) -> void:
	var label := menu.get_node_or_null(node_path) as Label
	if label != null:
		label.text = text


func _set_label_visible(menu: Control, node_path: String, visible: bool) -> void:
	var label := menu.get_node_or_null(node_path) as Label
	if label != null:
		label.visible = visible


func _set_button_disabled(menu: Control, node_path: String, disabled: bool) -> void:
	var button := menu.get_node_or_null(node_path) as BaseButton
	if button != null:
		button.disabled = disabled
