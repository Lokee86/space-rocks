package teams

import "testing"

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{name: "ffa", config: Config{Structure: StructureFFA}},
		{name: "co-op", config: Config{Structure: StructureCoOp}},
		{name: "custom player selected", config: Config{Structure: StructureCustom, AssignmentMode: AssignmentPlayerSelected}},
		{name: "custom owner assigned", config: Config{Structure: StructureCustom, AssignmentMode: AssignmentOwnerAssigned}},
		{name: "auto-balanced", config: Config{Structure: StructureAutoBalanced, AutoTeamCount: 2}},
		{name: "auto-balanced eight", config: Config{Structure: StructureAutoBalanced, AutoTeamCount: 8}},
		{name: "unknown structure", config: Config{Structure: Structure("unknown")}, wantErr: true},
		{name: "ffa assignment mode", config: Config{Structure: StructureFFA, AssignmentMode: AssignmentPlayerSelected}, wantErr: true},
		{name: "co-op team count", config: Config{Structure: StructureCoOp, AutoTeamCount: 2}, wantErr: true},
		{name: "custom missing mode", config: Config{Structure: StructureCustom}, wantErr: true},
		{name: "custom unknown mode", config: Config{Structure: StructureCustom, AssignmentMode: AssignmentMode("unknown")}, wantErr: true},
		{name: "custom team count", config: Config{Structure: StructureCustom, AssignmentMode: AssignmentPlayerSelected, AutoTeamCount: 2}, wantErr: true},
		{name: "auto-balanced too few", config: Config{Structure: StructureAutoBalanced, AutoTeamCount: 1}, wantErr: true},
		{name: "auto-balanced too many", config: Config{Structure: StructureAutoBalanced, AutoTeamCount: 9}, wantErr: true},
		{name: "auto-balanced assignment mode", config: Config{Structure: StructureAutoBalanced, AutoTeamCount: 2, AssignmentMode: AssignmentOwnerAssigned}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateConfig(test.config)
			if (err != nil) != test.wantErr {
				t.Fatalf("ValidateConfig(%+v) error = %v, wantErr %v", test.config, err, test.wantErr)
			}
		})
	}
}
