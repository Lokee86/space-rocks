package teams

import "testing"

func TestRelationshipBetween(t *testing.T) {
	tests := []struct {
		name      string
		structure Structure
		leftID    string
		leftTeam  ID
		rightID   string
		rightTeam ID
		want      Relationship
	}{
		{name: "self ffa", structure: StructureFFA, leftID: "player-1", leftTeam: NoTeam, rightID: "player-1", rightTeam: NoTeam, want: RelationshipSelf},
		{name: "ffa opposing", structure: StructureFFA, leftID: "player-1", rightID: "player-2", want: RelationshipOpposing},
		{name: "co-op same team", structure: StructureCoOp, leftID: "player-1", rightID: "player-2", want: RelationshipSameTeam},
		{name: "custom unaffiliated", structure: StructureCustom, leftID: "player-1", leftTeam: NoTeam, rightID: "player-2", rightTeam: Team1, want: RelationshipUnaffiliated},
		{name: "custom same team", structure: StructureCustom, leftID: "player-1", leftTeam: Team2, rightID: "player-2", rightTeam: Team2, want: RelationshipSameTeam},
		{name: "custom opposing", structure: StructureCustom, leftID: "player-1", leftTeam: Team2, rightID: "player-2", rightTeam: Team3, want: RelationshipOpposing},
		{name: "auto unaffiliated", structure: StructureAutoBalanced, leftID: "player-1", leftTeam: Team1, rightID: "player-2", rightTeam: NoTeam, want: RelationshipUnaffiliated},
		{name: "auto same team", structure: StructureAutoBalanced, leftID: "player-1", leftTeam: Team2, rightID: "player-2", rightTeam: Team2, want: RelationshipSameTeam},
		{name: "auto opposing", structure: StructureAutoBalanced, leftID: "player-1", leftTeam: Team2, rightID: "player-2", rightTeam: Team3, want: RelationshipOpposing},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := RelationshipBetween(test.structure, test.leftID, test.leftTeam, test.rightID, test.rightTeam)
			if err != nil {
				t.Fatalf("RelationshipBetween() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("RelationshipBetween() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestRelationshipBetweenRejectsUnknownStructureAndInvalidTeams(t *testing.T) {
	tests := []struct {
		name      string
		structure Structure
		leftTeam  ID
		rightTeam ID
	}{
		{name: "unknown structure", structure: Structure("unknown")},
		{name: "invalid left team", structure: StructureCustom, leftTeam: ID("team_9"), rightTeam: Team1},
		{name: "invalid right team", structure: StructureCustom, leftTeam: Team1, rightTeam: ID("team_9")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := RelationshipBetween(test.structure, "player-1", test.leftTeam, "player-2", test.rightTeam); err == nil {
				t.Fatal("RelationshipBetween returned nil error")
			}
		})
	}
}
