extends GutTest

const PacketCodec := preload("res://scripts/networking/packets/packet_codec.gd")


func test_encode_returns_json_text_that_decodes_to_original_fields() -> void:
	var packet := {
		"type": "example_packet",
		"room_code": "ABC123",
		"ready": true,
	}

	var encoded = PacketCodec.encode(packet)
	var decoded = JSON.parse_string(encoded.wire_message)

	assert_true(encoded.ok)
	assert_eq(typeof(encoded.wire_message), TYPE_STRING)
	assert_false(encoded.wire_message.is_empty())
	assert_eq(typeof(decoded), TYPE_DICTIONARY)
	assert_eq(decoded["type"], packet["type"])
	assert_eq(decoded["room_code"], packet["room_code"])
	assert_eq(decoded["ready"], packet["ready"])


func test_decode_returns_dictionary_for_valid_object_json() -> void:
	var decoded = PacketCodec.decode("{\"type\":\"example_packet\",\"x\":12.5}")

	assert_true(decoded.ok)
	assert_eq(typeof(decoded.packet), TYPE_DICTIONARY)
	assert_eq(decoded.packet["type"], "example_packet")
	assert_eq(decoded.packet["x"], 12.5)


func test_decode_returns_null_for_invalid_json() -> void:
	var decoded = PacketCodec.decode("{invalid json")

	assert_false(decoded.ok)
	assert_true(decoded.error.contains("Invalid packet JSON"))
	assert_eq(decoded.raw, "{invalid json")


func test_decode_rejects_array_root() -> void:
	var decoded = PacketCodec.decode("[1, 2, 3]")

	assert_false(decoded.ok)
	assert_eq(decoded.error, "Packet JSON must decode to a Dictionary")


func test_decode_rejects_missing_type() -> void:
	var decoded = PacketCodec.decode("{\"payload\":{}}")

	assert_false(decoded.ok)
	assert_eq(decoded.error, "Packet envelope is missing required 'type' field")


func test_decode_rejects_empty_type() -> void:
	var decoded = PacketCodec.decode("{\"type\":\"   \"}")

	assert_false(decoded.ok)
	assert_eq(decoded.error, "Packet envelope field 'type' must not be empty")


func test_decode_rejects_non_string_type() -> void:
	var decoded = PacketCodec.decode("{\"type\":12}")

	assert_false(decoded.ok)
	assert_eq(decoded.error, "Packet envelope field 'type' must be a String")


func test_decode_rejects_non_dictionary_payload() -> void:
	var decoded = PacketCodec.decode("{\"type\":\"example_packet\",\"payload\":1}")

	assert_false(decoded.ok)
	assert_eq(decoded.error, "Packet envelope field 'payload' must be a Dictionary when present")


func test_decode_allows_type_without_payload() -> void:
	var decoded = PacketCodec.decode("{\"type\":\"example_packet\"}")

	assert_true(decoded.ok)
	assert_eq(decoded.packet["type"], "example_packet")


func test_decode_accepts_lowercase_lane_packet_fixture() -> void:
	var decoded = PacketCodec.decode(
		"{\"type\":\"world_full\",\"lane\":\"world\",\"sequence\":7,\"baseline_id\":\"baseline-1\",\"snapshot_id\":\"snapshot-1\",\"chunk_index\":0,\"chunk_count\":1,\"is_final_chunk\":true,\"ships\":[{\"id\":\"ship-1\",\"ship_type\":\"v_wing\",\"x\":1,\"y\":2,\"rotation\":0,\"health\":100,\"shields\":0,\"thrusting\":false,\"target_kind\":\"player\",\"target_id\":\"player-1\"}],\"bullets\":[],\"asteroids\":[],\"pickups\":[]}"
	)

	assert_true(decoded.ok)
	assert_eq(decoded.packet["type"], "world_full")
	assert_eq(decoded.packet["ships"][0]["ship_type"], "v_wing")
