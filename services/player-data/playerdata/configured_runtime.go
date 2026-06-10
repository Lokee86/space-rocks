package playerdata

import "errors"

type RuntimeConfig struct {
	RailsBaseURL       string
	RailsInternalToken string
	RailsBearerToken   string
	SQLitePath         string
}

func NewConfiguredRuntime(config RuntimeConfig) (*Runtime, error) {
	var accountStore Store
	if config.RailsBaseURL != "" {
		store, err := NewRailsStore(RailsStoreConfig{
			BaseURL:       config.RailsBaseURL,
			InternalToken: config.RailsInternalToken,
			BearerToken:   config.RailsBearerToken,
		})
		if err != nil {
			return nil, err
		}
		accountStore = store
	} else {
		accountStore = NewMemoryStore()
	}

	var localStore Store
	if config.SQLitePath != "" {
		store, err := NewSQLiteStore(SQLiteStoreConfig{Path: config.SQLitePath})
		if err != nil {
			return nil, err
		}
		if err := store.InitSchema(); err != nil {
			_ = store.Close()
			return nil, err
		}
		localStore = store
	} else {
		localStore = NewMemoryStore()
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
