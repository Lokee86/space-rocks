extends GutTest

const CompactLanePacket := preload("res://scripts/protocol/realtime/compact_lane_packet.gd")
func test_shared_realtime_wire_fixtures_match_current_decoder() -> void:
	var fixture_root := ProjectSettings.globalize_path("res://").path_join("../shared/packets/fixtures/realtime_wire").simplify_path()
	var paths := _fixture_paths(fixture_root)
	assert_gt(paths.size(), 0, "expected shared realtime wire fixtures")
	for path in paths:
		var fixture := _load_fixture(path)
		assert_true(fixture.has("name"), "%s missing name" % path)
		assert_true(fixture.has("readable"), "%s missing readable" % path)
		assert_true(fixture.has("compact"), "%s missing compact" % path)
		assert_true(fixture.has("expanded"), "%s missing expanded" % path)
		var actual: Dictionary = CompactLanePacket.expand_packet(fixture["compact"])
		_assert_json_equal(actual, fixture["expanded"], str(fixture["name"]))


func _fixture_paths(root: String) -> PackedStringArray:
	var paths := PackedStringArray()
	_collect_fixture_paths(root, paths)
	paths.sort()
	return paths


func _collect_fixture_paths(path: String, paths: PackedStringArray) -> void:
	var directory := DirAccess.open(path)
	assert_not_null(directory, "unable to open fixture directory %s" % path)
	if directory == null:
		return
	directory.list_dir_begin()
	var entry := directory.get_next()
	while entry != "":
		if entry != "." and entry != "..":
			var child := path.path_join(entry)
			if directory.current_is_dir():
				_collect_fixture_paths(child, paths)
			elif entry.ends_with(".json"):
				paths.append(child)
		entry = directory.get_next()
	directory.list_dir_end()


func _load_fixture(path: String) -> Dictionary:
	var file := FileAccess.open(path, FileAccess.READ)
	assert_not_null(file, "unable to read fixture %s" % path)
	if file == null:
		return {}
	var parsed = JSON.parse_string(file.get_as_text())
	assert_true(parsed is Dictionary, "fixture %s must contain a JSON object" % path)
	return parsed if parsed is Dictionary else {}


func _assert_json_equal(actual, expected, context: String) -> void:
	if actual is Dictionary and expected is Dictionary:
		assert_eq(actual.size(), expected.size(), "%s dictionary size" % context)
		for key in expected.keys():
			assert_true(actual.has(key), "%s missing key %s" % [context, key])
			if actual.has(key):
				_assert_json_equal(actual[key], expected[key], "%s.%s" % [context, key])
		return
	if actual is Array and expected is Array:
		assert_eq(actual.size(), expected.size(), "%s array size" % context)
		for index in range(min(actual.size(), expected.size())):
			_assert_json_equal(actual[index], expected[index], "%s[%d]" % [context, index])
		return
	if (actual is int or actual is float) and (expected is int or expected is float):
		assert_eq(float(actual), float(expected), context)
		return
	assert_eq(actual, expected, context)
