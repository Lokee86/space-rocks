extends Node

const public_build: bool = false

const PRIVATE_DEVTOGGLE_ACTIONS := [
	"DevToggle0",
	"DevToggle1",
	"DevToggle2",
	"DevToggle3",
	"DevToggle4",
	"DevToggle5",
	"DevToggle6",
	"DevToggle7",
	"DevToggle10",
]

func _ready() -> void:
	if !public_build:
		return

	for action_name in PRIVATE_DEVTOGGLE_ACTIONS:
		if InputMap.has_action(action_name):
			InputMap.action_erase_events(action_name)
