extends GutTest

const PacketCodec := preload("res://scripts/networking/packets/packet_codec.gd")


func test_encode_returns_json_text_that_decodes_to_original_fields() -> void:
	var packet := {
		"type": "example_packet",
		"room_code": "ABC123",
		"ready": true,
	}

	var encoded := PacketCodec.encode(packet)
	var decoded = JSON.parse_string(encoded)

	assert_eq(typeof(encoded), TYPE_STRING)
	assert_eq(typeof(decoded), TYPE_DICTIONARY)
	assert_eq(decoded["type"], packet["type"])
	assert_eq(decoded["room_code"], packet["room_code"])
	assert_eq(decoded["ready"], packet["ready"])


func test_decode_returns_dictionary_for_valid_object_json() -> void:
	var decoded = PacketCodec.decode("{\"type\":\"example_packet\",\"x\":12.5}")

	assert_eq(typeof(decoded), TYPE_DICTIONARY)
	assert_eq(decoded["type"], "example_packet")
	assert_eq(decoded["x"], 12.5)


func test_decode_returns_null_for_invalid_json() -> void:
	assert_null(PacketCodec.decode("{invalid json"))
	assert_engine_error("Parse JSON failed")
