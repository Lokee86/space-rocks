extends Node2D

const LOCAL_OFFSET := Vector2(60, -70)

@onready var content: Control = %HBoxContainer
@onready var player_info_label: Label = %PlayerInfoLabel
@onready var player_network_label: Label = %PlayerNetworkLabel
@onready var player_info_panel: Control = player_info_label.get_parent()
@onready var player_network_panel: Control = player_network_label.get_parent()


func _ready() -> void:
	configure_as_player_child()


func _process(_delta: float) -> void:
	if visible:
		global_rotation = 0.0


func configure_as_player_child() -> void:
	top_level = false
	visible = false
	position = LOCAL_OFFSET
	rotation = 0.0

	if content != null:
		content.position = Vector2.ZERO


func show_basic(text: String) -> void:
	show()
	player_info_panel.show()
	player_network_panel.hide()
	player_info_label.text = text


func show_network(text: String) -> void:
	show()
	player_info_panel.hide()
	player_network_panel.show()
	player_network_label.text = text


func hide_label() -> void:
	hide()