extends GutTest

const ProfileReadoutScene := preload("res://scenes/ui/transmission_displays/profile_readout.tscn")


func test_apply_profile_populates_identity_labels() -> void:
	var readout := ProfileReadoutScene.instantiate()
	add_child_autofree(readout)
	await get_tree().process_frame

	readout.apply_profile({
		"callsign": "Ada",
		"activity_status": "ACTIVE",
	})

	assert_eq(readout.get_node("ReadoutContainer/VBoxContainer/CallsignActivityContainer/CallsignLabel").text, "CALLSIGN: Ada")
	assert_eq(readout.get_node("ReadoutContainer/VBoxContainer/CallsignActivityContainer/ActivityLabel").text, "STATUS: ACTIVE")


func test_apply_profile_populates_stat_labels() -> void:
	var readout := ProfileReadoutScene.instantiate()
	add_child_autofree(readout)
	await get_tree().process_frame

	readout.apply_profile({
		"total_score": 100,
		"high_score": 75,
		"games_played": 4,
		"wins": 2,
		"ship_deaths": 3,
	})

	assert_eq(readout.get_node("ReadoutContainer/VBoxContainer/ScoreContainer/TotalScoreContainer/VBoxContainer/TotalScoreValueLabel").text, "100")
	assert_eq(readout.get_node("ReadoutContainer/VBoxContainer/ScoreContainer/HighScoreContainer/VBoxContainer/HighScoreValueLabel").text, "75")
	assert_eq(readout.get_node("ReadoutContainer/VBoxContainer/StatContainer/MissionsContainer/VBoxContainer/MissionsValueLabel").text, "4")
	assert_eq(readout.get_node("ReadoutContainer/VBoxContainer/StatContainer/WinsContainer/VBoxContainer/WinsValueLabel").text, "2")
	assert_eq(readout.get_node("ReadoutContainer/VBoxContainer/StatContainer/ShipLossesContainer/VBoxContainer/ShipLossesValueLabel").text, "3")


func test_apply_profile_defaults_missing_values_to_guest_offline_zero() -> void:
	var readout := ProfileReadoutScene.instantiate()
	add_child_autofree(readout)
	await get_tree().process_frame

	readout.apply_profile({})

	assert_eq(readout.get_node("ReadoutContainer/VBoxContainer/CallsignActivityContainer/CallsignLabel").text, "CALLSIGN: Guest")
	assert_eq(readout.get_node("ReadoutContainer/VBoxContainer/CallsignActivityContainer/ActivityLabel").text, "STATUS: OFFLINE")
	assert_eq(readout.get_node("ReadoutContainer/VBoxContainer/ScoreContainer/TotalScoreContainer/VBoxContainer/TotalScoreValueLabel").text, "0")
	assert_eq(readout.get_node("ReadoutContainer/VBoxContainer/ScoreContainer/HighScoreContainer/VBoxContainer/HighScoreValueLabel").text, "0")
	assert_eq(readout.get_node("ReadoutContainer/VBoxContainer/StatContainer/MissionsContainer/VBoxContainer/MissionsValueLabel").text, "0")
	assert_eq(readout.get_node("ReadoutContainer/VBoxContainer/StatContainer/WinsContainer/VBoxContainer/WinsValueLabel").text, "0")
	assert_eq(readout.get_node("ReadoutContainer/VBoxContainer/StatContainer/ShipLossesContainer/VBoxContainer/ShipLossesValueLabel").text, "0")
