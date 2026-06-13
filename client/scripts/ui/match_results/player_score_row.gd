extends HBoxContainer
class_name PlayerScoreRow


func apply_row(row: Dictionary) -> void:
	var player_id := str(row.get("player_id", row.get("game_player_id", "Player")))
	(%PlayerIDLabel as Label).text = player_id
	(%GameDeathsLabel as Label).text = str(int(row.get("ship_deaths", 0)))
	(%GameScoreLabel as Label).text = str(int(row.get("score", 0)))
