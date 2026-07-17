package game

import "github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"

// SetTeamStructure applies the room's locked team structure to this match.
// Assignment and balancing remain room-owned; Game only records and reads the result.
func (game *Game) SetTeamStructure(structure teams.Structure) {
	game.mu.Lock()
	defer game.mu.Unlock()
	switch structure {
	case teams.StructureFFA, teams.StructureCoOp, teams.StructureCustom, teams.StructureAutoBalanced:
		game.teamStructure = structure
	default:
		game.teamStructure = teams.StructureFFA
	}
}

func (game *Game) PlayerTeam(playerID string) teams.ID {
	game.mu.Lock()
	defer game.mu.Unlock()
	teamID, _ := game.playerTeamLocked(playerID)
	return teamID
}

func (game *Game) PlayerRelationship(leftPlayerID, rightPlayerID string) teams.Relationship {
	game.mu.Lock()
	defer game.mu.Unlock()

	leftTeam, leftOK := game.playerTeamLocked(leftPlayerID)
	rightTeam, rightOK := game.playerTeamLocked(rightPlayerID)
	if !leftOK || !rightOK {
		return teams.RelationshipUnaffiliated
	}
	relationship, err := teams.RelationshipBetween(game.teamStructure, leftPlayerID, leftTeam, rightPlayerID, rightTeam)
	if err != nil {
		return teams.RelationshipUnaffiliated
	}
	return relationship
}

func (game *Game) playerTeamLocked(playerID string) (teams.ID, bool) {
	if session, ok := game.playerSessions[playerID]; ok && session != nil {
		return session.TeamID, true
	}
	if record, ok := game.participantRecords[playerID]; ok && record != nil {
		return record.TeamID, true
	}
	return teams.NoTeam, false
}
