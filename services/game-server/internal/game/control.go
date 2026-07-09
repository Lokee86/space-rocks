package game

type Control struct {
	game *Game
}

func NewControl(game *Game) *Control {
	return &Control{game: game}
}

func (target *Control) ObserverKey() any {
	return target.game
}
