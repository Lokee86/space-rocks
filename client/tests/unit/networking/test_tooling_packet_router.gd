extends GutTest

const ToolingPacketRouter := preload("res://scripts/networking/inbound/tooling_packet_router.gd")


func test_tooling_command_result_is_routed_to_terminal_result_signal() -> void:
	var router := ToolingPacketRouter.new()
	var received: Array = []
	router.tooling_command_result_received.connect(func(packet: Dictionary) -> void:
		received.append(packet)
	)
	var packet := {
		"type": "tooling_command_result",
		"request_id": "request-1",
		"command_type": "debug_clear_bullets",
		"applied": true,
	}

	router.dispatch(packet)

	assert_eq(received.size(), 1)
	assert_eq(received[0], packet)


func test_unknown_tooling_packet_still_uses_unknown_signal() -> void:
	var router := ToolingPacketRouter.new()
	var received: Array = []
	router.unknown_packet_received.connect(func(packet: Dictionary) -> void:
		received.append(packet)
	)

	router.dispatch({"type": "not_a_tooling_packet"})

	assert_eq(received.size(), 1)
