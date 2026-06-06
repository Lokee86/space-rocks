extends GutTest

const GameplayStatePacketReader := preload("res://scripts/gameplay/state/gameplay_state_packet_reader.gd")
const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
const WorldStateFixture := preload("res://tests/fixtures/world_state_fixture.gd")


func test_read_extracts_state_packet_facts() -> void:
	var state := WorldStateFixture.state()
	var local_player: Dictionary = state[Packets.FIELD_PLAYERS][WorldStateFixture.LOCAL_PLAYER_ID]
	local_player[Packets.FIELD_HEALTH] = 75
	local_player[Packets.FIELD_SHIELDS] = 20
	state[Packets.FIELD_PLAYER_LIFECYCLE] = {
		WorldStateFixture.LOCAL_PLAYER_ID: "active",
		WorldStateFixture.REMOTE_PLAYER_ID: "pending_respawn",
	}
	state[Packets.FIELD_EVENTS] = [{"type": "example"}]

	var facts := GameplayStatePacketReader.read(state)

	assert_eq(facts["self_id"], WorldStateFixture.LOCAL_PLAYER_ID)
	assert_eq(facts["server_players"], state[Packets.FIELD_PLAYERS])
	assert_eq(facts["server_bullets"], state[Packets.FIELD_BULLETS])
	assert_eq(facts["server_asteroids"], state[Packets.FIELD_ASTEROIDS])
	assert_eq(facts["server_events"], state[Packets.FIELD_EVENTS])
	assert_true(facts["has_lives"])
	assert_eq(facts["lives"], 3)
	assert_eq(facts["player_lifecycle"][WorldStateFixture.LOCAL_PLAYER_ID], "active")
	assert_eq(facts["player_lifecycle"][WorldStateFixture.REMOTE_PLAYER_ID], "pending_respawn")
	assert_eq(
		facts["server_players"][WorldStateFixture.LOCAL_PLAYER_ID][Packets.FIELD_HEALTH],
		75
	)
	assert_eq(
		facts["server_players"][WorldStateFixture.LOCAL_PLAYER_ID][Packets.FIELD_SHIELDS],
		20
	)


func test_read_uses_existing_defaults_for_optional_fields() -> void:
	var state := WorldStateFixture.state()
	state.erase(Packets.FIELD_BULLETS)
	state.erase(Packets.FIELD_ASTEROIDS)
	state.erase(Packets.FIELD_EVENTS)
	state.erase(Packets.FIELD_LIVES)
	state.erase(Packets.FIELD_PLAYER_LIFECYCLE)

	var facts := GameplayStatePacketReader.read(state)

	assert_eq(facts["server_bullets"], {})
	assert_eq(facts["server_asteroids"], {})
	assert_eq(facts["server_events"], [])
	assert_false(facts["has_lives"])
	assert_eq(facts["lives"], 0)
	assert_eq(facts["player_lifecycle"], {})
	assert_false(facts.has("debug_status"))
	assert_false(facts.has("debug_statuses"))
	assert_false(facts.has("debug_collision_bodies"))


func test_read_ignores_non_array_events() -> void:
	var state := WorldStateFixture.state()
	state[Packets.FIELD_EVENTS] = "not-events"

	var facts := GameplayStatePacketReader.read(state)

	assert_eq(facts["server_events"], [])


func test_read_excludes_debug_fields_from_gameplay_state() -> void:
	var state := WorldStateFixture.state()
	state["debug_status"] = {"Player-1": {"invincible": true}}
	state["debug_statuses"] = {"Player-1": {"invincible": true}}
	state["debug_collision_bodies"] = [{"kind": "player"}]

	var facts := GameplayStatePacketReader.read(state)

	assert_false(facts.has("debug_status"))
	assert_false(facts.has("debug_statuses"))
	assert_false(facts.has("debug_collision_bodies"))


func test_read_preserves_server_sent_msec() -> void:
	var state := WorldStateFixture.state()
	state["server_sent_msec"] = 123456

	var facts := GameplayStatePacketReader.read(state)

	assert_eq(facts["server_sent_msec"], 123456)

	state.erase("server_sent_msec")
	facts = GameplayStatePacketReader.read(state)

	assert_eq(facts["server_sent_msec"], -1)
