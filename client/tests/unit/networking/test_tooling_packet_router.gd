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


func test_debug_readouts_are_routed_to_dedicated_signals() -> void:
	var router := ToolingPacketRouter.new()
	var statuses: Array = []
	var catalogs: Array = []
	router.debug_status_received.connect(func(packet: Dictionary) -> void:
		statuses.append(packet)
	)
	router.debug_shape_catalog_received.connect(func(packet: Dictionary) -> void:
		catalogs.append(packet)
	)
	var status_packet := {"type": "debug_status", "request_id": "status-1"}
	var catalog_packet := {"type": "debug_shape_catalog", "request_id": "catalog-1", "shapes": {}}

	router.dispatch(status_packet)
	router.dispatch(catalog_packet)

	assert_eq(statuses, [status_packet])
	assert_eq(catalogs, [catalog_packet])


func test_unknown_tooling_packet_still_uses_unknown_signal() -> void:
	var router := ToolingPacketRouter.new()
	var received: Array = []
	router.unknown_packet_received.connect(func(packet: Dictionary) -> void:
		received.append(packet)
	)

	router.dispatch({"type": "not_a_tooling_packet"})

	assert_eq(received.size(), 1)
