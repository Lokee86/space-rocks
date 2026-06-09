package playerdata

import "github.com/Lokee86/space-rocks/player-data/protocol"

type Store interface {
	LoadStats(identity protocol.PlayerDataIdentity) (protocol.PlayerDataStats, bool, error)
	RecordMatchResult(command protocol.PlayerDataRecordMatchResult) (protocol.PlayerDataStats, bool, error)
}
