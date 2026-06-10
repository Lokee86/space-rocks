package main

import (
	"os"

	"github.com/Lokee86/space-rocks/player-data/playerdata"
)

func buildPlayerDataRuntime() (*playerdata.Runtime, error) {
	return playerdata.NewConfiguredRuntime(playerdata.RuntimeConfig{
		RailsBaseURL:       os.Getenv("PLAYER_DATA_RAILS_BASE_URL"),
		RailsInternalToken: os.Getenv("PLAYER_DATA_RAILS_INTERNAL_TOKEN"),
		RailsBearerToken:   os.Getenv("PLAYER_DATA_RAILS_BEARER_TOKEN"),
		SQLitePath:         os.Getenv("PLAYER_DATA_SQLITE_PATH"),
	})
}
