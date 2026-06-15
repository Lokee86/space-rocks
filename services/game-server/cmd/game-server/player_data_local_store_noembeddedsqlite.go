//go:build noembeddedsqlite

package main

import "github.com/Lokee86/space-rocks/player-data/playerdata"

func playerDataLocalStorePath() string {
	return ""
}

func playerDataLocalStoreFactory() playerdata.LocalStoreFactory {
	return nil
}
