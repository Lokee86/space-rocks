extends Control


func apply_profile(profile: Dictionary) -> void:
	var callsign := str(profile.get("callsign", "Guest"))
	var activity_status := str(profile.get("activity_status", "OFFLINE")).to_upper()
	var total_score := int(profile.get("total_score", 0))
	var high_score := int(profile.get("high_score", 0))
	var games_played := int(profile.get("games_played", 0))
	var wins := int(profile.get("wins", 0))
	var ship_deaths := int(profile.get("ship_deaths", 0))

	_set_label_text("%CallsignLabel", "ReadoutContainer/VBoxContainer/CallsignActivityContainer/CallsignLabel", "CALLSIGN: " + callsign)
	_set_label_text("%ActivityLabel", "ReadoutContainer/VBoxContainer/CallsignActivityContainer/ActivityLabel", "STATUS: " + activity_status)
	_set_label_text("%TotalScoreValueLabel", "ReadoutContainer/VBoxContainer/ScoreContainer/TotalScoreContainer/VBoxContainer/TotalScoreValueLabel", str(total_score))
	_set_label_text("%HighScoreValueLabel", "ReadoutContainer/VBoxContainer/ScoreContainer/HighScoreContainer/VBoxContainer/HighScoreValueLabel", str(high_score))
	_set_label_text("%MissionsValueLabel", "ReadoutContainer/VBoxContainer/StatContainer/MissionsContainer/VBoxContainer/MissionsValueLabel", str(games_played))
	_set_label_text("%WinsValueLabel", "ReadoutContainer/VBoxContainer/StatContainer/WinsContainer/VBoxContainer/WinsValueLabel", str(wins))
	_set_label_text("%ShipLossesValueLabel", "ReadoutContainer/VBoxContainer/StatContainer/ShipLossesContainer/VBoxContainer/ShipLossesValueLabel", str(ship_deaths))


func _set_label_text(unique_label_name: String, fallback_path: String, text: String) -> void:
	var label := get_node_or_null(unique_label_name) as Label
	if label == null:
		label = get_node_or_null(fallback_path) as Label
	if label != null:
		label.text = text
