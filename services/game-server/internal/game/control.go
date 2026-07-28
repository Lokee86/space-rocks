package game

type Control struct {
	game *Game
}

func NewControl(game *Game) *Control {
	return &Control{game: game}
}
