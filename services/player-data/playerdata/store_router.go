package playerdata

import (
	"errors"
	"fmt"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

type StoreRouter struct {
	accountStore Store
	localStore   Store
	guestStore   Store
}

func NewStoreRouter(accountStore Store, localStore Store, guestStore Store) *StoreRouter {
	return &StoreRouter{
		accountStore: accountStore,
		localStore:   localStore,
		guestStore:   guestStore,
	}
}

func (r *StoreRouter) LoadStats(identity protocol.PlayerDataIdentity) (protocol.PlayerDataStats, bool, error) {
	store, err := r.storeForIdentityKind(identity.IdentityKind)
	if err != nil {
		return protocol.PlayerDataStats{}, false, err
	}
	return store.LoadStats(identity)
}

func (r *StoreRouter) RecordMatchResult(command protocol.PlayerDataRecordMatchResult) (protocol.PlayerDataStats, bool, error) {
	store, err := r.storeForIdentityKind(command.Identity.IdentityKind)
	if err != nil {
		return protocol.PlayerDataStats{}, false, err
	}
	return store.RecordMatchResult(command)
}

func (r *StoreRouter) ListLocalProfiles() ([]LocalProfileSummary, error) {
	localProfileStore, ok := r.localStore.(LocalProfileStore)
	if !ok {
		return nil, errors.New("local profile management is unavailable")
	}

	return localProfileStore.ListLocalProfiles()
}

func (r *StoreRouter) CreateLocalProfile(localProfileID string, displayName string, stats protocol.PlayerDataStats) (LocalProfileSummary, error) {
	localProfileStore, ok := r.localStore.(LocalProfileStore)
	if !ok {
		return LocalProfileSummary{}, errors.New("local profile management is unavailable")
	}

	return localProfileStore.CreateLocalProfile(localProfileID, displayName, stats)
}

func (r *StoreRouter) DeleteLocalProfile(localProfileID string) error {
	localProfileStore, ok := r.localStore.(LocalProfileStore)
	if !ok {
		return errors.New("local profile management is unavailable")
	}

	return localProfileStore.DeleteLocalProfile(localProfileID)
}

func (r *StoreRouter) UpdateLocalProfileDisplayName(localProfileID string, displayName string) (LocalProfileSummary, error) {
	localProfileStore, ok := r.localStore.(LocalProfileStore)
	if !ok {
		return LocalProfileSummary{}, errors.New("local profile management is unavailable")
	}

	return localProfileStore.UpdateLocalProfileDisplayName(localProfileID, displayName)
}

func (r *StoreRouter) GetDefaultLocalProfile() (LocalProfileDefault, error) {
	localProfileStore, ok := r.localStore.(LocalProfileStore)
	if !ok {
		return LocalProfileDefault{}, errors.New("local profile management is unavailable")
	}

	return localProfileStore.GetDefaultLocalProfile()
}

func (r *StoreRouter) SetDefaultLocalProfile(identityKind string, localProfileID string) (LocalProfileDefault, error) {
	localProfileStore, ok := r.localStore.(LocalProfileStore)
	if !ok {
		return LocalProfileDefault{}, errors.New("local profile management is unavailable")
	}

	return localProfileStore.SetDefaultLocalProfile(identityKind, localProfileID)
}

func (r *StoreRouter) storeForIdentityKind(identityKind string) (Store, error) {
	switch identityKind {
	case IdentityKindAuthenticatedAccount:
		return r.accountStore, nil
	case IdentityKindLocalProfile:
		return r.localStore, nil
	case IdentityKindGuest:
		return r.guestStore, nil
	case "":
		return nil, errors.New("missing identity_kind")
	default:
		return nil, fmt.Errorf("unknown identity_kind %q", identityKind)
	}
}
