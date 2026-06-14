package playerdata

import (
	"errors"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

type Config struct {
	Store Store
}

type Runtime struct {
	store      Store
	dispatcher *Dispatcher
}

func NewRuntime(config Config) (*Runtime, error) {
	if config.Store == nil {
		return nil, errors.New("store is required")
	}

	return &Runtime{
		store:      config.Store,
		dispatcher: NewDispatcher(config.Store),
	}, nil
}

func (r *Runtime) Handle(payload []byte) ([]byte, error) {
	return r.dispatcher.Handle(payload)
}

func (r *Runtime) ListLocalProfiles() ([]LocalProfileSummary, error) {
	if r == nil || r.store == nil {
		return nil, errors.New("local profile management is unavailable")
	}

	localProfileStore, ok := r.store.(LocalProfileStore)
	if !ok {
		return nil, errors.New("local profile management is unavailable")
	}

	return localProfileStore.ListLocalProfiles()
}

func (r *Runtime) CreateLocalProfile(localProfileID string, displayName string, stats protocol.PlayerDataStats) (LocalProfileSummary, error) {
	if r == nil || r.store == nil {
		return LocalProfileSummary{}, errors.New("local profile management is unavailable")
	}

	localProfileStore, ok := r.store.(LocalProfileStore)
	if !ok {
		return LocalProfileSummary{}, errors.New("local profile management is unavailable")
	}

	return localProfileStore.CreateLocalProfile(localProfileID, displayName, stats)
}

func (r *Runtime) GetDefaultLocalProfile() (LocalProfileDefault, error) {
	if r == nil || r.store == nil {
		return LocalProfileDefault{}, errors.New("local profile management is unavailable")
	}

	localProfileStore, ok := r.store.(LocalProfileStore)
	if !ok {
		return LocalProfileDefault{}, errors.New("local profile management is unavailable")
	}

	return localProfileStore.GetDefaultLocalProfile()
}

func (r *Runtime) SetDefaultLocalProfile(identityKind string, localProfileID string) (LocalProfileDefault, error) {
	if r == nil || r.store == nil {
		return LocalProfileDefault{}, errors.New("local profile management is unavailable")
	}

	localProfileStore, ok := r.store.(LocalProfileStore)
	if !ok {
		return LocalProfileDefault{}, errors.New("local profile management is unavailable")
	}

	return localProfileStore.SetDefaultLocalProfile(identityKind, localProfileID)
}
