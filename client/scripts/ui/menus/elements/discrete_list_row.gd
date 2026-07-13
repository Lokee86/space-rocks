class_name DiscreteListRow
extends Control

signal selected(item: Dictionary)


func configure(_display_text: String, _item_data: Dictionary) -> void:
	push_error("DiscreteListRow must implement configure")