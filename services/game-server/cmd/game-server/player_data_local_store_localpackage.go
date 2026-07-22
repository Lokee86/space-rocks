//go:build localpackage

package main

import (
	"os"
	"path/filepath"

	"github.com/Lokee86/space-rocks/player-data/playerdata"
	"github.com/Lokee86/space-rocks/player-data/playerdata/embeddedsqlite"
)

func playerDataLocalStorePath() string {
	return runtimePath(filepath.Join("player-data", playerdata.DefaultSQLiteFilename))
}

func playerDataLocalStoreFactory() playerdata.LocalStoreFactory {
	return func(path string) (playerdata.Store, error) {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return nil, err
		}
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
