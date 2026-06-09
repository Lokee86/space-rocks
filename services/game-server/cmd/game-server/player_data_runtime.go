package main

import "github.com/Lokee86/space-rocks/player-data/playerdata"

func buildPlayerDataRuntime() (*playerdata.Runtime, error) {
	return playerdata.NewInMemoryRuntime()
}
