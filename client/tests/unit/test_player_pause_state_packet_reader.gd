extends GutTest

const PlayerPauseStatePacketReader := preload("res://scripts/gameplay/state/player_pause_state_packet_reader.gd")
const Packets := preload("res://scripts/generated/networking/packets/packets.gd")


func test_is_player_pause_state_returns_true_for_player_pause_state() -> void:
	var packet := {
		Packets.FIELD_TYPE: Packets.TYPE_PLAYER_PAUSE_STATE,
	}

	assert_true(PlayerPauseStatePacketReader.is_player_pause_state(packet))


func test_is_player_pause_state_returns_false_for_world_full_packet() -> void:
	var packet := {
		Packets.FIELD_TYPE: Packets.TYPE_WORLD_FULL,
	}

	assert_false(PlayerPauseStatePacketReader.is_player_pause_state(packet))


func test_read_extracts_player_id_and_paused_true() -> void:
	var packet := {
		Packets.FIELD_PLAYER_ID: "player-1",
		Packets.FIELD_PAUSED: true,
	}

	var facts := PlayerPauseStatePacketReader.read(packet)

	assert_eq(facts["player_id"], "player-1")
	assert_true(facts["paused"])


func test_read_defaults_missing_fields() -> void:
	var facts := PlayerPauseStatePacketReader.read({})

	assert_eq(facts["player_id"], "")
	assert_false(facts["paused"])

