extends GutTest

const DebugShapeCatalogPacketReader := preload("res://scripts/devtools/hitboxes/debug_shape_catalog_packet_reader.gd")


func test_read_returns_shapes_dictionary_for_valid_packet() -> void:
	var packet := {
		"shapes": {
			"bullet": {"id": "bullet"}
		}
	}

	var result := DebugShapeCatalogPacketReader.read(packet)

	assert_true(result.has("shapes"))
	assert_eq(result["shapes"], packet["shapes"])


func test_read_returns_empty_shapes_for_malformed_shapes() -> void:
	var packet := {
		"shapes": "not-a-dictionary"
	}

	var result := DebugShapeCatalogPacketReader.read(packet)

	assert_true(result.has("shapes"))
	assert_eq(result["shapes"], {})
