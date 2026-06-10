package matchreporting

import (
	"github.com/Lokee86/space-rocks/player-data/playerdata"
	"github.com/Lokee86/space-rocks/player-data/protocol"
	serverplayerdata "github.com/Lokee86/space-rocks/server/internal/playerdata"
)

func BuildRecordMatchResultCommands(summary serverplayerdata.MatchResultSummary) []protocol.PlayerDataRecordMatchResult {
	commands := make([]protocol.PlayerDataRecordMatchResult, 0, len(summary.Players))
	for _, player := range summary.Players {
		command := protocol.PlayerDataRecordMatchResult{
			Type:       protocol.PacketTypePlayerDataRecordMatchResult,
			ResultID:   summary.MatchID + ":" + player.GamePlayerID,
			MatchID:    summary.MatchID,
			Score:      player.Score,
			ShipDeaths: player.ShipDeaths,
			Won:        player.Won,
		}

		switch {
		case player.AccountID != "":
			command.Identity = protocol.PlayerDataIdentity{
				IdentityKind: playerdata.IdentityKindAuthenticatedAccount,
				AccountID:    player.AccountID,
			}
		case player.LocalProfileID != "":
			command.Identity = protocol.PlayerDataIdentity{
				IdentityKind:   playerdata.IdentityKindLocalProfile,
				LocalProfileID: player.LocalProfileID,
			}
		default:
			command.Identity = protocol.PlayerDataIdentity{
				IdentityKind: playerdata.IdentityKindGuest,
			}
		}

		commands = append(commands, command)
	}

	return commands
}
