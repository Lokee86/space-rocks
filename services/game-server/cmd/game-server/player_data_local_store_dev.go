//go:build !noembeddedsqlite

package main

import (
	"github.com/Lokee86/space-rocks/player-data/playerdata"
	"github.com/Lokee86/space-rocks/player-data/playerdata/embeddedsqlite"
)

func playerDataLocalStorePath() string {
	return playerdata.DefaultSQLitePath()
}

func playerDataLocalStoreFactory() playerdata.LocalStoreFactory {
	return func(path string) (playerdata.Store, error) {
		store, err := embeddedsqlite.New(embeddedsqlite.Config{Path: path})
		if err != nil {
			return nil, err
		}
		if err := store.InitSchema(); err != nil {
			_ = store.Close()
			return nil, err
		}
		return store, nil
	}
}
