package playerdata

import "github.com/Lokee86/space-rocks/player-data/protocol"

type Store interface {
	LoadStats(identity protocol.PlayerDataIdentity) (protocol.PlayerDataStats, bool, error)
	RecordMatchResult(command protocol.PlayerDataRecordMatchResult) (protocol.PlayerDataStats, bool, error)
}

type LocalProfileStore interface {
	ListLocalProfiles() ([]LocalProfileSummary, error)
	CreateLocalProfile(localProfileID string, displayName string, stats protocol.PlayerDataStats) (LocalProfileSummary, error)
	GetDefaultLocalProfile() (LocalProfileDefault, error)
	SetDefaultLocalProfile(identityKind string, localProfileID string) (LocalProfileDefault, error)
}

type LocalProfileDefault struct {
	IdentityKind   string
	LocalProfileID string
	DisplayName    string
}
