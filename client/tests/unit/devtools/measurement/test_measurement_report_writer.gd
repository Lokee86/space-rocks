extends GutTest

const ClientMeasurementReportWriter := preload("res://scripts/devtools/measurement/measurement_report_writer.gd")

var test_directory := ""


func before_each() -> void:
	test_directory = "user://measurement-report-writer-tests-%d" % Time.get_ticks_usec()


func after_each() -> void:
	_remove_tree(test_directory)


func test_writes_versioned_json_to_a_sanitized_filename() -> void:
	var writer := ClientMeasurementReportWriter.new(test_directory, func() -> String:
		return "2026-07-18T12:34:56+00:00"
	)
	var report := {"status": "completed", "frame_timing": {"count": 4}}

	var result := writer.write(report, "run/with spaces")

	assert_true(result["success"])
	assert_eq(result["error"], "")
	assert_true(result["path"].begins_with(test_directory))
	assert_true(result["path"].ends_with("measurement-v1-2026-07-18T12_34_56_00_00-run_with_spaces.json"))
	assert_true(DirAccess.dir_exists_absolute(test_directory))
	assert_false(FileAccess.file_exists("%s.tmp" % result["path"]))

	var saved_report = JSON.parse_string(FileAccess.get_file_as_string(result["path"]))
	assert_eq(saved_report["status"], "completed")
	assert_eq(saved_report["frame_timing"]["count"], 4.0)


func test_failed_finalize_returns_path_and_removes_temporary_file() -> void:
	var writer := ClientMeasurementReportWriter.new(test_directory, func() -> String:
		return "fixed-timestamp"
	)
	var final_path := writer.build_path("run/id")
	assert_true(DirAccess.make_dir_recursive_absolute(final_path) == OK)

	var result := writer.write({"status": "completed"}, "run/id")

	assert_false(result["success"])
	assert_eq(result["path"], final_path)
	assert_true(result["error"].contains("failed to finalize measurement report"))
	assert_true(DirAccess.dir_exists_absolute(final_path))
	assert_false(FileAccess.file_exists("%s.tmp" % final_path))


func test_directory_failure_does_not_create_report() -> void:
	var blocker_path := test_directory.path_join("not-a-directory")
	assert_true(DirAccess.make_dir_recursive_absolute(test_directory) == OK)
	var blocker := FileAccess.open(blocker_path, FileAccess.WRITE)
	assert_not_null(blocker)
	blocker.close()

	var writer := ClientMeasurementReportWriter.new(blocker_path, func() -> String:
		return "fixed-timestamp"
	)
	var result := writer.write({"status": "completed"}, "run")

	assert_false(result["success"])
	assert_eq(result["path"], writer.build_path("run"))
	assert_true(result["error"].contains("failed to create measurement directory"))
	assert_false(FileAccess.file_exists(result["path"]))


func _remove_tree(path: String) -> void:
	var directory := DirAccess.open(path)
	if directory == null:
		return
	directory.list_dir_begin()
	while true:
		var entry := directory.get_next()
		if entry.is_empty():
			break
		var entry_path := path.path_join(entry)
		if directory.current_is_dir():
			_remove_tree(entry_path)
		else:
			DirAccess.remove_absolute(entry_path)
	directory.list_dir_end()
	DirAccess.remove_absolute(path)
