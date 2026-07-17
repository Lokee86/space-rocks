package teams

// TeamID is the canonical identity for a gameplay team.
type TeamID string

// ID preserves the legacy team identity name.
type ID = TeamID

const (
	NoTeam TeamID = ""
	Team1  TeamID = "team_1"
	Team2  TeamID = "team_2"
	Team3  TeamID = "team_3"
	Team4  TeamID = "team_4"
	Team5  TeamID = "team_5"
	Team6  TeamID = "team_6"
	Team7  TeamID = "team_7"
	Team8  TeamID = "team_8"

	Team1ID = Team1
	Team2ID = Team2
	Team3ID = Team3
	Team4ID = Team4
	Team5ID = Team5
	Team6ID = Team6
	Team7ID = Team7
	Team8ID = Team8
)

func OrderedIDs() []TeamID {
	return []TeamID{Team1, Team2, Team3, Team4, Team5, Team6, Team7, Team8}
}
