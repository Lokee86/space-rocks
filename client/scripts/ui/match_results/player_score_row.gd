extends HBoxContainer
class_name PlayerScoreRow

const TeamPresentation := preload("res://scripts/teams/team_presentation.gd")


func apply_row(row: Dictionary) -> void:
	var player_id := str(row.get("player_id", row.get("game_player_id", "Player")))
	var team_id := str(row.get("team_id", ""))
	(%PlayerIDLabel as Label).text = player_id
	(%GameDeathsLabel as Label).text = str(int(row.get("ship_deaths", 0)))
	(%GameScoreLabel as Label).text = str(int(row.get("score", 0)))
	(%TeamSwatch as Control).visible = not team_id.is_empty()
	(%TeamLabel as Control).visible = not team_id.is_empty()
	if not team_id.is_empty():
		(%TeamSwatch as ColorRect).color = TeamPresentation.color(team_id)
		(%TeamLabel as Label).text = TeamPresentation.display_name(team_id)
