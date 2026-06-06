extends GutTest

const DebugShapeCatalogStore := preload("res://scripts/devtools/hitboxes/debug_shape_catalog_store.gd")


func test_store_returns_stored_shape_by_id() -> void:
	var store := DebugShapeCatalogStore.new()
	store.apply_catalog_state({
		"shapes": {
			"bullet": {"id": "bullet", "kind": "bullet"}
		}
	})

	assert_eq(store.shape_for_id("bullet"), {"id": "bullet", "kind": "bullet"})


func test_store_returns_empty_dictionary_for_missing_shape_id() -> void:
	var store := DebugShapeCatalogStore.new()
	store.apply_catalog_state({
		"shapes": {
			"bullet": {"id": "bullet"}
		}
	})

	assert_eq(store.shape_for_id("missing"), {})
