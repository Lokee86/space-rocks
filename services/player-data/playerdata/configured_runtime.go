package playerdata

import (
	"errors"
	"path/filepath"
	"runtime"
)

const DefaultSQLiteFilename = "player-data.sqlite3"

type RuntimeConfig struct {
	RailsBaseURL       string
	RailsInternalToken string
	SQLitePath         string
	LocalStoreFactory  LocalStoreFactory
}

type LocalStoreFactory func(path string) (Store, error)

func NewRuntimeFromEnv(getenv func(string) string) (*Runtime, error) {
	return NewConfiguredRuntime(RuntimeConfig{
		RailsBaseURL:       getenv("PLAYER_DATA_RAILS_BASE_URL"),
		RailsInternalToken: getenv("PLAYER_DATA_RAILS_INTERNAL_TOKEN"),
	})
}

func DefaultSQLitePath() string {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join("services", "player-data", "data", DefaultSQLiteFilename)
	}

	playerDataDir := filepath.Dir(filepath.Dir(currentFile))
	return filepath.Join(playerDataDir, "data", DefaultSQLiteFilename)
}

func NewConfiguredRuntime(config RuntimeConfig) (*Runtime, error) {
	var accountStore Store
	if config.RailsBaseURL != "" {
		store, err := NewRailsStore(RailsStoreConfig{
			BaseURL:       config.RailsBaseURL,
			InternalToken: config.RailsInternalToken,
		})
		if err != nil {
			return nil, err
		}
		accountStore = store
	} else {
		accountStore = NewMemoryStore()
	}

	var localStore Store
	if config.SQLitePath == "" {
		localStore = NewNoopStore()
	} else {
		if config.LocalStoreFactory == nil {
			return nil, errors.New("local store factory is required when SQLitePath is set")
		}

		store, err := config.LocalStoreFactory(config.SQLitePath)
		if err != nil {
			return nil, err
		}
		localStore = store
	}

	guestStore := NewGuestMemoryStore()

	runtime, err := NewRuntime(Config{
		Store: NewStoreRouter(accountStore, localStore, guestStore),
	})
	if err != nil {
		if closer, ok := localStore.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
		return nil, errors.New(err.Error())
	}

	return runtime, nil
}
