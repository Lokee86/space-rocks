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

func (r *Runtime) LoadStats(identity protocol.PlayerDataIdentity) (protocol.PlayerDataStats, bool, error) {
	if r == nil || r.store == nil {
		return protocol.PlayerDataStats{}, false, errors.New("player-data runtime is required")
	}

	return r.store.LoadStats(identity)
}

func (r *Runtime) LocalProfileSeedStats(seedFromGuestStats bool) (protocol.PlayerDataStats, error) {
	if !seedFromGuestStats {
		return protocol.PlayerDataStats{}, nil
	}

	stats, found, err := r.LoadStats(protocol.PlayerDataIdentity{IdentityKind: IdentityKindGuest})
	if err != nil {
		return protocol.PlayerDataStats{}, err
	}
	if !found {
		return protocol.PlayerDataStats{}, errors.New("guest stats unavailable")
	}

	return stats, nil
}

func (r *Runtime) ListLocalProfiles() ([]LocalProfileSummary, error) {
	if r == nil || r.store == nil {
		return nil, ErrLocalProfileUnavailable
	}

	localProfileStore, ok := r.store.(LocalProfileStore)
	if !ok {
		return nil, ErrLocalProfileUnavailable
	}

	return localProfileStore.ListLocalProfiles()
}

func (r *Runtime) CreateLocalProfile(localProfileID string, displayName string, stats protocol.PlayerDataStats) (LocalProfileSummary, error) {
	if r == nil || r.store == nil {
		return LocalProfileSummary{}, ErrLocalProfileUnavailable
	}

	localProfileStore, ok := r.store.(LocalProfileStore)
	if !ok {
		return LocalProfileSummary{}, ErrLocalProfileUnavailable
	}

	return localProfileStore.CreateLocalProfile(localProfileID, displayName, stats)
}

func (r *Runtime) DeleteLocalProfile(localProfileID string) error {
	if r == nil || r.store == nil {
		return ErrLocalProfileUnavailable
	}

	localProfileStore, ok := r.store.(LocalProfileStore)
	if !ok {
		return ErrLocalProfileUnavailable
	}

	return localProfileStore.DeleteLocalProfile(localProfileID)
}

func (r *Runtime) UpdateLocalProfileDisplayName(localProfileID string, displayName string) (LocalProfileSummary, error) {
	if r == nil || r.store == nil {
		return LocalProfileSummary{}, ErrLocalProfileUnavailable
	}

	localProfileStore, ok := r.store.(LocalProfileStore)
	if !ok {
		return LocalProfileSummary{}, ErrLocalProfileUnavailable
	}

	return localProfileStore.UpdateLocalProfileDisplayName(localProfileID, displayName)
}

func (r *Runtime) GetDefaultLocalProfile() (LocalProfileDefault, error) {
	if r == nil || r.store == nil {
		return LocalProfileDefault{}, ErrLocalProfileUnavailable
	}

	localProfileStore, ok := r.store.(LocalProfileStore)
	if !ok {
		return LocalProfileDefault{}, ErrLocalProfileUnavailable
	}

	return localProfileStore.GetDefaultLocalProfile()
}

func (r *Runtime) SetDefaultLocalProfile(identityKind string, localProfileID string) (LocalProfileDefault, error) {
	if r == nil || r.store == nil {
		return LocalProfileDefault{}, ErrLocalProfileUnavailable
	}

	localProfileStore, ok := r.store.(LocalProfileStore)
	if !ok {
		return LocalProfileDefault{}, ErrLocalProfileUnavailable
	}

	return localProfileStore.SetDefaultLocalProfile(identityKind, localProfileID)
}
