package game

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/objectives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rules"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

// PlayerMatchFact is the game-owned historical participant fact used by result orchestration.
type PlayerMatchFact struct {
	GamePlayerID string
	TeamID       teams.ID
	Score        int
	ShipDeaths   int
}

type FinalMatchState struct {
	MatchID       string
	TraceID       string
	ModeID        string
	TeamStructure teams.Structure
	Decision      rules.MatchDecision
	Players       []PlayerMatchFact
	Awards        GameplayAwardSnapshot
	Objectives    []objectives.Snapshot
}

type MatchDeathFact = lives.DeathFact

type participantRecord struct {
	ID     string
	TeamID teams.ID
	Score  int
}

func cloneFinalMatchState(source FinalMatchState) FinalMatchState {
	clone := source
	clone.Decision.Players = append([]rules.PlayerDecision(nil), source.Decision.Players...)
	clone.Players = append([]PlayerMatchFact(nil), source.Players...)
	clone.Awards.Counters = append(clone.Awards.Counters[:0:0], source.Awards.Counters...)
	clone.Awards.TeamTotals = append(clone.Awards.TeamTotals[:0:0], source.Awards.TeamTotals...)
	clone.Awards.Combos = append(clone.Awards.Combos[:0:0], source.Awards.Combos...)
	clone.Awards.Streaks = append(clone.Awards.Streaks[:0:0], source.Awards.Streaks...)
	clone.Objectives = make([]objectives.Snapshot, len(source.Objectives))
	for index, snapshot := range source.Objectives {
		clone.Objectives[index] = snapshot
		clone.Objectives[index].Members = append([]string(nil), snapshot.Members...)
		if snapshot.Associations != nil {
			clone.Objectives[index].Associations = make(map[string]string, len(snapshot.Associations))
			for key, value := range snapshot.Associations {
				clone.Objectives[index].Associations[key] = value
			}
		}
	}
	return clone
}
