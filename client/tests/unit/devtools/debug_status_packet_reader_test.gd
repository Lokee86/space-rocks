extends GutTest

const DebugStatusPacketReader := preload("res://scripts/devtools/debug_status_packet_reader.gd")


func test_read_returns_debug_status_and_debug_statuses_dicts_for_valid_packet() -> void:
	var packet := {
		"debug_status": {
			"invincible": true,
		},
		"debug_statuses": {
			"player-1": {
				"infinite_lives": false,
			},
		},
	}

	var state := DebugStatusPacketReader.read(packet)

	assert_eq(state["debug_status"], packet["debug_status"])
	assert_eq(state["debug_statuses"], packet["debug_statuses"])


func test_read_defaults_malformed_debug_status_to_empty_dict() -> void:
	var state := DebugStatusPacketReader.read({
		"debug_status": "not-a-dictionary",
		"debug_statuses": {},
	})

	assert_eq(state["debug_status"], {})


func test_read_defaults_malformed_debug_statuses_to_empty_dict() -> void:
	var state := DebugStatusPacketReader.read({
		"debug_status": {},
		"debug_statuses": ["not", "a", "dictionary"],
	})

	assert_eq(state["debug_statuses"], {})
